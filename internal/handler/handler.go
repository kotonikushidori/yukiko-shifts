package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/yourorg/shift-app/internal/model"
	"github.com/yourorg/shift-app/internal/repository"
	"github.com/yourorg/shift-app/internal/validator"
)

// ============================================================
// ヘルパー
// ============================================================

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func currentUserID(r *http.Request) int64 {
	_, claims, _ := jwtauth.FromContext(r.Context())
	id, _ := claims["user_id"].(float64)
	return int64(id)
}

func currentUserRole(r *http.Request) string {
	_, claims, _ := jwtauth.FromContext(r.Context())
	role, _ := claims["role"].(string)
	return role
}

func currentTenantID(r *http.Request) int64 {
	_, claims, _ := jwtauth.FromContext(r.Context())
	id, _ := claims["tenant_id"].(float64)
	return int64(id)
}

func RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if currentUserRole(r) != string(model.RoleAdmin) {
			writeError(w, http.StatusForbidden, "管理者権限が必要です")
			return
		}
		next(w, r)
	}
}

// ============================================================
// AuthHandler
// ============================================================

type AuthHandler struct {
	userRepo  *repository.UserRepository
	tokenAuth *jwtauth.JWTAuth
}

func NewAuthHandler(repo *repository.UserRepository, tokenAuth *jwtauth.JWTAuth) *AuthHandler {
	return &AuthHandler{userRepo: repo, tokenAuth: tokenAuth}
}

// POST /api/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "リクエスト形式が不正です")
		return
	}

	user, err := h.userRepo.FindByEmployeeID(r.Context(), req.EmployeeID)
	if err != nil || user == nil {
		writeError(w, http.StatusUnauthorized, "IDまたはパスワードが正しくありません")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "IDまたはパスワードが正しくありません")
		return
	}

	_, tokenStr, _ := h.tokenAuth.Encode(map[string]any{
		"user_id": user.ID,
		"tenant_id": user.TenantID,
		"role":    user.Role,
		"name":    user.Name,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	writeJSON(w, http.StatusOK, model.LoginResponse{Token: tokenStr, User: *user})
}

// ============================================================
// ShiftHandler
// ============================================================

type ShiftHandler struct {
	shiftRepo *repository.ShiftRepository
	userRepo  *repository.UserRepository
	validator *validator.ShiftValidator
}

func NewShiftHandler(
	shiftRepo *repository.ShiftRepository,
	userRepo *repository.UserRepository,
	v *validator.ShiftValidator,
) *ShiftHandler {
	return &ShiftHandler{shiftRepo: shiftRepo, userRepo: userRepo, validator: v}
}

// GET /api/workers  — 作業者一覧（管理者用）
func (h *ShiftHandler) GetWorkers(w http.ResponseWriter, r *http.Request) {
	workers, err := h.userRepo.FindWorkers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "データ取得エラー")
		return
	}
	writeJSON(w, http.StatusOK, workers)
}

// GET /api/shifts/board?from=2025-05-25&to=2025-05-31
// シフトボード表示用（管理者）
func (h *ShiftHandler) GetBoard(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		from = time.Now()
	}
	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		to = from.AddDate(0, 0, 13) // デフォルト2週間
	}

	assignments, err := h.shiftRepo.FindAssignmentsByDateRange(r.Context(), from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "データ取得エラー")
		return
	}
	writeJSON(w, http.StatusOK, assignments)
}

// GET /api/shifts/my?from=2025-05-25&to=2025-05-31
// 自分のシフト確認（作業者）
func (h *ShiftHandler) GetMyShifts(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	from, _ := time.Parse("2006-01-02", fromStr)
	to, _ := time.Parse("2006-01-02", toStr)
	if from.IsZero() {
		from = time.Now()
	}
	if to.IsZero() {
		to = from.AddDate(0, 0, 13)
	}

	all, err := h.shiftRepo.FindAssignmentsByDateRange(r.Context(), from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "データ取得エラー")
		return
	}
	// 自分のアサインのみフィルタ
	var mine []model.ShiftAssignment
	for _, a := range all {
		if a.UserID == userID {
			mine = append(mine, a)
		}
	}
	writeJSON(w, http.StatusOK, mine)
}

// POST /api/sites/{siteID}/shifts/{date}/assign
// アサイン追加（管理者）— バリデーション付き
func (h *ShiftHandler) CreateAssign(w http.ResponseWriter, r *http.Request) {
	siteID, err := strconv.ParseInt(chi.URLParam(r, "siteID"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "siteID が不正です")
		return
	}
	dateStr := chi.URLParam(r, "date")
	workDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "date が不正です（YYYY-MM-DD形式）")
		return
	}

	var req model.AssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "リクエスト形式が不正です")
		return
	}

	// 作業者情報取得（バリデーションメッセージ用）
	users, _ := h.userRepo.FindAll(r.Context())
	userName := ""
	for _, u := range users {
		if u.ID == req.UserID {
			userName = u.Name
			break
		}
	}

	// ★ 二重アサインバリデーション
	if err := h.validator.ValidateAssign(
		r.Context(), req.UserID, userName, workDate, req.TimeSlot, siteID,
	); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	adminID := currentUserID(r)
	id, err := h.shiftRepo.CreateAssignment(r.Context(), model.ShiftAssignment{
		TenantID:  currentTenantID(r),
		SiteID:    siteID,
		UserID:    req.UserID,
		WorkDate:  workDate,
		TimeSlot:  req.TimeSlot,
		CreatedBy: &adminID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "登録エラー")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// DELETE /api/shifts/assign/{id}
func (h *ShiftHandler) DeleteAssign(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := h.shiftRepo.DeleteAssignment(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "削除エラー")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ============================================================
// DailyReportHandler
// ============================================================

type DailyReportHandler struct {
	reportRepo *repository.DailyReportRepository
}

func NewDailyReportHandler(repo *repository.DailyReportRepository) *DailyReportHandler {
	return &DailyReportHandler{reportRepo: repo}
}

// PUT /api/reports/{date}  — 日報登録・更新（作業者）
func (h *DailyReportHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	dateStr := chi.URLParam(r, "date")
	workDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "date が不正です")
		return
	}

	// 過去30日より古い日付は更新不可
	if time.Since(workDate) > 30*24*time.Hour {
		writeError(w, http.StatusForbidden, "30日より前の日報は修正できません")
		return
	}

	var req model.DailyReportUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "リクエスト形式が不正です")
		return
	}

	// 入力バリデーション
	if err := validator.ValidateDailyReport(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	rpt := model.DailyReport{
		TenantID:      currentTenantID(r),
		UserID:        userID,
		WorkDate:      workDate,
		Status:        req.Status,
		SiteID:        req.SiteID,
		SiteID2:       req.SiteID2,
		ClientName:    req.ClientName,
		ManDays:       req.ManDays,
		OvertimeHours: req.OvertimeHours,
		UsedCar:       req.UsedCar,
		Note:          req.Note,
	}

	if err := h.reportRepo.UpsertDailyReport(r.Context(), rpt); err != nil {
		writeError(w, http.StatusInternalServerError, "保存エラー")
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GET /api/reports/my?year=2025&month=5
func (h *DailyReportHandler) GetMyMonthly(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	if year == 0 {
		year = time.Now().Year()
	}
	if month == 0 {
		month = int(time.Now().Month())
	}
	rows, err := h.reportRepo.FindByUserMonth(r.Context(), userID, year, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "データ取得エラー")
		return
	}
	// 未入力日も返す
	missing, _ := h.reportRepo.FindMonthlyMissing(r.Context(), userID, year, month)
	writeJSON(w, http.StatusOK, map[string]any{
		"reports": rows,
		"missing": missing,
	})
}

// PUT /api/reports/site-client  — 現場×月単位で元請名を一括更新
func (h *DailyReportHandler) UpdateSiteClient(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	var req struct {
		Year       int    `json:"year"`
		Month      int    `json:"month"`
		SiteID     int64  `json:"site_id"`
		ClientName string `json:"client_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.SiteID == 0 {
		writeError(w, http.StatusBadRequest, "リクエスト形式が不正です")
		return
	}
	if req.Year == 0 || req.Month == 0 {
		now := time.Now()
		req.Year, req.Month = now.Year(), int(now.Month())
	}
	var cn *string
	if req.ClientName != "" {
		cn = &req.ClientName
	}
	if err := h.reportRepo.UpdateClientNameBySiteMonth(
		r.Context(), userID, req.Year, req.Month, req.SiteID, cn,
	); err != nil {
		writeError(w, http.StatusInternalServerError, "更新エラー")
		return
	}
	w.WriteHeader(http.StatusOK)
}

// POST /api/reports/submit?year=2025&month=5  — 月報提出
func (h *DailyReportHandler) Submit(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r)
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	if err := h.reportRepo.SubmitMonthlyReport(r.Context(), userID, year, month); err != nil {
		writeError(w, http.StatusInternalServerError, "提出エラー")
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GET /api/reports/summary?year=2025&month=5  — 月次サマリ（管理者）
func (h *DailyReportHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	if year == 0 {
		year = time.Now().Year()
	}
	if month == 0 {
		month = int(time.Now().Month())
	}
	rows, err := h.reportRepo.GetMonthlySummary(r.Context(), year, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "データ取得エラー")
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

// ============================================================
// SiteHandler
// ============================================================

type SiteHandler struct {
	siteRepo *repository.SiteRepository
}

func NewSiteHandler(siteRepo *repository.SiteRepository) *SiteHandler {
	return &SiteHandler{siteRepo: siteRepo}
}

// GET /api/sites
func (h *SiteHandler) List(w http.ResponseWriter, r *http.Request) {
	sites, err := h.siteRepo.FindAll(r.Context(), currentTenantID(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "データ取得エラー")
		return
	}
	writeJSON(w, http.StatusOK, sites)
}

// GET /api/sites/{id}
func (h *SiteHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id が不正です")
		return
	}
	site, err := h.siteRepo.FindByID(r.Context(), currentTenantID(r), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "データ取得エラー")
		return
	}
	if site == nil {
		writeError(w, http.StatusNotFound, "現場が見つかりません")
		return
	}
	writeJSON(w, http.StatusOK, site)
}

// POST /api/sites
func (h *SiteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.SiteUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "リクエスト形式が不正です")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "現場名は必須です")
		return
	}
	if req.Status == "" {
		req.Status = "active"
	}

	site := model.Site{
		TenantID:  currentTenantID(r),
		Name:      req.Name,
		Client:    req.Client,
		Address:   req.Address,
		BudgetYen: req.BudgetYen,
		Note:      req.Note,
		Status:    req.Status,
	}
	createdBy := currentUserID(r)
	site.CreatedBy = &createdBy

	if req.StartDate != nil && *req.StartDate != "" {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "start_date の形式が不正です（YYYY-MM-DD）")
			return
		}
		site.StartDate = &t
	}
	if req.EndDate != nil && *req.EndDate != "" {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "end_date の形式が不正です（YYYY-MM-DD）")
			return
		}
		site.EndDate = &t
	}

	id, err := h.siteRepo.Create(r.Context(), site)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "登録エラー")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// PUT /api/sites/{id}
func (h *SiteHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id が不正です")
		return
	}

	// 存在確認
	existing, err := h.siteRepo.FindByID(r.Context(), currentTenantID(r), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "データ取得エラー")
		return
	}
	if existing == nil {
		writeError(w, http.StatusNotFound, "現場が見つかりません")
		return
	}

	var req model.SiteUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "リクエスト形式が不正です")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "現場名は必須です")
		return
	}

	site := model.Site{
		ID:        id,
		TenantID:  currentTenantID(r),
		Name:      req.Name,
		Client:    req.Client,
		Address:   req.Address,
		BudgetYen: req.BudgetYen,
		Note:      req.Note,
		Status:    req.Status,
	}
	if site.Status == "" {
		site.Status = existing.Status
	}

	if req.StartDate != nil && *req.StartDate != "" {
		t, err := time.Parse("2006-01-02", *req.StartDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "start_date の形式が不正です（YYYY-MM-DD）")
			return
		}
		site.StartDate = &t
	}
	if req.EndDate != nil && *req.EndDate != "" {
		t, err := time.Parse("2006-01-02", *req.EndDate)
		if err != nil {
			writeError(w, http.StatusBadRequest, "end_date の形式が不正です（YYYY-MM-DD）")
			return
		}
		site.EndDate = &t
	}

	if err := h.siteRepo.Update(r.Context(), site); err != nil {
		writeError(w, http.StatusInternalServerError, "更新エラー")
		return
	}
	w.WriteHeader(http.StatusOK)
}

// ============================================================
// LockHandler
// ============================================================

type LockHandler struct {
	lockRepo *repository.LockRepository
}

func NewLockHandler(repo *repository.LockRepository) *LockHandler {
	return &LockHandler{lockRepo: repo}
}

// GET /api/shifts/lock?year=Y&month=M
func (h *LockHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	tenantID := currentTenantID(r)
	year, _  := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	if year == 0 || month == 0 {
		now := time.Now()
		year, month = now.Year(), int(now.Month())
	}
	lock, err := h.lockRepo.GetStatus(r.Context(), tenantID, year, month)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "取得エラー")
		return
	}
	if lock == nil {
		writeJSON(w, http.StatusOK, map[string]any{"locked": false})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"locked":    true,
		"locked_at": lock.LockedAt,
		"locked_by": lock.LockedBy,
	})
}

// POST /api/admin/shifts/lock  body: { year, month }
func (h *LockHandler) Lock(w http.ResponseWriter, r *http.Request) {
	tenantID := currentTenantID(r)
	userID   := currentUserID(r)
	var req struct {
		Year  int `json:"year"`
		Month int `json:"month"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Year == 0 || req.Month == 0 {
		writeError(w, http.StatusBadRequest, "year/month が不正です")
		return
	}
	if err := h.lockRepo.Lock(r.Context(), tenantID, req.Year, req.Month, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "ロックエラー")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"locked": true})
}

// DELETE /api/admin/shifts/lock  body: { year, month }
func (h *LockHandler) Unlock(w http.ResponseWriter, r *http.Request) {
	tenantID := currentTenantID(r)
	var req struct {
		Year  int `json:"year"`
		Month int `json:"month"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Year == 0 || req.Month == 0 {
		writeError(w, http.StatusBadRequest, "year/month が不正です")
		return
	}
	if err := h.lockRepo.Unlock(r.Context(), tenantID, req.Year, req.Month); err != nil {
		writeError(w, http.StatusInternalServerError, "解除エラー")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"locked": false})
}
