-- name: GetUserTokenByID :one
SELECT * FROM user_tokens
WHERE id = $1 LIMIT 1;

-- name: GetUserTokenByHash :one
SELECT * FROM user_tokens
WHERE token_hash = $1 LIMIT 1;

-- name: ListUserTokens :many
SELECT * FROM user_tokens;

-- name: CreateUserToken :one
INSERT INTO user_tokens (
  id, user_id, purpose, token_hash, expires_at
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: RevokeUserToken :one
UPDATE user_tokens
SET used_at = now()
WHERE id = $1
  AND used_at IS NULL
  AND expires_at > now()
RETURNING *;

-- name: DeleteUserToken :one
DELETE FROM user_tokens
WHERE id = $1
RETURNING id;

-- name: DeleteStaleUserTokens :execrows
DELETE FROM user_tokens
WHERE used_at IS NOT NULL
   OR expires_at < now();

-- name: RevokeUserTokensForUser :exec
UPDATE user_tokens
SET used_at = now()
WHERE user_id = $1
  AND purpose = $2
  AND used_at IS NULL;

-- name: ConsumeUserToken :one
UPDATE user_tokens
SET used_at = now()
WHERE token_hash = $1
  AND purpose = $2
  AND used_at IS NULL
  AND expires_at > now()
RETURNING user_id;
