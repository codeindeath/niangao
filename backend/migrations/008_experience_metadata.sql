-- ============================================
-- 年糕 App — 数据库迁移 008
-- 经验元信息：创作者、来源类型、打分理由
-- ============================================

-- 1. 原内容创作者（书籍作者、名人等）
ALTER TABLE experiences ADD COLUMN creator_name VARCHAR(100);

-- 2. 来源类型：platform（平台生产）vs user（用户原创）
ALTER TABLE experiences ADD COLUMN source_type VARCHAR(20) NOT NULL DEFAULT 'user'
  CHECK (source_type IN ('platform', 'user'));

-- 3. 打分理由（≤15 字简短说明）
ALTER TABLE experiences ADD COLUMN score_reason VARCHAR(100);

-- 4. 索引
CREATE INDEX idx_exp_source_type ON experiences(source_type);
