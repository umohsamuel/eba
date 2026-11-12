-- +goose Up
UPDATE users SET email = lower(email) WHERE email <> lower(email);

CREATE UNIQUE INDEX users_email_lower_key ON users (lower(email));

-- +goose Down
DROP INDEX IF EXISTS users_email_lower_key;
