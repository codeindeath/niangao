-- ============================================
-- 年糕 App — 数据库 Schema v2
-- 适配：火山引擎 RDS PostgreSQL + 微信登录
-- ============================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgvector";

-- ============================================
-- 用户表（微信登录）
-- ============================================
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  wechat_openid VARCHAR(128) UNIQUE NOT NULL,
  wechat_unionid VARCHAR(128),
  nickname VARCHAR(30) NOT NULL DEFAULT '',
  avatar_url TEXT,
  bio VARCHAR(200),
  -- 统计缓存
  experience_count INTEGER NOT NULL DEFAULT 0,
  bookmark_count INTEGER NOT NULL DEFAULT 0,
  practiced_count INTEGER NOT NULL DEFAULT 0,
  -- 时间
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_openid ON users(wechat_openid);

-- ============================================
-- 经验
-- ============================================
CREATE TYPE domain_type AS ENUM (
  'career',
  'relationship',
  'cognition',
  'life',
  'emotion'
);

CREATE TABLE experiences (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  content VARCHAR(100) NOT NULL,
  interpretation TEXT CHECK (char_length(interpretation) <= 500),
  domain domain_type NOT NULL,
  is_official BOOLEAN NOT NULL DEFAULT FALSE,
  source_label VARCHAR(100),
  like_count INTEGER NOT NULL DEFAULT 0,
  bookmark_count INTEGER NOT NULL DEFAULT 0,
  embedding VECTOR(1536),
  interpretation_generated BOOLEAN NOT NULL DEFAULT FALSE,
  status VARCHAR(20) NOT NULL DEFAULT 'published' CHECK (status IN ('published', 'hidden', 'flagged')),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_exp_author ON experiences(author_id);
CREATE INDEX idx_exp_domain ON experiences(domain);
CREATE INDEX idx_exp_created ON experiences(created_at DESC);
CREATE INDEX idx_exp_status ON experiences(status);
CREATE INDEX idx_exp_embedding ON experiences USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- ============================================
-- 点赞
-- ============================================
CREATE TABLE likes (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  experience_id UUID NOT NULL REFERENCES experiences(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, experience_id)
);

CREATE OR REPLACE FUNCTION update_like_count()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE experiences SET like_count = like_count + 1 WHERE id = NEW.experience_id;
  ELSIF TG_OP = 'DELETE' THEN
    UPDATE experiences SET like_count = like_count - 1 WHERE id = OLD.experience_id;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER on_like_change
  AFTER INSERT OR DELETE ON likes
  FOR EACH ROW EXECUTE FUNCTION update_like_count();

-- ============================================
-- 收藏
-- ============================================
CREATE TABLE bookmarks (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  experience_id UUID NOT NULL REFERENCES experiences(id) ON DELETE CASCADE,
  practiced BOOLEAN NOT NULL DEFAULT FALSE,
  practiced_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, experience_id)
);

CREATE OR REPLACE FUNCTION update_bookmark_count()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE experiences SET bookmark_count = bookmark_count + 1 WHERE id = NEW.experience_id;
  ELSIF TG_OP = 'DELETE' THEN
    UPDATE experiences SET bookmark_count = bookmark_count - 1 WHERE id = OLD.experience_id;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER on_bookmark_change
  AFTER INSERT OR DELETE ON bookmarks
  FOR EACH ROW EXECUTE FUNCTION update_bookmark_count();

-- ============================================
-- AI 对话
-- ============================================
CREATE TABLE conversations (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title VARCHAR(100),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conv_user ON conversations(user_id, updated_at DESC);

CREATE TABLE messages (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  role VARCHAR(10) NOT NULL CHECK (role IN ('user', 'assistant')),
  content TEXT NOT NULL,
  referenced_experience_ids UUID[] DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_msg_conv ON messages(conversation_id, created_at);

-- ============================================
-- 用户统计触发器
-- ============================================
CREATE OR REPLACE FUNCTION update_experience_count()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE users SET experience_count = experience_count + 1 WHERE id = NEW.author_id;
  ELSIF TG_OP = 'DELETE' THEN
    UPDATE users SET experience_count = experience_count - 1 WHERE id = OLD.author_id;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER on_experience_change
  AFTER INSERT OR DELETE ON experiences
  FOR EACH ROW EXECUTE FUNCTION update_experience_count();

CREATE OR REPLACE FUNCTION update_user_bookmark_count()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'INSERT' THEN
    UPDATE users SET bookmark_count = bookmark_count + 1 WHERE id = NEW.user_id;
  ELSIF TG_OP = 'DELETE' THEN
    UPDATE users SET bookmark_count = bookmark_count - 1 WHERE id = OLD.user_id;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER on_user_bookmark_change
  AFTER INSERT OR DELETE ON bookmarks
  FOR EACH ROW EXECUTE FUNCTION update_user_bookmark_count();

CREATE OR REPLACE FUNCTION update_practiced_count()
RETURNS TRIGGER AS $$
BEGIN
  IF TG_OP = 'UPDATE' THEN
    IF NEW.practiced = TRUE AND OLD.practiced = FALSE THEN
      UPDATE users SET practiced_count = practiced_count + 1 WHERE id = NEW.user_id;
    ELSIF NEW.practiced = FALSE AND OLD.practiced = TRUE THEN
      UPDATE users SET practiced_count = practiced_count - 1 WHERE id = NEW.user_id;
    END IF;
  ELSIF TG_OP = 'INSERT' AND NEW.practiced = TRUE THEN
    UPDATE users SET practiced_count = practiced_count + 1 WHERE id = NEW.user_id;
  END IF;
  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER on_practiced_change
  AFTER INSERT OR UPDATE OF practiced ON bookmarks
  FOR EACH ROW EXECUTE FUNCTION update_practiced_count();
