-- 003: users に苗字・名前カラムを追加
ALTER TABLE users ADD COLUMN last_name  TEXT;
ALTER TABLE users ADD COLUMN first_name TEXT;
