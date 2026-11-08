-- +goose Up
CREATE TABLE user_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    purpose     VARCHAR(32) NOT NULL,
    token_hash  CHAR(64) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_tokens_user_purpose ON user_tokens (user_id, purpose);

-- +goose Down
DROP TABLE IF EXISTS user_tokens;
DROP INDEX IF EXISTS idx_user_tokens_user_purpose;
