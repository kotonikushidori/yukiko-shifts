package main

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/shift-app/internal/model"
	"github.com/yourorg/shift-app/internal/repository"
	"github.com/yourorg/shift-app/internal/testutil"
)

// ─── モック reminderSender ────────────────────────────────────

type mockReminderSender struct {
	mu    sync.Mutex
	calls []reminderCall
}

type reminderCall struct {
	subs  []model.PushSubscription
	title string
	body  string
}

func (m *mockReminderSender) SendAll(subs []model.PushSubscription, title, body, url string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = append(m.calls, reminderCall{subs: subs, title: title, body: body})
}

// ─── sendTomorrowReminders ───────────────────────────────────

func TestSendTomorrowReminders_SendsToWorkerWithAssignment(t *testing.T) {
	db       := testutil.NewDB(t)
	pushRepo := repository.NewPushRepository(db)
	ctx      := context.Background()
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	// 作業者 (userID=2) に明日のシフトを作成
	_, err := db.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO shift_assignments (tenant_id, site_id, user_id, work_date, time_slot)
		VALUES (1, 1, 2, '%s', 'AM')`, tomorrow))
	if err != nil {
		t.Fatalf("insert assignment: %v", err)
	}

	// 作業者のサブスクリプションを登録
	_ = pushRepo.Upsert(ctx, 1, 2, "https://fcm.example.com/push/w1", "k", "a")

	ms := &mockReminderSender{}
	sendTomorrowReminders(db, pushRepo, ms)

	if len(ms.calls) != 1 {
		t.Fatalf("SendAll 呼び出し回数: got %d, want 1", len(ms.calls))
	}
	call := ms.calls[0]
	if len(call.subs) != 1 {
		t.Fatalf("送信先サブスク数: got %d, want 1", len(call.subs))
	}
	if call.title != "シフトリマインド" {
		t.Errorf("title: got %q, want %q", call.title, "シフトリマインド")
	}
	// 現場名とスロットが含まれること
	for _, want := range []string{"南平岸４条", "AM"} {
		if !containsStrReminder(call.body, want) {
			t.Errorf("body %q に %q が含まれていない", call.body, want)
		}
	}
}

func TestSendTomorrowReminders_NoAssignment_NoCalls(t *testing.T) {
	db       := testutil.NewDB(t)
	pushRepo := repository.NewPushRepository(db)
	ctx      := context.Background()

	// サブスクリプション登録のみ（シフトなし）
	_ = pushRepo.Upsert(ctx, 1, 2, "https://fcm.example.com/push/w1", "k", "a")

	ms := &mockReminderSender{}
	sendTomorrowReminders(db, pushRepo, ms)

	if len(ms.calls) != 0 {
		t.Fatalf("シフトなし: SendAll が呼ばれた (%d calls)", len(ms.calls))
	}
}

func TestSendTomorrowReminders_NoSubscription_NoCalls(t *testing.T) {
	db       := testutil.NewDB(t)
	pushRepo := repository.NewPushRepository(db)
	ctx      := context.Background()
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	// シフトはあるがサブスクリプションなし
	_, _ = db.ExecContext(ctx, fmt.Sprintf(`
		INSERT INTO shift_assignments (tenant_id, site_id, user_id, work_date, time_slot)
		VALUES (1, 1, 2, '%s', 'ALL')`, tomorrow))

	ms := &mockReminderSender{}
	sendTomorrowReminders(db, pushRepo, ms)

	if len(ms.calls) != 0 {
		t.Fatalf("サブスクなし: SendAll が呼ばれた (%d calls)", len(ms.calls))
	}
}

func TestSendTomorrowReminders_NilSender_NoOp(t *testing.T) {
	db       := testutil.NewDB(t)
	pushRepo := repository.NewPushRepository(db)
	// パニックしないこと
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nil sender でパニック: %v", r)
		}
	}()
	sendTomorrowReminders(db, pushRepo, nil)
}

func TestSendTomorrowReminders_MultipleAssignments_GroupedByUser(t *testing.T) {
	db       := testutil.NewDB(t)
	pushRepo := repository.NewPushRepository(db)
	ctx      := context.Background()
	tomorrow := time.Now().Add(24 * time.Hour).Format("2006-01-02")

	// 作業者 (userID=2) に AM/PM 2件のシフト → 1回の SendAll にまとまること
	for _, slot := range []string{"AM", "PM"} {
		_, _ = db.ExecContext(ctx, fmt.Sprintf(`
			INSERT INTO shift_assignments (tenant_id, site_id, user_id, work_date, time_slot)
			VALUES (1, 1, 2, '%s', '%s')`, tomorrow, slot))
	}
	_ = pushRepo.Upsert(ctx, 1, 2, "https://fcm.example.com/push/w1", "k", "a")

	ms := &mockReminderSender{}
	sendTomorrowReminders(db, pushRepo, ms)

	// ユーザーごとに 1回の SendAll
	if len(ms.calls) != 1 {
		t.Fatalf("SendAll 呼び出し: got %d, want 1 (1ユーザー=1回)", len(ms.calls))
	}
	// body に AM と PM 両方含まれること
	body := ms.calls[0].body
	if !containsStrReminder(body, "AM") || !containsStrReminder(body, "PM") {
		t.Errorf("AM/PM が body に含まれていない: %q", body)
	}
}

func containsStrReminder(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
