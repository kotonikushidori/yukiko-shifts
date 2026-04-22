-- 009_add_qr_token.sql
-- 作業者QRコードログイン用トークンをusersテーブルに追加

ALTER TABLE users ADD COLUMN qr_token TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_qr_token ON users(qr_token) WHERE qr_token IS NOT NULL;
UPDATE users SET qr_token = lower(hex(randomblob(16))) WHERE role = 'worker' AND qr_token IS NULL;
