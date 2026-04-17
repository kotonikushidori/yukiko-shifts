package validator

import (
	"context"
	"fmt"
	"time"

	"github.com/yourorg/shift-app/internal/model"
)

// ConflictChecker は二重アサインチェックのインターフェース
type ConflictChecker interface {
	// FindConflicts は指定作業者×日付の既存アサインを全現場横断で返す
	FindAssignmentsByUserDate(ctx context.Context, userID int64, date time.Time) ([]model.ShiftAssignment, error)
}

type ShiftValidator struct {
	repo ConflictChecker
}

func New(repo ConflictChecker) *ShiftValidator {
	return &ShiftValidator{repo: repo}
}

// ConflictError は二重アサインエラーの詳細を持つ
type ConflictError struct {
	UserName    string
	WorkDate    time.Time
	ExistSlot   model.TimeSlot
	ExistSite   string
	RequestSlot model.TimeSlot
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf(
		"%s さんは %s の %s に「%s」へ配置済みのため %s は追加できません",
		e.UserName,
		e.WorkDate.Format("1/2"),
		e.ExistSlot,
		e.ExistSite,
		e.RequestSlot,
	)
}

// ValidateAssign はアサイン追加前の衝突チェックを行う。
// エラーなし → nil, 衝突あり → *ConflictError
func (v *ShiftValidator) ValidateAssign(
	ctx context.Context,
	userID int64,
	userName string,
	date time.Time,
	newSlot model.TimeSlot,
	excludeSiteID int64, // 同一現場内の既存は除外してチェック（編集時用）
) error {
	existing, err := v.repo.FindAssignmentsByUserDate(ctx, userID, date)
	if err != nil {
		return fmt.Errorf("conflict check query: %w", err)
	}

	for _, a := range existing {
		if a.SiteID == excludeSiteID {
			continue // 同一現場内の既存スロットは後続処理で UNIQUE制約が担保
		}
		if a.TimeSlot.Conflicts(newSlot) {
			siteName := ""
			if a.SiteName != nil {
				siteName = *a.SiteName
			}
			return &ConflictError{
				UserName:    userName,
				WorkDate:    date,
				ExistSlot:   a.TimeSlot,
				ExistSite:   siteName,
				RequestSlot: newSlot,
			}
		}
	}
	return nil
}

// ============================================================
// 月報バリデーション
// ============================================================

type ReportValidationError struct {
	Field   string
	Message string
}

func (e *ReportValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// isHalfDay は半日系ステータス（half / half_am / half_pm）かどうかを返す
func isHalfDay(s model.AttendStatus) bool {
	return s == model.StatusHalf || s == model.StatusHalfAM || s == model.StatusHalfPM
}

// ValidateDailyReport は日報入力値の基本チェックを行う。
// 希望入力段階（PUT /api/reports/{date}）では site_id は任意。
// 月報提出時の厳密チェックは ValidateDailyReportForSubmit を使う。
func ValidateDailyReport(req model.DailyReportUpsertRequest) error {
	isWorking := req.Status == model.StatusPresent || isHalfDay(req.Status)
	if isWorking {
		if req.ManDays < 0 || req.ManDays > 2 {
			return &ReportValidationError{
				Field:   "man_days",
				Message: "人工数は0〜2.0の範囲で入力してください",
			}
		}
		if req.OvertimeHours < 0 || req.OvertimeHours > 12 {
			return &ReportValidationError{
				Field:   "overtime_hours",
				Message: "残業時間は0〜12時間で入力してください",
			}
		}
	}
	return nil
}

// ValidateDailyReportForSubmit は月報提出時の厳密バリデーション。
// 出勤日には現場・人工数の入力を必須とする。
func ValidateDailyReportForSubmit(req model.DailyReportUpsertRequest) error {
	if req.Status == model.StatusPresent || isHalfDay(req.Status) {
		if req.SiteID == nil {
			return &ReportValidationError{
				Field:   "site_id",
				Message: "出勤日は現場名を入力してください",
			}
		}
		if req.ManDays <= 0 || req.ManDays > 2 {
			return &ReportValidationError{
				Field:   "man_days",
				Message: "人工数は0.5〜2.0の範囲で入力してください",
			}
		}
	}
	return nil
}
