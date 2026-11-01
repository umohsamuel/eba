-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE id = $1 LIMIT 1;

-- name: ListRefreshTokens :many
SELECT * FROM refresh_tokens;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
  user_id, token
) VALUES (
  $1, $2
)
RETURNING *;

-- name: UpdateRefreshToken :one
UPDATE refresh_tokens
SET revoked = $2
WHERE id = $1
RETURNING *;

-- name: DeleteRefreshToken :exec
DELETE FROM refresh_tokens
WHERE id = $1;