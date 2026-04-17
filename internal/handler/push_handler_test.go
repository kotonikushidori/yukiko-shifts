package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"

	"github.com/yourorg/shift-app/internal/handler"
	"github.com/yourorg/shift-app/internal/model"
	"github.com/yourorg/shift-app/internal/repository"
	"github.com/yourorg/shift-app/internal/testutil"
)

// ─── モック PushSender ────────────────────────────────────────

type mockSender struct {
	mu        sync.Mutex
	publicKey string
	calls     []mockCall
}

type mockCall struct {
	subs  []model.PushSubscription
	title string
	body  string
}

func (m *mockSender) PublicKey() string { return m.publicKey }
func (m *mockSender) SendAll(subs []model.PushSubscription, title, body, url string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, mockCall{subs: subs, title: title, body: body})
}
func (m *mockSender) lastCall() *mockCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.calls) == 0 {
		return nil
	}
	c := m.calls[len(m.calls)-1]
	return &c
}

// ─── テスト用ルーター ─────────────────────────────────────────

const testJWTSecret = "test-secret-for-push-handler-tests"

// buildRouter は認証ミドルウェア込みのルーターを構築する
func buildRouter(h *handler.PushHandler) (http.Handler, *jwtauth.JWTAuth) {
	tokenAuth := jwtauth.New("HS256", []byte(testJWTSecret), nil)
	r := chi.NewRouter()

	// 認証不要: VAPID 公開鍵
	r.Get("/api/push/vapid-key", h.GetVapidKey)

	// 認証必要グループ
	r.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(tokenAuth))
		r.Use(jwtauth.Authenticator(tokenAuth))
		r.Use(injectClaimsMiddleware)

		r.Post("/api/push/subscribe",   h.Subscribe)
		r.Delete("/api/push/subscribe", h.Unsubscribe)
		r.Post("/api/push/hope-submit", h.HopeSubmit)
	})
	return r, tokenAuth
}

// injectClaimsMiddleware は JWT クレームを handler が参照できるよう context にセットする
func injectClaimsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// bearerToken は指定クレームで JWT トークン文字列を生成する
func bearerToken(t *testing.T, ta *jwtauth.JWTAuth, claims map[string]any) string {
	t.Helper()
	_, tokenStr, err := ta.Encode(claims)
	if err != nil {
		t.Fatalf("token encode: %v", err)
	}
	return "Bearer " + tokenStr
}

// adminClaims は tenantID=1, userID=1, role=admin のクレームを返す
func adminClaims() map[string]any {
	return map[string]any{"tenant_id": float64(1), "user_id": float64(1), "role": "admin"}
}

// workerClaims は tenantID=1, userID=2, role=worker のクレームを返す
func workerClaims() map[string]any {
	return map[string]any{"tenant_id": float64(1), "user_id": float64(2), "role": "worker"}
}

// ─── GetVapidKey ─────────────────────────────────────────────

func TestGetVapidKey_NilSender(t *testing.T) {
	db       := testutil.NewDB(t)
	pushRepo := repository.NewPushRepository(db)
	userRepo := repository.NewUserRepository(db)
	h        := handler.NewPushHandler(pushRepo, userRepo, nil)
	router, _ := buildRouter(h)

	req := httptest.NewRequest("GET", "/api/push/vapid-key", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["public_key"] != "" {
		t.Errorf("nil sender: public_key should be empty, got %q", resp["public_key"])
	}
}

func TestGetVapidKey_WithSender(t *testing.T) {
	db       := testutil.NewDB(t)
	pushRepo := repository.NewPushRepository(db)
	userRepo := repository.NewUserRepository(db)
	ms       := &mockSender{publicKey: "BNxxxxPublicKey"}
	h        := handler.NewPushHandler(pushRepo, userRepo, ms)
	router, _ := buildRouter(h)

	req := httptest.NewRequest("GET", "/api/push/vapid-key", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status: got %d, want 200", rec.Code)
	}
	var resp map[string]string
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["public_key"] != "BNxxxxPublicKey" {
		t.Errorf("public_key: got %q", resp["public_key"])
	}
}

// ─── Subscribe ───────────────────────────────────────────────

func TestSubscribe_SavesToDB(t *testing.T) {
	db       := testutil.NewDB(t)
	pushRepo := repository.NewPushRepository(db)
	userRepo := repository.NewUserRepository(db)
	h        := handler.NewPushHandler(pushRepo, userRepo, &mockSender{})
	router, ta := buildRouter(h)

	body, _ := json.Marshal(map[string]string{
		"endpoint": "https://fcm.example.com/push/abc",
		"p256dh":   "p256dhKey==",
		"auth":     "authKey==",
	})
	req := httptest.NewRequest("POST", "/api/push/subscribe", bytes.NewReader(body))
	req.Header.Set("Authorization", bearerToken(t, ta, workerClaims()))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status: got %d, want 204", rec.Code)
	}

	// DB に保存されていること
	subs, err := pushRepo.GetByUserID(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("subs count: got %d, want 1", len(subs))
	}
	if subs[0].Endpoint != "https://fcm.example.com/push/abc" {
		t.Errorf("endpoint: %s", subs[0].Endpoint)
	}
}

func TestSubscribe_BadRequest(t *testing.T) {
	db       := testutil.NewDB(t)
	pushRepo := repository.NewPushRepository(db)
	userRepo := repository.NewUserRepository(db)
	h        := handler.NewPushHandler(pushRepo, userRepo, &mockSender{})
	router, ta := buildRouter(h)

	// endpoint なし
	body, _ := json.Marshal(map[string]string{"p256dh": "k"})
	req := httptest.NewRequest("POST", "/api/push/subscribe", bytes.NewReader(body))
	req.Header.Set("Authorization", bearerToken(t, ta, workerClaims()))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status: got %d, want 400", rec.Code)
	}
}

// ─── Unsubscribe ─────────────────────────────────────────────

func TestUnsubscribe_DeletesFromDB(t *testing.T) {
	db       := testutil.NewDB(t)
	pushRepo := repository.NewPushRepository(db)
	userRepo := repository.NewUserRepository(db)
	h        := handler.NewPushHandler(pushRepo, userRepo, &mockSender{})
	router, ta := buildRouter(h)

	const ep = "https://fcm.example.com/push/toremove"
	ctx := context.Background()
	_ = pushRepo.Upsert(ctx, 1, 2, ep, "k", "a")

	body, _ := json.Marshal(map[string]string{"endpoint": ep})
	req := httptest.NewRequest("DELETE", "/api/push/subscribe", bytes.NewReader(body))
	req.Header.Set("Authorization", bearerToken(t, ta, workerClaims()))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status: got %d, want 204", rec.Code)
	}

	subs, _ := pushRepo.GetByUserID(ctx, 2)
	if len(subs) != 0 {
		t.Fatalf("after unsubscribe: got %d subs, want 0", len(subs))
	}
}

// ─── HopeSubmit ──────────────────────────────────────────────

func TestHopeSubmit_NilSender_NoOp(t *testing.T) {
	db       := testutil.NewDB(t)
	pushRepo := repository.NewPushRepository(db)
	userRepo := repository.NewUserRepository(db)
	h        := handler.NewPushHandler(pushRepo, userRepo, nil) // sender = nil
	router, ta := buildRouter(h)

	body, _ := json.Marshal(map[string]int{"year": 2026, "month": 5})
	req := httptest.NewRequest("POST", "/api/push/hope-submit", bytes.NewReader(body))
	req.Header.Set("Authorization", bearerToken(t, ta, workerClaims()))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status: got %d, want 204", rec.Code)
	}
}

func TestHopeSubmit_SendsToAdmins(t *testing.T) {
	db       := testutil.NewDB(t)
	pushRepo := repository.NewPushRepository(db)
	userRepo := repository.NewUserRepository(db)
	ms       := &mockSender{publicKey: "pub"}
	h        := handler.NewPushHandler(pushRepo, userRepo, ms)
	router, ta := buildRouter(h)

	ctx := context.Background()

	// 管理者のサブスクリプションを登録
	_ = pushRepo.Upsert(ctx, 1, 1, "https://fcm.example.com/push/admin", "k", "a")
	// 別の作業者のサブスク（管理者には届かないことを確認）
	_ = pushRepo.Upsert(ctx, 1, 3, "https://fcm.example.com/push/worker3", "k", "a")

	body, _ := json.Marshal(map[string]int{"year": 2026, "month": 5})
	// 作業者 (userID=2, 田中一郎) が提出
	req := httptest.NewRequest("POST", "/api/push/hope-submit", bytes.NewReader(body))
	req.Header.Set("Authorization", bearerToken(t, ta, workerClaims()))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status: got %d, want 204 (body: %s)", rec.Code, rec.Body.String())
	}

	call := ms.lastCall()
	if call == nil {
		t.Fatal("SendAll が呼ばれなかった")
	}

	// 管理者のサブスク 1件のみ届く
	if len(call.subs) != 1 {
		t.Fatalf("送信先: got %d subs, want 1", len(call.subs))
	}
	if call.subs[0].Endpoint != "https://fcm.example.com/push/admin" {
		t.Errorf("wrong endpoint: %s", call.subs[0].Endpoint)
	}

	// タイトル・本文の確認
	if call.title != "希望提出完了" {
		t.Errorf("title: got %q, want %q", call.title, "希望提出完了")
	}
	// 作業者名と年月が含まれること
	for _, want := range []string{"田中", "2026", "5"} {
		if !containsStr(call.body, want) {
			t.Errorf("body %q に %q が含まれていない", call.body, want)
		}
	}
}

func TestHopeSubmit_NoAdminSubs_NoSendAllCall(t *testing.T) {
	db       := testutil.NewDB(t)
	pushRepo := repository.NewPushRepository(db)
	userRepo := repository.NewUserRepository(db)
	ms       := &mockSender{publicKey: "pub"}
	h        := handler.NewPushHandler(pushRepo, userRepo, ms)
	router, ta := buildRouter(h)

	// 管理者のサブスクリプションを登録しない

	body, _ := json.Marshal(map[string]int{"year": 2026, "month": 5})
	req := httptest.NewRequest("POST", "/api/push/hope-submit", bytes.NewReader(body))
	req.Header.Set("Authorization", bearerToken(t, ta, workerClaims()))
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status: got %d, want 204", rec.Code)
	}

	// サブスクリプションが 0件なので SendAll は呼ばれるが subs が空のはず
	call := ms.lastCall()
	if call != nil && len(call.subs) > 0 {
		t.Fatalf("宛先ゼロなのに送信された: %v", call.subs)
	}
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
