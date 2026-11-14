-- name: InsertUserRecoveryCode :exec
INSERT INTO user_recovery_codes (id, user_id, code_hash)
VALUES ($1, $2, $3);

-- name: DeleteUserRecoveryCodes :exec
DELETE FROM user_recovery_codes
WHERE user_id = $1;

-- name: ConsumeUserRecoveryCode :execrows
UPDATE user_recovery_codes
SET used_at = now()
WHERE user_id = $1
  AND code_hash = $2
  AND used_at IS NULL;

-- name: CountUnusedUserRecoveryCodes :one
SELECT count(*) FROM user_recovery_codes
WHERE user_id = $1
  AND used_at IS NULL;

