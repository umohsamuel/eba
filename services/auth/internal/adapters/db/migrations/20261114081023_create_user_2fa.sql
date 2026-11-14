-- +goose Up
CREATE TABLE user_totp (
    user_id          UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    secret_encrypted BYTEA NOT NULL,
    confirmed_at     TIMESTAMPTZ,
    last_used_step   BIGINT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_recovery_codes (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code_hash  CHAR(64) NOT NULL UNIQUE,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_user_recovery_codes_user ON user_recovery_codes (user_id);

ALTER TABLE user_tokens ADD COLUMN attempts INT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE user_tokens DROP COLUMN attempts;
DROP TABLE IF EXISTS user_recovery_codes;
DROP TABLE IF EXISTS user_totp;