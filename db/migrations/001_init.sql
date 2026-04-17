-- ============================================================
-- 施工会社シフト管理システム DB設計 v2
-- マルチテナント対応版
-- ============================================================

-- ──────────────────────────────────────
-- テナントマスタ（契約事業者）
-- ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS tenants (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    name         TEXT    NOT NULL,              -- 会社名
    slug         TEXT    NOT NULL UNIQUE,       -- URLスラッグ（例: nakaya-sakan）
    plan         TEXT    NOT NULL DEFAULT 'basic', -- 'basic' | 'pro'
    max_workers  INTEGER NOT NULL DEFAULT 30,   -- プランごとの上限人数
    status       TEXT    NOT NULL DEFAULT 'active', -- 'active' | 'suspended' | 'cancelled'
    contract_start DATE,
    contract_end   DATE,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ──────────────────────────────────────
-- ユーザー（全テナント共通テーブル）
-- ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS users (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id     INTEGER NOT NULL REFERENCES tenants(id),
    employee_id   TEXT    NOT NULL,             -- 社員ID（テナント内でユニーク）
    email         TEXT,
    password_hash TEXT    NOT NULL,
    name          TEXT    NOT NULL,
    role          TEXT    NOT NULL DEFAULT 'worker', -- 'superadmin' | 'admin' | 'worker'
    phone         TEXT,
    status        TEXT    NOT NULL DEFAULT 'active',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(tenant_id, employee_id)              -- テナント内でemployee_idはユニーク
);

-- ──────────────────────────────────────
-- 現場マスタ
-- ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS sites (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id    INTEGER NOT NULL REFERENCES tenants(id),
    name         TEXT    NOT NULL,
    client       TEXT,                          -- 元請会社
    address      TEXT,
    budget_yen   INTEGER,
    start_date   DATE,
    end_date     DATE,
    note         TEXT,
    status       TEXT    NOT NULL DEFAULT 'active', -- 'active' | 'completed' | 'deleted'
    created_by   INTEGER REFERENCES users(id),
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ──────────────────────────────────────
-- シフト設定（現場×日付×作業者）
-- ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS shift_assignments (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id   INTEGER NOT NULL REFERENCES tenants(id),
    site_id     INTEGER NOT NULL REFERENCES sites(id),
    user_id     INTEGER NOT NULL REFERENCES users(id),
    work_date   DATE    NOT NULL,
    time_slot   TEXT    NOT NULL,               -- 'AM' | 'PM' | 'ALL'
    created_by  INTEGER REFERENCES users(id),
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(tenant_id, site_id, user_id, work_date, time_slot)
);

-- ──────────────────────────────────────
-- 日報
-- ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS daily_reports (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id      INTEGER NOT NULL REFERENCES tenants(id),
    user_id        INTEGER NOT NULL REFERENCES users(id),
    work_date      DATE    NOT NULL,
    status         TEXT    NOT NULL DEFAULT 'present', -- 'present' | 'absent' | 'half'
    site_id        INTEGER REFERENCES sites(id),
    site_id2       INTEGER REFERENCES sites(id),
    man_days       REAL    DEFAULT 1.0,
    overtime_hours REAL    DEFAULT 0.0,
    used_car       INTEGER DEFAULT 0,
    note           TEXT,
    submitted_at   DATETIME,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(tenant_id, user_id, work_date)
);

-- ──────────────────────────────────────
-- 全体連絡
-- ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS announcements (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id   INTEGER NOT NULL REFERENCES tenants(id),
    title       TEXT    NOT NULL,
    body        TEXT    NOT NULL,
    created_by  INTEGER NOT NULL REFERENCES users(id),
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- ──────────────────────────────────────
-- 既読管理
-- ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS announcement_reads (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    announcement_id INTEGER NOT NULL REFERENCES announcements(id),
    user_id         INTEGER NOT NULL REFERENCES users(id),
    read_at         DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(announcement_id, user_id)
);

-- ──────────────────────────────────────
-- シフトロック（希望入力の締め切り管理）
-- ──────────────────────────────────────
CREATE TABLE IF NOT EXISTS shift_locks (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id  INTEGER NOT NULL REFERENCES tenants(id),
    year       INTEGER NOT NULL,
    month      INTEGER NOT NULL,
    locked_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    locked_by  INTEGER NOT NULL REFERENCES users(id),
    UNIQUE(tenant_id, year, month)
);

-- ============================================================
-- インデックス
-- ============================================================

-- テナント絞り込みが全クエリで走るので tenant_id を先頭に
CREATE INDEX IF NOT EXISTS idx_users_tenant
    ON users(tenant_id, status);

CREATE INDEX IF NOT EXISTS idx_sites_tenant
    ON sites(tenant_id, status);

CREATE INDEX IF NOT EXISTS idx_shift_tenant_user_date
    ON shift_assignments(tenant_id, user_id, work_date);

CREATE INDEX IF NOT EXISTS idx_shift_tenant_site_date
    ON shift_assignments(tenant_id, site_id, work_date);

CREATE INDEX IF NOT EXISTS idx_report_tenant_user_date
    ON daily_reports(tenant_id, user_id, work_date);

CREATE INDEX IF NOT EXISTS idx_report_tenant_date
    ON daily_reports(tenant_id, work_date);

CREATE INDEX IF NOT EXISTS idx_announce_tenant
    ON announcements(tenant_id, created_at);

-- ============================================================
-- 初期データ（開発用）
-- ============================================================

-- 開発用テナント
INSERT OR IGNORE INTO tenants (id, name, slug, plan, max_workers)
VALUES (1, '中屋敷左官工業株式会社', 'nakaya-sakan', 'pro', 50);

-- 開発用管理者（パスワード: admin1234 → 本番前に必ず変更）
INSERT OR IGNORE INTO users (tenant_id, employee_id, name, role, password_hash)
VALUES (1, 'admin', '管理者', 'admin',
        '$2a$10$2nldkLeypKnx0KJiLicK4.EtluomWZHKQ/KJo1BY3Vaj4iwyL2j86');

-- 開発用現場マスタ
INSERT OR IGNORE INTO sites (tenant_id, name, status) VALUES
  (1, '南平岸４条',           'active'),
  (1, '島松隊舎',             'active'),
  (1, 'CR澄川',               'active'),
  (1, 'すすきの駅前',         'active'),
  (1, 'トヨタ分室',           'active'),
  (1, '北広工業',             'active'),
  (1, '住材モールアックス',   'active'),
  (1, '新単団地',             'active'),
  (1, '第一生命ビルスロープ', 'active'),
  (1, '北１０東４',           'active'),
  (1, 'CR札幌東',             'active'),
  (1, 'CMS宮の沢',            'active'),
  (1, 'CMS北３東７',          'active'),
  (1, '豊平４−８',           'active'),
  (1, '山鼻モールアックス',   'active');

-- 開発用作業者（パスワード: worker1234）
INSERT OR IGNORE INTO users (tenant_id, employee_id, name, role, password_hash) VALUES
  (1, 'w001', '作業者01', 'worker', '$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1, 'w002', '作業者02', 'worker', '$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1, 'w003', '作業者03', 'worker', '$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm'),
  (1, 'w004', '作業者04', 'worker', '$2a$10$lK7o4tvufPn7hiNeMMuhEerIEK8K/Snge9PbCNNvWPtqcwMEV7XBm');
