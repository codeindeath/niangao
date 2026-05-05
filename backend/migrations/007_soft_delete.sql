-- ============================================
-- 年糕 App — 数据库迁移 007
-- 软删除：核准经验软删（保留在公共池），其他硬删
-- ============================================

ALTER TABLE experiences ADD COLUMN deleted_at TIMESTAMPTZ;

CREATE INDEX idx_exp_deleted_at ON experiences(deleted_at);
