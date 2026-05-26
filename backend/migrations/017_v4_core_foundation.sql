-- 年糕 App — 数据库迁移 017
-- V4 核心事实源：经验字段、互动事件、推荐会话、聊聊议题、AI Gateway 基座
-- 说明：
-- 1. 本迁移只增加 V4 事实源和并行表，不删除旧字段 / 旧表。
-- 2. 旧 likes/bookmarks/review_status/status 字段继续作为兼容来源，后续 API 逐步迁到 V4 表。
-- 3. 用户侧公开审核失败不再作为前台状态表达；V4 使用 visibility/lifecycle/quality_tier/recommendation_status 控制分发。

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- Users: 展示名和轻设置
-- ============================================================

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS display_name VARCHAR(30),
  ADD COLUMN IF NOT EXISTS display_name_updated_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS user_settings JSONB NOT NULL DEFAULT '{}'::jsonb;

UPDATE users
SET display_name = NULLIF(TRIM(nickname), '')
WHERE display_name IS NULL;

CREATE INDEX IF NOT EXISTS idx_users_display_name
  ON users (display_name)
  WHERE display_name IS NOT NULL AND deleted_at IS NULL;

-- ============================================================
-- Experiences: V4 canonical fields
-- ============================================================

ALTER TABLE experiences
  ADD COLUMN IF NOT EXISTS owner_user_id UUID,
  ADD COLUMN IF NOT EXISTS creator_id UUID,
  ADD COLUMN IF NOT EXISTS creator_display_name VARCHAR(100),
  ADD COLUMN IF NOT EXISTS experience_type VARCHAR(32) NOT NULL DEFAULT 'user_original',
  ADD COLUMN IF NOT EXISTS visibility VARCHAR(16) NOT NULL DEFAULT 'public',
  ADD COLUMN IF NOT EXISTS lifecycle_status VARCHAR(32) NOT NULL DEFAULT 'active',
  ADD COLUMN IF NOT EXISTS source_scene VARCHAR(32) NOT NULL DEFAULT 'note',
  ADD COLUMN IF NOT EXISTS topic VARCHAR(200) NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS quality_tier VARCHAR(32) NOT NULL DEFAULT 'unreviewed',
  ADD COLUMN IF NOT EXISTS source_reliability VARCHAR(16) NOT NULL DEFAULT 'medium',
  ADD COLUMN IF NOT EXISTS recommendation_status VARCHAR(32) NOT NULL DEFAULT 'ineligible',
  ADD COLUMN IF NOT EXISTS ai_citable BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS inspiration_count INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS collection_count INTEGER NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS interpretation_status VARCHAR(16) NOT NULL DEFAULT 'none',
  ADD COLUMN IF NOT EXISTS source_material_id UUID,
  ADD COLUMN IF NOT EXISTS production_batch_id UUID,
  ADD COLUMN IF NOT EXISTS source_chat_topic_id UUID,
  ADD COLUMN IF NOT EXISTS source_chat_message_id UUID,
  ADD COLUMN IF NOT EXISTS source_chat_message_snapshot TEXT,
  ADD COLUMN IF NOT EXISTS system_visibility_reason TEXT,
  ADD COLUMN IF NOT EXISTS public_processed_at TIMESTAMPTZ;

UPDATE experiences
SET
  owner_user_id = COALESCE(owner_user_id, author_id),
  experience_type = CASE
    WHEN COALESCE(source_type, '') = 'platform' OR is_official = TRUE THEN 'platform_selected'
    ELSE 'user_original'
  END,
  visibility = CASE
    WHEN is_private = TRUE THEN 'private'
    ELSE 'public'
  END,
  lifecycle_status = CASE
    WHEN deleted_at IS NOT NULL THEN 'deleted'
    WHEN status IN ('hidden', 'flagged') THEN 'hidden'
    WHEN review_status = 'rejected' THEN 'hidden'
    ELSE 'active'
  END,
  topic = COALESCE(NULLIF(topic, ''), topics, ''),
  inspiration_count = COALESCE(like_count, 0),
  collection_count = COALESCE(bookmark_count, 0),
  interpretation_status = CASE
    WHEN interpretation IS NOT NULL AND TRIM(interpretation) <> '' THEN 'ready'
    WHEN interpretation_generated = TRUE THEN 'ready'
    ELSE COALESCE(NULLIF(interpretation_status, ''), 'none')
  END,
  creator_display_name = COALESCE(NULLIF(creator_display_name, ''), NULLIF(creator_name, '')),
  source_scene = COALESCE(NULLIF(source_scene, ''), 'note');

UPDATE experiences e
SET creator_display_name = COALESCE(NULLIF(e.creator_display_name, ''), NULLIF(u.display_name, ''), NULLIF(u.nickname, ''))
FROM users u
WHERE e.author_id = u.id
  AND (e.creator_display_name IS NULL OR e.creator_display_name = '');

UPDATE experiences
SET quality_tier = CASE
  WHEN visibility = 'private' THEN 'private_only'
  WHEN lifecycle_status <> 'active' THEN 'unreviewed'
  WHEN review_status = 'pending' THEN 'unreviewed'
  WHEN review_status = 'private' THEN 'private_only'
  WHEN review_status = 'approved' AND quality_score >= 9.0 THEN 'high_trust'
  WHEN review_status = 'approved' AND quality_score >= 8.0 THEN 'ai_citable'
  WHEN review_status = 'approved' AND quality_score >= 6.5 THEN 'recommend_candidate'
  WHEN review_status = 'approved' THEN 'public_visible'
  ELSE COALESCE(NULLIF(quality_tier, ''), 'unreviewed')
END;

UPDATE experiences
SET
  recommendation_status = CASE
    WHEN visibility = 'public'
      AND lifecycle_status = 'active'
      AND quality_tier IN ('recommend_candidate', 'ai_citable', 'high_trust')
    THEN 'eligible'
    ELSE 'ineligible'
  END,
  ai_citable = CASE
    WHEN visibility = 'public'
      AND lifecycle_status = 'active'
      AND quality_tier IN ('ai_citable', 'high_trust')
    THEN TRUE
    ELSE FALSE
  END;

ALTER TABLE experiences
  ALTER COLUMN owner_user_id SET NOT NULL,
  ALTER COLUMN experience_type SET NOT NULL,
  ALTER COLUMN visibility SET NOT NULL,
  ALTER COLUMN lifecycle_status SET NOT NULL,
  ALTER COLUMN source_scene SET NOT NULL,
  ALTER COLUMN topic SET NOT NULL,
  ALTER COLUMN quality_tier SET NOT NULL,
  ALTER COLUMN source_reliability SET NOT NULL,
  ALTER COLUMN recommendation_status SET NOT NULL,
  ALTER COLUMN ai_citable SET NOT NULL,
  ALTER COLUMN inspiration_count SET NOT NULL,
  ALTER COLUMN collection_count SET NOT NULL,
  ALTER COLUMN interpretation_status SET NOT NULL;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'experiences_owner_user_id_fkey') THEN
    ALTER TABLE experiences
      ADD CONSTRAINT experiences_owner_user_id_fkey
      FOREIGN KEY (owner_user_id) REFERENCES users(id) ON DELETE CASCADE NOT VALID;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'experiences_experience_type_check') THEN
    ALTER TABLE experiences
      ADD CONSTRAINT experiences_experience_type_check
      CHECK (experience_type IN ('platform_selected', 'user_original')) NOT VALID;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'experiences_visibility_check') THEN
    ALTER TABLE experiences
      ADD CONSTRAINT experiences_visibility_check
      CHECK (visibility IN ('public', 'private')) NOT VALID;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'experiences_lifecycle_status_check') THEN
    ALTER TABLE experiences
      ADD CONSTRAINT experiences_lifecycle_status_check
      CHECK (lifecycle_status IN ('active', 'hidden', 'deleted', 'needs_review')) NOT VALID;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'experiences_quality_tier_check') THEN
    ALTER TABLE experiences
      ADD CONSTRAINT experiences_quality_tier_check
      CHECK (quality_tier IN ('unreviewed', 'private_only', 'public_visible', 'recommend_candidate', 'ai_citable', 'high_trust')) NOT VALID;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'experiences_recommendation_status_check') THEN
    ALTER TABLE experiences
      ADD CONSTRAINT experiences_recommendation_status_check
      CHECK (recommendation_status IN ('eligible', 'ineligible', 'suppressed')) NOT VALID;
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'experiences_interpretation_status_check') THEN
    ALTER TABLE experiences
      ADD CONSTRAINT experiences_interpretation_status_check
      CHECK (interpretation_status IN ('none', 'pending', 'ready', 'stale', 'failed')) NOT VALID;
  END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_exp_v4_owner_created
  ON experiences (owner_user_id, created_at DESC, id DESC)
  WHERE lifecycle_status <> 'deleted';

CREATE INDEX IF NOT EXISTS idx_exp_v4_public_recommend
  ON experiences (recommendation_status, quality_tier, created_at DESC)
  WHERE visibility = 'public' AND lifecycle_status = 'active' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_exp_v4_public_ai_citable
  ON experiences (ai_citable, quality_tier, created_at DESC)
  WHERE visibility = 'public' AND lifecycle_status = 'active' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_exp_v4_type_quality
  ON experiences (experience_type, quality_tier, created_at DESC)
  WHERE visibility = 'public' AND lifecycle_status = 'active' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_exp_v4_domain_subdomain
  ON experiences (domain, sub_domain, created_at DESC)
  WHERE visibility = 'public' AND lifecycle_status = 'active' AND deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_exp_v4_topic
  ON experiences (topic)
  WHERE topic <> '' AND deleted_at IS NULL;

-- ============================================================
-- Collections and inspirations
-- ============================================================

CREATE TABLE IF NOT EXISTS experience_collections (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  experience_id UUID NOT NULL REFERENCES experiences(id) ON DELETE CASCADE,
  status VARCHAR(16) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'removed')),
  source_context VARCHAR(32) NOT NULL DEFAULT 'feed',
  collected_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  removed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, experience_id)
);

CREATE INDEX IF NOT EXISTS idx_exp_collections_user_active
  ON experience_collections (user_id, collected_at DESC, id DESC)
  WHERE status = 'active';

CREATE INDEX IF NOT EXISTS idx_exp_collections_exp_active
  ON experience_collections (experience_id)
  WHERE status = 'active';

INSERT INTO experience_collections (user_id, experience_id, status, source_context, collected_at, created_at, updated_at)
SELECT b.user_id, b.experience_id, 'active', 'legacy_bookmark', b.created_at, b.created_at, b.created_at
FROM bookmarks b
ON CONFLICT (user_id, experience_id) DO NOTHING;

CREATE TABLE IF NOT EXISTS experience_inspirations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  experience_id UUID NOT NULL REFERENCES experiences(id) ON DELETE CASCADE,
  source_context VARCHAR(32) NOT NULL DEFAULT 'feed',
  inspired_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (user_id, experience_id)
);

CREATE INDEX IF NOT EXISTS idx_exp_inspirations_user
  ON experience_inspirations (user_id, inspired_at DESC);

CREATE INDEX IF NOT EXISTS idx_exp_inspirations_exp
  ON experience_inspirations (experience_id);

INSERT INTO experience_inspirations (user_id, experience_id, source_context, inspired_at, created_at)
SELECT l.user_id, l.experience_id, 'legacy_like', l.created_at, l.created_at
FROM likes l
ON CONFLICT (user_id, experience_id) DO NOTHING;

-- ============================================================
-- Experience events and metrics
-- ============================================================

CREATE TABLE IF NOT EXISTS experience_events (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  experience_id UUID REFERENCES experiences(id) ON DELETE SET NULL,
  event_type VARCHAR(32) NOT NULL CHECK (event_type IN (
    'expose', 'flip', 'collect', 'uncollect', 'inspire',
    'search_click', 'chat_citation_show', 'chat_citation_click'
  )),
  source_context VARCHAR(32) NOT NULL DEFAULT 'feed',
  context_id UUID,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_exp_events_user_time
  ON experience_events (user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_exp_events_exp_type_time
  ON experience_events (experience_id, event_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_exp_events_context
  ON experience_events (source_context, context_id);

CREATE TABLE IF NOT EXISTS experience_metrics (
  experience_id UUID NOT NULL REFERENCES experiences(id) ON DELETE CASCADE,
  window_type VARCHAR(8) NOT NULL CHECK (window_type IN ('7d', '30d', 'all')),
  exposure_count INTEGER NOT NULL DEFAULT 0,
  qualified_flip_count INTEGER NOT NULL DEFAULT 0,
  collect_count INTEGER NOT NULL DEFAULT 0,
  inspire_count INTEGER NOT NULL DEFAULT 0,
  chat_citation_count INTEGER NOT NULL DEFAULT 0,
  chat_clear_positive_count INTEGER NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (experience_id, window_type)
);

CREATE INDEX IF NOT EXISTS idx_exp_metrics_window
  ON experience_metrics (window_type, updated_at DESC);

INSERT INTO experience_metrics (experience_id, window_type, collect_count, inspire_count)
SELECT id, 'all', collection_count, inspiration_count
FROM experiences
ON CONFLICT (experience_id, window_type) DO UPDATE
SET collect_count = EXCLUDED.collect_count,
    inspire_count = EXCLUDED.inspire_count,
    updated_at = NOW();

-- ============================================================
-- Recommendation sessions and search sessions
-- ============================================================

CREATE TABLE IF NOT EXISTS recommendation_sessions (
  session_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  candidate_ids UUID[] NOT NULL DEFAULT '{}',
  returned_offset INTEGER NOT NULL DEFAULT 0,
  profile_version INTEGER NOT NULL DEFAULT 0,
  sort_seed BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '30 minutes',
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_recommendation_sessions_user
  ON recommendation_sessions (user_id, expires_at DESC);

CREATE INDEX IF NOT EXISTS idx_recommendation_sessions_expires
  ON recommendation_sessions (expires_at);

CREATE TABLE IF NOT EXISTS search_sessions (
  session_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  query TEXT NOT NULL DEFAULT '',
  result_ids UUID[] NOT NULL DEFAULT '{}',
  returned_offset INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '30 minutes'
);

CREATE INDEX IF NOT EXISTS idx_search_sessions_user
  ON search_sessions (user_id, expires_at DESC);

-- ============================================================
-- Chat V4
-- ============================================================

CREATE TABLE IF NOT EXISTS chat_topics (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status VARCHAR(16) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'deleted')),
  title VARCHAR(100),
  domain VARCHAR(50),
  sub_domain VARCHAR(50),
  topic VARCHAR(200) NOT NULL DEFAULT '',
  clarity_score NUMERIC(3,2),
  summary TEXT,
  last_opened_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_chat_topics_user_updated
  ON chat_topics (user_id, updated_at DESC)
  WHERE status = 'active' AND deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS chat_temp_sessions (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  status VARCHAR(16) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'promoted', 'discarded')),
  forced_new_topic BOOLEAN NOT NULL DEFAULT FALSE,
  promoted_topic_id UUID REFERENCES chat_topics(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  discarded_at TIMESTAMPTZ,
  purge_after TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_chat_temp_sessions_user_updated
  ON chat_temp_sessions (user_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS chat_messages (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  topic_id UUID REFERENCES chat_topics(id) ON DELETE CASCADE,
  temp_session_id UUID REFERENCES chat_temp_sessions(id) ON DELETE CASCADE,
  role VARCHAR(16) NOT NULL CHECK (role IN ('user', 'assistant')),
  content TEXT NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'sent' CHECK (status IN ('sent', 'failed', 'deleted')),
  risk_level VARCHAR(16) NOT NULL DEFAULT 'normal' CHECK (risk_level IN ('normal', 'high')),
  client_message_id VARCHAR(100),
  referenced_experience_ids UUID[] NOT NULL DEFAULT '{}',
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  CHECK (topic_id IS NOT NULL OR temp_session_id IS NOT NULL)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_chat_messages_client_id
  ON chat_messages (user_id, client_message_id)
  WHERE client_message_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_chat_messages_topic_time
  ON chat_messages (topic_id, created_at)
  WHERE topic_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_chat_messages_temp_time
  ON chat_messages (temp_session_id, created_at)
  WHERE temp_session_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS chat_citations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  message_id UUID NOT NULL REFERENCES chat_messages(id) ON DELETE CASCADE,
  experience_id UUID NOT NULL REFERENCES experiences(id) ON DELETE CASCADE,
  citation_type VARCHAR(32) NOT NULL CHECK (citation_type IN ('own', 'favorite', 'public_featured', 'public_original')),
  shown_at TIMESTAMPTZ,
  clicked_at TIMESTAMPTZ,
  collected_at TIMESTAMPTZ,
  inspired_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_citations_message
  ON chat_citations (message_id);

CREATE INDEX IF NOT EXISTS idx_chat_citations_experience
  ON chat_citations (experience_id);

-- ============================================================
-- User feedback
-- ============================================================

CREATE TABLE IF NOT EXISTS feedback (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  feedback_type VARCHAR(32) NOT NULL DEFAULT 'general',
  content TEXT NOT NULL,
  app_version VARCHAR(64),
  device VARCHAR(128),
  os_version VARCHAR(64),
  status VARCHAR(32) NOT NULL DEFAULT 'new' CHECK (status IN ('new', 'triaged', 'closed')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feedback_user_time
  ON feedback (user_id, created_at DESC)
  WHERE user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_feedback_status_time
  ON feedback (status, created_at DESC);

-- ============================================================
-- AI Gateway foundation
-- ============================================================

CREATE TABLE IF NOT EXISTS ai_function_configs (
  function_type VARCHAR(64) PRIMARY KEY,
  model VARCHAR(100) NOT NULL DEFAULT 'deepseek-v4-pro',
  key_alias VARCHAR(100) NOT NULL,
  prompt_version VARCHAR(64) NOT NULL,
  schema_version VARCHAR(64) NOT NULL,
  timeout_ms INTEGER NOT NULL DEFAULT 30000,
  max_tokens INTEGER NOT NULL DEFAULT 2048,
  temperature NUMERIC(3,2) NOT NULL DEFAULT 0.70,
  thinking VARCHAR(16) NOT NULL DEFAULT 'disabled' CHECK (thinking IN ('enabled', 'disabled')),
  response_format VARCHAR(32) NOT NULL DEFAULT 'json_object',
  queue_name VARCHAR(64) NOT NULL DEFAULT 'user_normal',
  fallback_strategy VARCHAR(64) NOT NULL DEFAULT 'fail_soft',
  status VARCHAR(16) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused', 'disabled')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ai_prompt_registry (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  function_type VARCHAR(64) NOT NULL REFERENCES ai_function_configs(function_type),
  prompt_version VARCHAR(64) NOT NULL,
  schema_version VARCHAR(64) NOT NULL,
  status VARCHAR(16) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'deprecated', 'disabled')),
  system_template_ref TEXT NOT NULL,
  developer_template_ref TEXT,
  user_template_ref TEXT NOT NULL,
  output_schema_contract_ref TEXT NOT NULL,
  output_schema_ref TEXT NOT NULL,
  parser_policy VARCHAR(64) NOT NULL DEFAULT 'strict_json',
  eval_suite_id VARCHAR(100),
  created_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (function_type, prompt_version, schema_version)
);

CREATE INDEX IF NOT EXISTS idx_ai_prompt_registry_active
  ON ai_prompt_registry (function_type, status);

CREATE TABLE IF NOT EXISTS ai_call_logs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  request_id UUID NOT NULL DEFAULT uuid_generate_v4(),
  function_type VARCHAR(64) NOT NULL,
  key_alias VARCHAR(100) NOT NULL,
  provider VARCHAR(64) NOT NULL DEFAULT 'deepseek',
  model VARCHAR(100) NOT NULL,
  prompt_version VARCHAR(64) NOT NULL,
  schema_version VARCHAR(64) NOT NULL,
  call_source VARCHAR(64),
  queue_name VARCHAR(64),
  priority INTEGER NOT NULL DEFAULT 0,
  user_id UUID REFERENCES users(id) ON DELETE SET NULL,
  production_batch_id UUID,
  source_material_id UUID,
  candidate_experience_id UUID,
  experience_id UUID REFERENCES experiences(id) ON DELETE SET NULL,
  chat_topic_id UUID REFERENCES chat_topics(id) ON DELETE SET NULL,
  chat_message_id UUID REFERENCES chat_messages(id) ON DELETE SET NULL,
  job_id UUID,
  input_tokens INTEGER,
  output_tokens INTEGER,
  latency_ms INTEGER,
  attempt_no INTEGER NOT NULL DEFAULT 1,
  retry_of_call_id UUID REFERENCES ai_call_logs(id) ON DELETE SET NULL,
  status VARCHAR(32) NOT NULL CHECK (status IN ('success', 'failed', 'timeout', 'invalid_output', 'empty_content', 'budget_blocked')),
  error_code VARCHAR(100),
  sanitized_input_summary TEXT,
  sanitized_output_summary TEXT,
  started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_call_logs_function_time
  ON ai_call_logs (function_type, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ai_call_logs_status_time
  ON ai_call_logs (status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_ai_call_logs_user_time
  ON ai_call_logs (user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS ai_jobs (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  function_type VARCHAR(64) NOT NULL,
  target_type VARCHAR(64) NOT NULL,
  target_id UUID,
  status VARCHAR(32) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'succeeded', 'failed', 'retrying', 'dead_letter')),
  priority INTEGER NOT NULL DEFAULT 0,
  attempt_count INTEGER NOT NULL DEFAULT 0,
  max_attempts INTEGER NOT NULL DEFAULT 3,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  last_error TEXT,
  scheduled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_jobs_status_schedule
  ON ai_jobs (status, scheduled_at, priority DESC);

-- Initial AI function config rows. Prompt registry entries are added with actual prompt files later.
INSERT INTO ai_function_configs (function_type, key_alias, prompt_version, schema_version, timeout_ms, max_tokens, temperature, thinking, queue_name, fallback_strategy)
VALUES
  ('chat', 'deepseek_chat_primary', 'chat_v1', 'chat_schema_v1', 60000, 2048, 0.70, 'disabled', 'user_realtime', 'retry_user_visible'),
  ('chat_summary', 'deepseek_chat_primary', 'chat_summary_v1', 'chat_summary_schema_v1', 30000, 1200, 0.30, 'disabled', 'user_background', 'use_recent_messages'),
  ('chat_topic_classify', 'deepseek_chat_primary', 'chat_topic_classify_v1', 'chat_topic_classify_schema_v1', 20000, 800, 0.20, 'disabled', 'user_normal', 'temporary_title'),
  ('experience_rewrite', 'deepseek_user_primary', 'experience_rewrite_v1', 'experience_rewrite_schema_v1', 30000, 800, 0.40, 'disabled', 'user_normal', 'keep_original'),
  ('moderation', 'deepseek_moderation_primary', 'moderation_v1', 'moderation_schema_v1', 20000, 800, 0.10, 'disabled', 'user_realtime', 'delay_public_distribution'),
  ('experience_extract', 'deepseek_content_primary', 'experience_extract_v1', 'experience_extract_schema_v1', 90000, 4096, 0.30, 'enabled', 'content_low', 'pause_unit'),
  ('experience_review', 'deepseek_content_primary', 'experience_review_v1', 'experience_review_schema_v1', 60000, 2048, 0.20, 'enabled', 'content_low', 'manual_or_retry'),
  ('experience_classify', 'deepseek_content_primary', 'experience_classify_v1', 'experience_classify_schema_v1', 30000, 1000, 0.20, 'disabled', 'content_low', 'pending_classification'),
  ('experience_interpretation', 'deepseek_content_primary', 'experience_interpretation_v1', 'experience_interpretation_schema_v1', 45000, 1800, 0.50, 'disabled', 'content_low', 'show_no_interpretation'),
  ('recommendation_ai', 'deepseek_recommendation_primary', 'recommendation_ai_v1', 'recommendation_ai_schema_v1', 30000, 1500, 0.20, 'disabled', 'user_normal', 'rule_recommendation'),
  ('translation_normalization', 'deepseek_content_primary', 'translation_normalization_v1', 'translation_normalization_schema_v1', 30000, 1200, 0.20, 'disabled', 'content_low', 'keep_original')
ON CONFLICT (function_type) DO NOTHING;

-- Production migrations are normally applied as a privileged database role,
-- while the API connects as the application role. Keep new V4 tables readable
-- and writable by the app role when that role exists.
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'niangao') THEN
    GRANT USAGE ON SCHEMA public TO niangao;
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE
      experience_collections,
      experience_inspirations,
      experience_events,
      experience_metrics,
      recommendation_sessions,
      search_sessions,
      chat_topics,
      chat_temp_sessions,
      chat_messages,
      chat_citations,
      feedback,
      ai_function_configs,
      ai_prompt_registry,
      ai_call_logs,
      ai_jobs
    TO niangao;
    GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO niangao;
  END IF;
END $$;
