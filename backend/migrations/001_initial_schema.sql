-- 年糕 App — 数据库 Schema
-- Supabase PostgreSQL + pgvector

-- ============================================
-- 扩展
-- ============================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgvector";

-- ============================================
-- 用户资料（扩展 Supabase Auth）
-- ============================================
CREATE TABLE profiles (
  id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  nickname VARCHAR(30) NOT NULL DEFAULT '',
  avatar_url TEXT,
  bio VARCHAR(200),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 自动创建 profile
CREATE OR REPLACE FUNCTION handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO public.profiles (id, nickname)
  VALUES (NEW.id, COALESCE(NEW.raw_user_meta_data->>'nickname', ''));
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION handle_new_user();

-- ============================================
-- 经验
-- ============================================
CREATE TYPE domain_type AS ENUM (
  'career',       -- 职场成长
  'relationship', -- 人际关系
  'cognition',    -- 认知升级
  'life',         -- 生活智慧
  'emotion'       -- 情感
);

CREATE TABLE experiences (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  author_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  content VARCHAR(100) NOT NULL,
  interpretation TEXT CHECK (char_length(interpretation) <= 500),
  domain domain_type NOT NULL,
  -- 冷启内容标记
  is_official BOOLEAN NOT NULL DEFAULT FALSE,
  source_label VARCHAR(100),
  -- 互动统计
  like_count INTEGER NOT NULL DEFAULT 0,
  bookmark_count INTEGER NOT NULL DEFAULT 0,
  -- AI 相关
  embedding VECTOR(1536),
  interpretation_generated BOOLEAN NOT NULL DEFAULT FALSE,
  -- 审核
  status VARCHAR(20) NOT NULL DEFAULT 'published' CHECK (status IN ('published', 'hidden', 'flagged')),
  -- 时间
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_experiences_author ON experiences(author_id);
CREATE INDEX idx_experiences_domain ON experiences(domain);
CREATE INDEX idx_experiences_created ON experiences(created_at DESC);
CREATE INDEX idx_experiences_status ON experiences(status);
CREATE INDEX idx_experiences_embedding ON experiences USING ivfflat (embedding vector_cosine_ops) WITH (lists = 100);

-- ============================================
-- 点赞
-- ============================================
CREATE TABLE likes (
  user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  experience_id UUID NOT NULL REFERENCES experiences(id) ON DELETE CASCADE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (user_id, experience_id)
);

-- 自动更新点赞计数
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
  user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  experience_id UUID NOT NULL REFERENCES experiences(id) ON DELETE CASCADE,
  -- 用户是否标记为"已实践"
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
  user_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
  title VARCHAR(100),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_conversations_user ON conversations(user_id, updated_at DESC);

CREATE TABLE messages (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
  role VARCHAR(10) NOT NULL CHECK (role IN ('user', 'assistant')),
  content TEXT NOT NULL,
  -- AI 消息中引用的经验
  referenced_experience_ids UUID[] DEFAULT '{}',
  -- 本轮检索到的经验（用于后续分析）
  retrieved_experience_ids UUID[] DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_conversation ON messages(conversation_id, created_at);

-- ============================================
-- RLS 策略
-- ============================================
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE experiences ENABLE ROW LEVEL SECURITY;
ALTER TABLE likes ENABLE ROW LEVEL SECURITY;
ALTER TABLE bookmarks ENABLE ROW LEVEL SECURITY;
ALTER TABLE conversations ENABLE ROW LEVEL SECURITY;
ALTER TABLE messages ENABLE ROW LEVEL SECURITY;

-- Profiles: 所有人可读，仅自己可写
CREATE POLICY "profiles_readable" ON profiles FOR SELECT USING (true);
CREATE POLICY "profiles_writable" ON profiles FOR UPDATE USING (auth.uid() = id);

-- Experiences: 已发布的可读，仅作者可写
CREATE POLICY "experiences_readable" ON experiences FOR SELECT USING (status = 'published' OR author_id = auth.uid());
CREATE POLICY "experiences_insertable" ON experiences FOR INSERT WITH CHECK (auth.uid() = author_id);
CREATE POLICY "experiences_updatable" ON experiences FOR UPDATE USING (auth.uid() = author_id);
CREATE POLICY "experiences_deletable" ON experiences FOR DELETE USING (auth.uid() = author_id);

-- Likes: 自己操作
CREATE POLICY "likes_readable" ON likes FOR SELECT USING (true);
CREATE POLICY "likes_manageable" ON likes FOR ALL USING (auth.uid() = user_id);

-- Bookmarks: 自己操作
CREATE POLICY "bookmarks_readable" ON bookmarks FOR SELECT USING (auth.uid() = user_id);
CREATE POLICY "bookmarks_manageable" ON bookmarks FOR ALL USING (auth.uid() = user_id);

-- Conversations & Messages: 仅自己的对话
CREATE POLICY "conversations_manageable" ON conversations FOR ALL USING (auth.uid() = user_id);
CREATE POLICY "messages_readable" ON messages FOR SELECT
  USING (EXISTS (SELECT 1 FROM conversations WHERE id = messages.conversation_id AND user_id = auth.uid()));
CREATE POLICY "messages_insertable" ON messages FOR INSERT
  WITH CHECK (EXISTS (SELECT 1 FROM conversations WHERE id = messages.conversation_id AND user_id = auth.uid()));
