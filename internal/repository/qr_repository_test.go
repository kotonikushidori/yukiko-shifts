package repository_test

import (
	"context"
	"testing"

	"github.com/yourorg/shift-app/internal/model"
	"github.com/yourorg/shift-app/internal/repository"
	"github.com/yourorg/shift-app/internal/testutil"
)

// ─── FindByQRToken ───────────────────────────────────────────

func TestUserRepository_FindByQRToken_Found(t *testing.T) {
	db   := testutil.NewDB(t)
	repo := repository.NewUserRepository(db)
	ctx  := context.Background()

	const token = "test-qr-token-abc123"
	if err := repo.UpdateQRToken(ctx, 2, token); err != nil {
		t.Fatalf("UpdateQRToken: %v", err)
	}

	u, err := repo.FindByQRToken(ctx, token)
	if err != nil {
		t.Fatalf("FindByQRToken: %v", err)
	}
	if u == nil {
		t.Fatal("got nil, want user")
	}
	if u.ID != 2 {
		t.Errorf("user ID: got %d, want 2", u.ID)
	}
	if u.EmployeeID != "w001" {
		t.Errorf("employee_id: got %q, want w001", u.EmployeeID)
	}
}

func TestUserRepository_FindByQRToken_NotFound(t *testing.T) {
	db   := testutil.NewDB(t)
	repo := repository.NewUserRepository(db)
	ctx  := context.Background()

	u, err := repo.FindByQRToken(ctx, "does-not-exist")
	if err != nil {
		t.Fatalf("FindByQRToken: %v", err)
	}
	if u != nil {
		t.Errorf("expected nil, got user id=%d", u.ID)
	}
}

// ─── UpdateQRToken ───────────────────────────────────────────

func TestUserRepository_UpdateQRToken_Reissue(t *testing.T) {
	db   := testutil.NewDB(t)
	repo := repository.NewUserRepository(db)
	ctx  := context.Background()

	const (
		oldToken = "old-token-111"
		newToken = "new-token-222"
	)

	// 初回セット
	if err := repo.UpdateQRToken(ctx, 2, oldToken); err != nil {
		t.Fatalf("first UpdateQRToken: %v", err)
	}
	// 再発行（上書き）
	if err := repo.UpdateQRToken(ctx, 2, newToken); err != nil {
		t.Fatalf("second UpdateQRToken: %v", err)
	}

	// 古いトークンでは見つからない
	old, _ := repo.FindByQRToken(ctx, oldToken)
	if old != nil {
		t.Error("古いトークンで検索できてしまう")
	}

	// 新しいトークンで見つかる
	u, err := repo.FindByQRToken(ctx, newToken)
	if err != nil {
		t.Fatalf("FindByQRToken new: %v", err)
	}
	if u == nil || u.ID != 2 {
		t.Error("新しいトークンで検索できない")
	}
}

// ─── GetAllQRTokens ──────────────────────────────────────────

func TestUserRepository_GetAllQRTokens(t *testing.T) {
	db   := testutil.NewDB(t)
	repo := repository.NewUserRepository(db)
	ctx  := context.Background()

	// 2人の作業者にトークンをセット（admin userID=1 は含まれない）
	if err := repo.UpdateQRToken(ctx, 2, "token-w001"); err != nil {
		t.Fatalf("UpdateQRToken w001: %v", err)
	}
	if err := repo.UpdateQRToken(ctx, 3, "token-w002"); err != nil {
		t.Fatalf("UpdateQRToken w002: %v", err)
	}

	rows, err := repo.GetAllQRTokens(ctx, 1)
	if err != nil {
		t.Fatalf("GetAllQRTokens: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}

	// name 昇順: 佐藤花子(w002) → 田中一郎(w001)
	if rows[0].EmployeeID != "w002" {
		t.Errorf("rows[0].EmployeeID: got %q, want w002", rows[0].EmployeeID)
	}
	if rows[0].QRToken != "token-w002" {
		t.Errorf("rows[0].QRToken: got %q, want token-w002", rows[0].QRToken)
	}
	if rows[1].EmployeeID != "w001" {
		t.Errorf("rows[1].EmployeeID: got %q, want w001", rows[1].EmployeeID)
	}
	if rows[1].QRToken != "token-w001" {
		t.Errorf("rows[1].QRToken: got %q, want token-w001", rows[1].QRToken)
	}
}

func TestUserRepository_GetAllQRTokens_ExcludesNullTokens(t *testing.T) {
	db   := testutil.NewDB(t)
	repo := repository.NewUserRepository(db)
	ctx  := context.Background()

	// トークンをセットしない → qr_token IS NULL → 結果に含まれない
	rows, err := repo.GetAllQRTokens(ctx, 1)
	if err != nil {
		t.Fatalf("GetAllQRTokens: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("got %d rows, want 0 (tokens are NULL)", len(rows))
	}
}

// ─── Create (QRToken 自動生成) ────────────────────────────────

func TestUserRepository_Create_AutoGeneratesQRToken(t *testing.T) {
	db   := testutil.NewDB(t)
	repo := repository.NewUserRepository(db)
	ctx  := context.Background()

	ln := "新規"
	fn := "作業者"
	id, err := repo.Create(ctx, model.User{
		TenantID:     1,
		EmployeeID:   "w999",
		Name:         ln + fn,
		LastName:     &ln,
		FirstName:    &fn,
		PasswordHash: "dummy-hash",
		Role:         model.RoleWorker,
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	u, err := repo.FindByID(ctx, 1, id)
	if err != nil {
		t.Fatalf("FindByID: %v", err)
	}
	if u == nil {
		t.Fatal("FindByID: got nil")
	}
	if u.QRToken == nil || *u.QRToken == "" {
		t.Error("Create: QRToken が自動生成されていない")
	}
	if len(*u.QRToken) != 32 { // hex(16 bytes) = 32文字
		t.Errorf("QRToken length: got %d, want 32", len(*u.QRToken))
	}
}

func TestUserRepository_Create_AdminHasNoQRToken(t *testing.T) {
	db   := testutil.NewDB(t)
	repo := repository.NewUserRepository(db)
	ctx  := context.Background()

	ln := "管理"
	fn := "者２"
	id, err := repo.Create(ctx, model.User{
		TenantID:     1,
		EmployeeID:   "admin2",
		Name:         ln + fn,
		LastName:     &ln,
		FirstName:    &fn,
		PasswordHash: "dummy-hash",
		Role:         model.RoleAdmin, // admin は QRToken 不要
		Status:       "active",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	u, err := repo.FindByID(ctx, 1, id)
	if err != nil || u == nil {
		t.Fatalf("FindByID: %v", err)
	}
	if u.QRToken != nil {
		t.Errorf("admin の QRToken はnilであるべき、got %q", *u.QRToken)
	}
}
