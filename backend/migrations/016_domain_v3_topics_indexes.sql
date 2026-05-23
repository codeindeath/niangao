-- 年糕 App — 数据库迁移 016
-- 领域体系 v3、话题字段、推荐排序字段与公开池索引

-- domain_type enum 无法容纳 v3 新领域；转为 VARCHAR 便于后续领域体系演进。
ALTER TABLE experiences
  ALTER COLUMN domain TYPE VARCHAR(50) USING domain::text;

DROP TYPE IF EXISTS domain_type;

ALTER TABLE experiences
  ADD COLUMN IF NOT EXISTS topics VARCHAR(200) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS random_sort DOUBLE PRECISION NOT NULL DEFAULT random();

UPDATE experiences SET topics = '' WHERE topics IS NULL;
UPDATE experiences SET domain = '' WHERE domain IS NULL;
UPDATE experiences SET random_sort = random() WHERE random_sort IS NULL;

ALTER TABLE experiences
  ALTER COLUMN domain SET DEFAULT '',
  ALTER COLUMN topics SET DEFAULT '',
  ALTER COLUMN topics SET NOT NULL,
  ALTER COLUMN random_sort SET DEFAULT random(),
  ALTER COLUMN random_sort SET NOT NULL;

-- 将历史 v2/v2.5 领域与子领域迁移到 v3。已是 v3 的数据不会被改动。
UPDATE experiences
SET sub_domain = CASE sub_domain
  WHEN 'career-planning' THEN 'jobhunt'
  WHEN 'skill-building' THEN 'productivity'
  WHEN 'side-hustle' THEN 'startup'
  WHEN 'workplace-comm' THEN 'work-comm'
  WHEN 'mental-model' THEN 'thinking'
  WHEN 'learning' THEN 'cog-learning'
  WHEN 'decision' THEN 'thinking'
  WHEN 'psychology' THEN 'self'
  WHEN 'finance' THEN 'shopping'
  WHEN 'time-mgmt' THEN 'productivity'
  WHEN 'habits' THEN 'selfcare'
  WHEN 'digital-life' THEN 'tools'
  WHEN 'regulation' THEN 'health'
  WHEN 'self-growth' THEN 'self'
  WHEN 'stress-mgmt' THEN 'health'
  WHEN 'intimate' THEN 'romance'
  WHEN 'family' THEN 'parents'
  WHEN 'social-skill' THEN 'friendship'
  WHEN 'communication' THEN 'friendship'
  ELSE sub_domain
END
WHERE sub_domain IN (
  'career-planning', 'skill-building', 'side-hustle', 'workplace-comm',
  'mental-model', 'learning', 'decision', 'psychology',
  'finance', 'time-mgmt', 'habits', 'digital-life',
  'regulation', 'self-growth', 'stress-mgmt',
  'intimate', 'family', 'social-skill', 'communication'
);

UPDATE experiences
SET sub_domain = CASE domain
  WHEN 'career' THEN 'jobhunt'
  WHEN 'life' THEN 'selfcare'
  WHEN 'emotion' THEN 'self'
  ELSE sub_domain
END
WHERE sub_domain IS NULL OR sub_domain = '';

UPDATE experiences
SET domain = CASE
  WHEN sub_domain IN ('health', 'housing', 'transit', 'diet', 'exercise') THEN 'vitality'
  WHEN sub_domain IN ('pets', 'travel', 'fashion', 'selfcare', 'shopping', 'fun') THEN 'living'
  WHEN sub_domain IN ('jobhunt', 'promotion', 'startup', 'work-comm', 'management', 'productivity') THEN 'work'
  WHEN sub_domain IN ('marriage', 'romance', 'friendship', 'parenting', 'parents', 'siblings') THEN 'relationship'
  WHEN sub_domain IN ('cog-learning', 'thinking', 'info', 'tools', 'creativity', 'expression') THEN 'cognition'
  WHEN sub_domain IN ('self', 'happiness', 'faith', 'mission', 'belonging') THEN 'meaning'
  WHEN domain = 'career' THEN 'work'
  WHEN domain = 'life' THEN 'living'
  WHEN domain = 'emotion' THEN 'meaning'
  ELSE domain
END;

CREATE INDEX IF NOT EXISTS idx_exp_public_feed
  ON experiences (created_at DESC)
  WHERE status = 'published' AND review_status = 'approved' AND is_private = FALSE AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_exp_public_domain_feed
  ON experiences (domain, created_at DESC)
  WHERE status = 'published' AND review_status = 'approved' AND is_private = FALSE AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_exp_public_random
  ON experiences (random_sort)
  WHERE status = 'published' AND review_status = 'approved' AND is_private = FALSE AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_bookmarks_user_created
  ON bookmarks (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_user_views_user_exp
  ON user_views (user_id, experience_id);
