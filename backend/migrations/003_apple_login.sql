-- ============================================
-- 年糕 App — 数据库迁移 003
-- 支持 Apple Sign In
-- ============================================

-- 1. wechat_openid 改为 nullable（Apple 用户没有它）
ALTER TABLE users ALTER COLUMN wechat_openid DROP NOT NULL;

-- 2. 添加 apple_user_id
ALTER TABLE users ADD COLUMN apple_user_id VARCHAR(255);

-- 3. 唯一约束
ALTER TABLE users ADD CONSTRAINT users_apple_user_id_unique UNIQUE (apple_user_id);

-- 4. 索引
CREATE INDEX idx_users_apple_id ON users(apple_user_id);

-- 5. 约束：必须有登录方式
ALTER TABLE users ADD CONSTRAINT users_login_method_check
  CHECK (apple_user_id IS NOT NULL);
