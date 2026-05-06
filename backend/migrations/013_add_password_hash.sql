-- 013: Add password_hash column to users for admin login
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);
