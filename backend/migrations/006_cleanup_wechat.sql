-- 006: Clean up vestigial WeChat columns
-- WeChat login was removed; these columns are no longer populated

ALTER TABLE users DROP COLUMN IF EXISTS wechat_openid;
ALTER TABLE users DROP COLUMN IF EXISTS wechat_unionid;
