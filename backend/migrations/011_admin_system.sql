-- 011_admin_system.sql
-- 管理后台基础设施

-- 1. 管理员标记
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT FALSE;

-- 设置第一个管理员（你的开发者账号）
-- UPDATE users SET is_admin = TRUE WHERE id = '<your-user-id>';

-- 2. 操作日志表
CREATE TABLE IF NOT EXISTS admin_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    admin_id UUID NOT NULL REFERENCES users(id),
    action_type VARCHAR(50) NOT NULL,
    target_type VARCHAR(50),
    target_id UUID,
    detail JSONB,
    result VARCHAR(20) DEFAULT 'success',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_logs_admin ON admin_logs(admin_id);
CREATE INDEX IF NOT EXISTS idx_admin_logs_time ON admin_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_admin_logs_type ON admin_logs(action_type);

-- 3. 系统配置表（持久化配置项）
CREATE TABLE IF NOT EXISTS system_config (
    key VARCHAR(100) PRIMARY KEY,
    value JSONB NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 默认配置
INSERT INTO system_config (key, value) VALUES
    ('review_mode', '"human_review"'),
    ('content_min_length', '10'),
    ('content_max_length', '100'),
    ('interpretation_max_length', '300'),
    ('title_max_length', '15'),
    ('publish_limit_per_day', '20'),
    ('chat_limit_per_day', '50'),
    ('sensitive_words_enabled', 'true'),
    ('registration_enabled', 'true'),
    ('ai_interpretation_enabled', 'true'),
    ('search_enabled', 'true')
ON CONFLICT (key) DO NOTHING;

-- 4. 敏感词表
CREATE TABLE IF NOT EXISTS sensitive_words (
    id SERIAL PRIMARY KEY,
    word VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sensitive_words_word ON sensitive_words(word);
