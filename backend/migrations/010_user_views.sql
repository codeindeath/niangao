-- 009: User views tracking table for profile stats
CREATE TABLE IF NOT EXISTS user_views (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    experience_id UUID NOT NULL REFERENCES experiences(id) ON DELETE CASCADE,
    viewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, experience_id)
);

CREATE INDEX idx_user_views_user ON user_views(user_id);
