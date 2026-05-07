-- 014_chat_retention: 对话保留 & 性能优化

-- 高效查询近N天消息（先建，防重复）
CREATE INDEX IF NOT EXISTS idx_msg_conv_time 
  ON messages(conversation_id, created_at DESC);

-- 30天自动清理函数
CREATE OR REPLACE FUNCTION cleanup_old_messages()
RETURNS void AS $$
BEGIN
  DELETE FROM messages 
  WHERE created_at < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;
