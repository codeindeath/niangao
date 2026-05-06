-- 012_user_deleted_at.sql
-- Add soft delete to users table for admin enable/disable

ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
