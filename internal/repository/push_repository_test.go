package repository_test

import (
	"context"
	"testing"

	"github.com/yourorg/shift-app/internal/repository"
	"github.com/yourorg/shift-app/internal/testutil"
)

func TestPushRepository_UpsertAndGet(t *testing.T) {
	db   := testutil.NewDB(t)
	repo := repository.NewPushRepository(db)
	ctx  := context.Background()

	const (
		tenantID int64  = 1
		userID   int64  = 2 // worker
		endpoint        = "https://fcm.googleapis.com/push/abc123"
		p256dh          = "p256dhkey=="
		auth            = "authsecret=="
	)

	// 登録
	if err := repo.Upsert(ctx, tenantID, userID, endpoint, p256dh, auth); err != nil {
		t.Fatalf("Upsert: %v", err)
	}

	// GetByUserID で取得できること
	subs, err := repo.GetByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("GetByUserID: got %d subs, want 1", len(subs))
	}
	if subs[0].Endpoint != endpoint {
		t.Errorf("endpoint: got %q, want %q", subs[0].Endpoint, endpoint)
	}
	if subs[0].P256dh != p256dh {
		t.Errorf("p256dh: got %q, want %q", subs[0].P256dh, p256dh)
	}
}

func TestPushRepository_UpsertUpdatesOnConflict(t *testing.T) {
	db   := testutil.NewDB(t)
	repo := repository.NewPushRepository(db)
	ctx  := context.Background()

	const endpoint = "https://fcm.googleapis.com/push/dup"

	// 初回登録
	if err := repo.Upsert(ctx, 1, 2, endpoint, "key1", "auth1"); err != nil {
		t.Fatalf("initial Upsert: %v", err)
	}

	// 同じ endpoint で上書き
	if err := repo.Upsert(ctx, 1, 2, endpoint, "key2", "auth2"); err != nil {
		t.Fatalf("second Upsert: %v", err)
	}

	// 1件のままで内容が更新されていること
	subs, _ := repo.GetByUserID(ctx, 2)
	if len(subs) != 1 {
		t.Fatalf("dup upsert: got %d subs, want 1", len(subs))
	}
	if subs[0].P256dh != "key2" {
		t.Errorf("p256dh not updated: got %q", subs[0].P256dh)
	}
	if subs[0].Auth != "auth2" {
		t.Errorf("auth not updated: got %q", subs[0].Auth)
	}
}

func TestPushRepository_Delete(t *testing.T) {
	db   := testutil.NewDB(t)
	repo := repository.NewPushRepository(db)
	ctx  := context.Background()

	const endpoint = "https://fcm.googleapis.com/push/todelete"

	_ = repo.Upsert(ctx, 1, 2, endpoint, "k", "a")

	if err := repo.Delete(ctx, 2, endpoint); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	subs, _ := repo.GetByUserID(ctx, 2)
	if len(subs) != 0 {
		t.Fatalf("after Delete: got %d subs, want 0", len(subs))
	}
}

func TestPushRepository_GetByRole(t *testing.T) {
	db   := testutil.NewDB(t)
	repo := repository.NewPushRepository(db)
	ctx  := context.Background()

	// userID=1 (admin), userID=2 (worker), userID=3 (worker) それぞれ登録
	_ = repo.Upsert(ctx, 1, 1, "https://example.com/push/admin",   "k", "a")
	_ = repo.Upsert(ctx, 1, 2, "https://example.com/push/worker1", "k", "a")
	_ = repo.Upsert(ctx, 1, 3, "https://example.com/push/worker2", "k", "a")

	// admin ロールのサブスクリプション取得 → 1件のみ
	adminSubs, err := repo.GetByRole(ctx, 1, "admin")
	if err != nil {
		t.Fatalf("GetByRole admin: %v", err)
	}
	if len(adminSubs) != 1 {
		t.Fatalf("admin subs: got %d, want 1", len(adminSubs))
	}
	if adminSubs[0].Endpoint != "https://example.com/push/admin" {
		t.Errorf("wrong endpoint: %s", adminSubs[0].Endpoint)
	}

	// worker ロールのサブスクリプション取得 → 2件
	workerSubs, err := repo.GetByRole(ctx, 1, "worker")
	if err != nil {
		t.Fatalf("GetByRole worker: %v", err)
	}
	if len(workerSubs) != 2 {
		t.Fatalf("worker subs: got %d, want 2", len(workerSubs))
	}
}

func TestPushRepository_GetByUserID_EmptyWhenNone(t *testing.T) {
	db   := testutil.NewDB(t)
	repo := repository.NewPushRepository(db)
	ctx  := context.Background()

	subs, err := repo.GetByUserID(ctx, 999)
	if err != nil {
		t.Fatalf("GetByUserID: %v", err)
	}
	if len(subs) != 0 {
		t.Fatalf("unexpected subs: %v", subs)
	}
}
