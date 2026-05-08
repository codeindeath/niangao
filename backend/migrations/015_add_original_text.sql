-- 015_add_original_text.sql
-- 新增 original_text 字段，存储英文原文（中文翻译后展示在 content）

ALTER TABLE experiences ADD COLUMN IF NOT EXISTS original_text TEXT;

COMMENT ON COLUMN experiences.original_text IS '原始语言原文（如英文），content 为对应的中文翻译';
