package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"

	"github.com/yourorg/shift-app/internal/handler"
	"github.com/yourorg/shift-app/internal/repository"
	"github.com/yourorg/shift-app/internal/validator"
)

// contextキー型（文字列キーの衝突防止）
type contextKey string

const (
	ctxTenantID contextKey = "tenant_id"
	ctxUserID   contextKey = "user_id"
	ctxRole     contextKey = "role"
)

func main() {
	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production-32chars!!")
	dbPath    := getEnv("DB_PATH",    "./shift.db")
	port      := getEnv("PORT",       "8989")

	db, err := sqlx.Open("sqlite3", dbPath+"?_foreign_keys=on&_journal_mode=WAL")
	if err != nil {
		log.Fatalf("DB open: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(1)

	if err := runMigrations(db.DB); err != nil {
		log.Fatalf("migration: %v", err)
	}

	tokenAuth  := jwtauth.New("HS256", []byte(jwtSecret), nil)
	userRepo   := repository.NewUserRepository(db)
	shiftRepo  := repository.NewShiftRepository(db)
	reportRepo := repository.NewDailyReportRepository(db)
	siteRepo   := repository.NewSiteRepository(db)
	lockRepo   := repository.NewLockRepository(db)
	shiftVal   := validator.New(shiftRepo)

	authH   := handler.NewAuthHandler(userRepo, tokenAuth)
	shiftH  := handler.NewShiftHandler(shiftRepo, userRepo, shiftVal)
	reportH := handler.NewDailyReportHandler(reportRepo)
	siteH   := handler.NewSiteHandler(siteRepo)
	lockH   := handler.NewLockHandler(lockRepo)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(corsMiddleware)

	// 静的ファイル
	r.Handle("/static/*", http.StripPrefix("/static/",
		http.FileServer(http.Dir("./frontend/static"))))

	// 認証不要
	r.Post("/api/auth/login", authH.Login)

	// 認証必要 + テナント自動注入
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Use(tenantMiddleware) // ★ JWTからtenant_idをContextに注入

		r.Get("/api/shifts/board",  shiftH.GetBoard)
		r.Get("/api/shifts/my",     shiftH.GetMyShifts)

		r.Post("/api/sites/{siteID}/shifts/{date}/assign",
			handler.RequireAdmin(shiftH.CreateAssign))
		r.Delete("/api/shifts/assign/{id}",
			handler.RequireAdmin(shiftH.DeleteAssign))

		// 現場マスタ
		r.Get("/api/sites",        siteH.List)
		r.Get("/api/sites/{id}",   siteH.Get)
		r.Post("/api/sites",       handler.RequireAdmin(siteH.Create))
		r.Put("/api/sites/{id}",   handler.RequireAdmin(siteH.Update))

		r.Put("/api/reports/{date}",        reportH.Upsert)
		r.Get("/api/reports/my",            reportH.GetMyMonthly)
		r.Put("/api/reports/site-client",   reportH.UpdateSiteClient)
		r.Post("/api/reports/submit",       reportH.Submit)

		r.Get("/api/reports/summary",
			handler.RequireAdmin(reportH.GetSummary))

		// シフトロック
		r.Get("/api/shifts/lock",                  lockH.GetStatus)
		r.Post("/api/admin/shifts/lock",            handler.RequireAdmin(lockH.Lock))
		r.Delete("/api/admin/shifts/lock",          handler.RequireAdmin(lockH.Unlock))
	})

	// SPAフォールバック
	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./frontend/templates/index.html")
	})

	log.Printf("起動: http://localhost:%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// ── テナントミドルウェア ─────────────────────────────────────────
// JWTのclaimsからtenant_idを取り出してContextに注入する。
// 全APIハンドラーはこのContextからtenant_idを取得するだけでOK。
func tenantMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := jwtauth.FromContext(r.Context())
		if err != nil || claims == nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		tenantID := int64(claims["tenant_id"].(float64))
		userID   := int64(claims["user_id"].(float64))
		role, _  := claims["role"].(string)

		ctx := context.WithValue(r.Context(), ctxTenantID, tenantID)
		ctx  = context.WithValue(ctx,          ctxUserID,   userID)
		ctx  = context.WithValue(ctx,          ctxRole,     role)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ── マイグレーション ─────────────────────────────────────────────
// db/migrations/ 以下の *.sql をファイル名昇順で実行する。
// ALTER TABLE など冪等でないステートメントは "duplicate column" エラーを無視する。
func runMigrations(db *sql.DB) error {
	entries, err := os.ReadDir("./db/migrations")
	if err != nil {
		return fmt.Errorf("ReadDir migrations: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		data, err := os.ReadFile("./db/migrations/" + name)
		if err != nil {
			return fmt.Errorf("ReadFile %s: %w", name, err)
		}
		if _, err := db.Exec(string(data)); err != nil {
			// SQLite の "duplicate column name" は冪等とみなして無視
			if !strings.Contains(err.Error(), "duplicate column name") {
				return fmt.Errorf("migration %s: %w", name, err)
			}
			log.Printf("migration %s: skipped (already applied)", name)
		} else {
			log.Printf("migration %s: applied", name)
		}
	}
	return nil
}

// ── CORS ─────────────────────────────────────────────────────────
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
