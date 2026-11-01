-- +goose Up
-- +goose StatementBegin
CREATE TABLE refresh_tokens (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 token VARCHAR(512) UNIQUE NOT NULL,
 expires_at TIMESTAMPTZ NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 revoked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS refresh_tokens;
DROP INDEX IF EXISTS idx_refresh_tokens_token;
-- +goose StatementEnd
