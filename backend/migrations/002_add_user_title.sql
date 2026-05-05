-- Migration 002: 用户个人称号 (title)
-- 普通用户可自定义称号，显示在经验卡片和个人主页
ALTER TABLE users ADD COLUMN IF NOT EXISTS title VARCHAR(20);
