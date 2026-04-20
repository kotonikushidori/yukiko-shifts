package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/yourorg/shift-app/internal/model"
)

// ============================================================
// ForemanRepository
// ============================================================

type ForemanRepository struct{ db *sqlx.DB }

func NewForemanRepository(db *sqlx.DB) *ForemanRepository { return &ForemanRepository{db: db} }

// GetPriorities は現場の職長優先順位リストを優先順に返す
func (r *ForemanRepository) GetPriorities(ctx context.Context, siteID int64) ([]model.ForemanPriority, error) {
	var rows []model.ForemanPriority
	err := r.db.SelectContext(ctx, &rows, `
		SELECT sfp.*, u.name AS user_name
		FROM site_foreman_priorities sfp
		JOIN users u ON sfp.user_id = u.id
		WHERE sfp.site_id = ?
		ORDER BY sfp.priority_order ASC`, siteID)
	return rows, err
}

// SetPriorities は現場の職長優先順位リストを全置換する
func (r *ForemanRepository) SetPriorities(ctx context.Context, siteID int64, items []model.ForemanPriority) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM site_foreman_priorities WHERE site_id = ?`, siteID); err != nil {
		return err
	}
	for i, item := range items {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO site_foreman_priorities (site_id, user_id, priority_order)
			VALUES (?, ?, ?)`, siteID, item.UserID, i); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetAssignments は指定日付範囲の職長アサインを返す
func (r *ForemanRepository) GetAssignments(ctx context.Context, tenantID int64, from, to string) ([]model.ForemanAssignment, error) {
	var rows []model.ForemanAssignment
	err := r.db.SelectContext(ctx, &rows, `
		SELECT fa.*, u.name AS user_name, s.name AS site_name
		FROM foreman_assignments fa
		JOIN users u ON fa.user_id = u.id
		JOIN sites s ON fa.site_id = s.id
		WHERE fa.tenant_id = ? AND fa.work_date BETWEEN ? AND ?
		ORDER BY fa.work_date, s.name`, tenantID, from, to)
	return rows, err
}

// UpsertAssignment は職長アサインを登録/更新する
func (r *ForemanRepository) UpsertAssignment(
	ctx context.Context, tenantID, siteID int64, workDate string, userID int64, isManual bool,
) error {
	isManualInt := 0
	if isManual {
		isManualInt = 1
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO foreman_assignments (tenant_id, site_id, work_date, user_id, is_manual)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(site_id, work_date) DO UPDATE SET
			user_id   = excluded.user_id,
			is_manual = excluded.is_manual`,
		tenantID, siteID, workDate, userID, isManualInt)
	return err
}

// DeleteAssignment は職長アサインを削除する
func (r *ForemanRepository) DeleteAssignment(ctx context.Context, tenantID, siteID int64, workDate string) error {
	_, err := r.db.ExecContext(ctx,
		`DELETE FROM foreman_assignments WHERE tenant_id=? AND site_id=? AND work_date=?`,
		tenantID, siteID, workDate)
	return err
}

// SuggestForeman は指定現場×日付で優先順位最上位の在勤職長資格者を返す
// 返値 userID が nil の場合は候補なし
func (r *ForemanRepository) SuggestForeman(ctx context.Context, siteID int64, workDate string) (userID *int64, userName string, err error) {
	var row struct {
		UserID   int64  `db:"user_id"`
		UserName string `db:"user_name"`
	}
	err = r.db.GetContext(ctx, &row, `
		SELECT sfp.user_id, u.name AS user_name
		FROM site_foreman_priorities sfp
		JOIN users u ON sfp.user_id = u.id
		JOIN shift_assignments sa
		  ON sa.user_id  = sfp.user_id
		 AND sa.site_id  = ?
		 AND sa.work_date = ?
		WHERE sfp.site_id = ? AND u.is_foreman_qualified = 1
		ORDER BY sfp.priority_order ASC
		LIMIT 1`, siteID, workDate, siteID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", err
	}
	return &row.UserID, row.UserName, nil
}

// GetQualifiedCandidates は指定現場×日付に出勤予定の職長資格者を返す
func (r *ForemanRepository) GetQualifiedCandidates(ctx context.Context, siteID int64, workDate string) ([]model.ForemanCandidate, error) {
	var rows []struct {
		UserID   int64  `db:"user_id"`
		UserName string `db:"user_name"`
	}
	err := r.db.SelectContext(ctx, &rows, `
		SELECT DISTINCT sa.user_id, u.name AS user_name
		FROM shift_assignments sa
		JOIN users u ON sa.user_id = u.id
		WHERE sa.site_id = ? AND sa.work_date = ? AND u.is_foreman_qualified = 1
		ORDER BY u.name`, siteID, workDate)
	if err != nil {
		return nil, err
	}

	candidates := make([]model.ForemanCandidate, len(rows))
	for i, row := range rows {
		candidates[i] = model.ForemanCandidate{UserID: row.UserID, UserName: row.UserName}
	}
	return candidates, nil
}
