-- 002: daily_reports に元請名カラムを追加
-- SQLite は ADD COLUMN IF NOT EXISTS 非対応のため try/ignore で対応
ALTER TABLE daily_reports ADD COLUMN client_name TEXT;
