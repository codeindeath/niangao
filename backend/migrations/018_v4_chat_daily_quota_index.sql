-- 018_v4_chat_daily_quota_index.sql
-- Index for V4 chat daily quota checks.

CREATE INDEX IF NOT EXISTS idx_chat_messages_user_daily_quota
  ON chat_messages (user_id, role, created_at DESC)
  WHERE status <> 'deleted';
