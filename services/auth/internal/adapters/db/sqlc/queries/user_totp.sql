-- name: GetUserTOTP :one
SELECT * FROM user_totp
WHERE user_id = $1;

-- name: IsUserMFAEnabled :one
SELECT EXISTS (
  SELECT 1 FROM user_totp
  WHERE user_id = $1 AND confirmed_at IS NOT NULL
);

-- name: UpsertPendingUserTOTP :execrows
INSERT INTO user_totp (user_id, secret_encrypted)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE
SET secret_encrypted = EXCLUDED.secret_encrypted,
    last_used_step   = NULL,
    created_at       = now()
WHERE user_totp.confirmed_at IS NULL;

-- name: ConfirmUserTOTP :execrows
UPDATE user_totp
SET confirmed_at   = now(),
    last_used_step = $2
WHERE user_id = $1
  AND confirmed_at IS NULL;

-- name: AdvanceUserTOTPStep :execrows
UPDATE user_totp
SET last_used_step = $2
WHERE user_id = $1
  AND confirmed_at IS NOT NULL
  AND (last_used_step IS NULL OR last_used_step < $2);

-- name: DeleteUserTOTP :exec
DELETE FROM user_totp
WHERE user_id = $1;