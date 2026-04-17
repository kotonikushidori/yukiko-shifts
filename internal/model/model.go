package model

import "time"

// ============================================================
// ドメインモデル定義
// ============================================================

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleWorker Role = "worker"
)

type TimeSlot string

const (
	SlotAM  TimeSlot = "AM"
	SlotPM  TimeSlot = "PM"
	SlotAll TimeSlot = "ALL"
)

// Conflicts returns true if two time slots overlap.
// ALL conflicts with everything; AM/PM conflict with same or ALL.
func (t TimeSlot) Conflicts(other TimeSlot) bool {
	if t == SlotAll || other == SlotAll {
		return true
	}
	return t == other
}

type AttendStatus string

const (
	StatusPresent AttendStatus = "present"  // ○
	StatusAbsent  AttendStatus = "absent"   // ×
	StatusHalf    AttendStatus = "half"     // △（未分類）
	StatusHalfAM  AttendStatus = "half_am"  // △前（午前のみ可）
	StatusHalfPM  AttendStatus = "half_pm"  // △後（午後のみ可）
)

// ──────────────────────────────────────
// User
// ──────────────────────────────────────
type User struct {
	ID           int64     `db:"id"           json:"id"`
	TenantID     int64     `db:"tenant_id"    json:"tenant_id"`
	EmployeeID   string    `db:"employee_id"  json:"employee_id"`
	Email        *string   `db:"email"        json:"email,omitempty"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Name         string    `db:"name"         json:"name"`
	Role         Role      `db:"role"         json:"role"`
	Phone        *string   `db:"phone"        json:"phone,omitempty"`
	Status       string    `db:"status"       json:"status"`
	CreatedAt    time.Time `db:"created_at"   json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"   json:"updated_at"`
}

// ──────────────────────────────────────
// Site（現場）
// ──────────────────────────────────────
type Site struct {
	ID        int64      `db:"id"         json:"id"`
	TenantID  int64      `db:"tenant_id"  json:"tenant_id,omitempty"`
	Name      string     `db:"name"       json:"name"`
	Client    *string    `db:"client"     json:"client,omitempty"`
	Address   *string    `db:"address"    json:"address,omitempty"`
	BudgetYen *int64     `db:"budget_yen" json:"budget_yen,omitempty"`
	StartDate *time.Time `db:"start_date" json:"start_date,omitempty"`
	EndDate   *time.Time `db:"end_date"   json:"end_date,omitempty"`
	Note      *string    `db:"note"       json:"note,omitempty"`
	Status    string     `db:"status"     json:"status"`
	CreatedBy *int64     `db:"created_by" json:"created_by,omitempty"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

// SiteUpsertRequest は現場の登録・更新リクエスト
type SiteUpsertRequest struct {
	Name      string  `json:"name"`
	Client    *string `json:"client"`
	Address   *string `json:"address"`
	BudgetYen *int64  `json:"budget_yen"`
	StartDate *string `json:"start_date"` // YYYY-MM-DD or null
	EndDate   *string `json:"end_date"`   // YYYY-MM-DD or null
	Note      *string `json:"note"`
	Status    string  `json:"status"` // "active" | "completed"
}

// ──────────────────────────────────────
// ShiftAssignment（シフト配置）
// ──────────────────────────────────────
type ShiftAssignment struct {
	ID        int64     `db:"id"         json:"id"`
	TenantID  int64     `db:"tenant_id"  json:"tenant_id,omitempty"`
	SiteID    int64     `db:"site_id"    json:"site_id"`
	UserID    int64     `db:"user_id"    json:"user_id"`
	WorkDate  time.Time `db:"work_date"  json:"work_date"`
	TimeSlot  TimeSlot  `db:"time_slot"  json:"time_slot"`
	CreatedBy *int64    `db:"created_by" json:"created_by,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	// JOINで取得する関連情報
	UserName *string `db:"user_name"  json:"user_name,omitempty"`
	SiteName *string `db:"site_name"  json:"site_name,omitempty"`
}

// ──────────────────────────────────────
// DailyReport（日報）
// ──────────────────────────────────────
type DailyReport struct {
	ID             int64        `db:"id"              json:"id"`
	TenantID       int64        `db:"tenant_id"       json:"tenant_id,omitempty"`
	UserID         int64        `db:"user_id"         json:"user_id"`
	WorkDate       time.Time    `db:"work_date"       json:"work_date"`
	Status         AttendStatus `db:"status"          json:"status"`
	SiteID         *int64       `db:"site_id"         json:"site_id,omitempty"`
	SiteID2        *int64       `db:"site_id2"        json:"site_id2,omitempty"`
	ClientName     *string      `db:"client_name"     json:"client_name,omitempty"`
	ManDays        float64      `db:"man_days"        json:"man_days"`
	OvertimeHours  float64      `db:"overtime_hours"  json:"overtime_hours"`
	UsedCar        bool         `db:"used_car"        json:"used_car"`
	Note           *string      `db:"note"            json:"note,omitempty"`
	SubmittedAt    *time.Time   `db:"submitted_at"    json:"submitted_at,omitempty"`
	UpdatedAt      time.Time    `db:"updated_at"      json:"updated_at"`

	// JOINで取得
	UserName  *string `db:"user_name"  json:"user_name,omitempty"`
	SiteName  *string `db:"site_name"  json:"site_name,omitempty"`
	SiteName2 *string `db:"site_name2" json:"site_name2,omitempty"`
}

// ──────────────────────────────────────
// ShiftLock（希望入力ロック）
// ──────────────────────────────────────
type ShiftLock struct {
	ID       int64     `db:"id"        json:"id"`
	TenantID int64     `db:"tenant_id" json:"-"`
	Year     int       `db:"year"      json:"year"`
	Month    int       `db:"month"     json:"month"`
	LockedAt time.Time `db:"locked_at" json:"locked_at"`
	LockedBy int64     `db:"locked_by" json:"locked_by"`
}

// ──────────────────────────────────────
// Announcement（全体連絡）
// ──────────────────────────────────────
type Announcement struct {
	ID        int64     `db:"id"         json:"id"`
	Title     string    `db:"title"      json:"title"`
	Body      string    `db:"body"       json:"body"`
	CreatedBy int64     `db:"created_by" json:"created_by"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`

	CreatedByName *string `db:"created_by_name" json:"created_by_name,omitempty"`
	ReadAt        *time.Time `db:"read_at"      json:"read_at,omitempty"` // 自分の既読日時
}

// ──────────────────────────────────────
// API Request/Response 型
// ──────────────────────────────────────

type LoginRequest struct {
	EmployeeID string `json:"employee_id"`
	Password   string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type AssignRequest struct {
	UserID   int64    `json:"user_id"`
	TimeSlot TimeSlot `json:"time_slot"`
}

type DailyReportUpsertRequest struct {
	Status        AttendStatus `json:"status"`
	SiteID        *int64       `json:"site_id,omitempty"`
	SiteID2       *int64       `json:"site_id2,omitempty"`
	ClientName    *string      `json:"client_name,omitempty"`
	ManDays       float64      `json:"man_days"`
	OvertimeHours float64      `json:"overtime_hours"`
	UsedCar       bool         `json:"used_car"`
	Note          *string      `json:"note,omitempty"`
}

// MonthlySummaryRow は月次サマリの1行
type MonthlySummaryRow struct {
	UserID        int64   `json:"user_id"`
	UserName      string  `json:"user_name"`
	TotalPresent  int     `json:"total_present"`
	TotalAbsent   int     `json:"total_absent"`
	TotalManDays  float64 `json:"total_man_days"`
	TotalOvertime float64 `json:"total_overtime"`
	MissingDays   int     `json:"missing_days"` // 出勤予定だが未入力の日数
}

// ──────────────────────────────────────
// Tenant（テナント）
// ──────────────────────────────────────
type Plan string

const (
	PlanBasic Plan = "basic"
	PlanPro   Plan = "pro"
)

type Tenant struct {
	ID            int64      `db:"id"             json:"id"`
	Name          string     `db:"name"           json:"name"`
	Slug          string     `db:"slug"           json:"slug"`
	Plan          Plan       `db:"plan"           json:"plan"`
	MaxWorkers    int        `db:"max_workers"    json:"max_workers"`
	Status        string     `db:"status"         json:"status"`
	ContractStart *time.Time `db:"contract_start" json:"contract_start,omitempty"`
	ContractEnd   *time.Time `db:"contract_end"   json:"contract_end,omitempty"`
	CreatedAt     time.Time  `db:"created_at"     json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"     json:"updated_at"`
}
