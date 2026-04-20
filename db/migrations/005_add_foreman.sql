-- 005_add_foreman.sql
-- 職長機能: 資格フラグ / 現場別優先順位 / 確定アサイン

-- 職長資格フラグをユーザーテーブルに追加
ALTER TABLE users ADD COLUMN is_foreman_qualified INTEGER NOT NULL DEFAULT 0;

-- 現場ごとの職長候補優先順位（1位が最優先）
CREATE TABLE IF NOT EXISTS site_foreman_priorities (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    site_id        INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    user_id        INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    priority_order INTEGER NOT NULL DEFAULT 0,
    UNIQUE(site_id, user_id)
);

-- 職長確定アサイン（現場×日で一意）
CREATE TABLE IF NOT EXISTS foreman_assignments (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id  INTEGER NOT NULL REFERENCES tenants(id),
    site_id    INTEGER NOT NULL REFERENCES sites(id) ON DELETE CASCADE,
    work_date  TEXT    NOT NULL,
    user_id    INTEGER NOT NULL REFERENCES users(id),
    is_manual  INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(site_id, work_date)
);
