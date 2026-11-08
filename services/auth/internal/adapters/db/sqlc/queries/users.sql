-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * FROM users
ORDER BY email;

-- name: CreateUser :one
INSERT INTO users (
  email, password_hash
) VALUES (
  $1, $2
)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
  set email = $2,
  password_hash = $3
WHERE id = $1
RETURNING *;

-- name: DeleteUser :one
DELETE FROM users
WHERE id = $1
RETURNING id;

-- name: UpdateUserPassword :execrows
UPDATE users
SET password_hash = $2
WHERE id = $1;
