-- ============================================
-- 年糕 App — 数据库迁移 004
-- 经验质量体系：二级领域、准入审核、质量打分
-- ============================================

-- 1. 二级领域
ALTER TABLE experiences ADD COLUMN sub_domain VARCHAR(50);

-- 2. 私密经验
ALTER TABLE experiences ADD COLUMN is_private BOOLEAN NOT NULL DEFAULT FALSE;

-- 3. 审核状态
CREATE TYPE review_status AS ENUM ('pending', 'approved', 'rejected', 'private');
ALTER TABLE experiences ADD COLUMN review_status review_status NOT NULL DEFAULT 'pending';

-- 4. 审核理由
ALTER TABLE experiences ADD COLUMN review_reason TEXT;

-- 5. 质量分（0-10，一位小数）
ALTER TABLE experiences ADD COLUMN quality_score NUMERIC(3,1);

-- 6. 评分明细（JSON）
ALTER TABLE experiences ADD COLUMN score_details JSONB;

-- 7. 索引
CREATE INDEX idx_exp_sub_domain ON experiences(sub_domain);
CREATE INDEX idx_exp_review_status ON experiences(review_status);
CREATE INDEX idx_exp_quality ON experiences(quality_score DESC);

-- 8. 更新现有数据：标记为已审核（之前的是种子数据）
UPDATE experiences SET review_status = 'approved' WHERE is_official = TRUE;
UPDATE experiences SET review_status = 'private' WHERE is_official = FALSE;
