package push_test

import (
	"testing"

	"github.com/yourorg/shift-app/internal/model"
	"github.com/yourorg/shift-app/internal/push"
)

// ─── NewSender ───────────────────────────────────────────────

func TestNewSender_NilWhenKeysEmpty(t *testing.T) {
	if push.NewSender("", "") != nil {
		t.Fatal("両キーが空のとき nil を返すべき")
	}
	if push.NewSender("private", "") != nil {
		t.Fatal("公開鍵が空のとき nil を返すべき")
	}
	if push.NewSender("", "public") != nil {
		t.Fatal("秘密鍵が空のとき nil を返すべき")
	}
}

func TestNewSender_NotNilWhenKeysProvided(t *testing.T) {
	s := push.NewSender("private-key", "public-key")
	if s == nil {
		t.Fatal("両キーが揃っているとき非 nil を返すべき")
	}
}

func TestSender_PublicKey(t *testing.T) {
	const wantKey = "BNxxxxxxPublicKey"
	s := push.NewSender("private", wantKey)
	if got := s.PublicKey(); got != wantKey {
		t.Fatalf("PublicKey: got %q, want %q", got, wantKey)
	}
}

// ─── SendAll (nilセーフ確認) ──────────────────────────────────
// 実際の送信先 URL へのネットワーク通信は行わない。
// webpush.SendNotification が不正な endpoint に対してエラーを返しても
// SendAll はパニックせずログするだけであることを確認する。

func TestSender_SendAll_InvalidEndpointNoParnic(t *testing.T) {
	s := push.NewSender("private", "public")
	subs := []model.PushSubscription{
		{Endpoint: "https://invalid.example/push/xxx", P256dh: "aaa", Auth: "bbb"},
	}
	// パニックしないこと、エラーはログに出るだけ
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("SendAll がパニックした: %v", r)
		}
	}()
	s.SendAll(subs, "タイトル", "本文", "/")
}
