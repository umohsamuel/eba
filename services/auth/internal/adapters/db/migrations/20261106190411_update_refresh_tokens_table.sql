-- +goose Up
DELETE FROM refresh_tokens;
ALTER TABLE refresh_tokens RENAME COLUMN token TO token_hash;
ALTER TABLE refresh_tokens ALTER COLUMN token_hash TYPE CHAR(64);
DROP INDEX IF EXISTS idx_refresh_tokens_token;

-- +goose Down
DELETE FROM refresh_tokens; 
ALTER TABLE refresh_tokens ALTER COLUMN token_hash TYPE VARCHAR(512);
ALTER TABLE refresh_tokens RENAME COLUMN token_hash TO token;
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens(token);
