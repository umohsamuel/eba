-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE id = $1 LIMIT 1;

-- name: ListRefreshTokens :many
SELECT * FROM refresh_tokens;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
  id, user_id, token_hash, expires_at
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET revoked = $2
WHERE id = $1
RETURNING *;

-- name: DeleteRefreshToken :one
DELETE FROM refresh_tokens
WHERE id = $1
RETURNING id;

-- name: RevokeRefreshTokensForUser :exec
UPDATE refresh_tokens
SET revoked = TRUE
WHERE user_id = $1
  AND revoked = FALSE;
