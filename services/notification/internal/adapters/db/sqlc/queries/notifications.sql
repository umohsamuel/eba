-- name: InsertNotification :exec
INSERT INTO notifications (
  id, event_id, channel, template, recipient, status
) VALUES (
  $1, $2, $3, $4, $5, $6
)
ON CONFLICT (event_id) DO NOTHING;

-- name: GetNotificationByEventID :one
SELECT * FROM notifications
WHERE event_id = $1;

-- name: MarkNotificationSent :exec
UPDATE notifications
SET status = 'sent',
    attempts = attempts + 1,
    provider_message_id = $2,
    last_error = NULL,
    sent_at = now(),
    updated_at = now()
WHERE id = $1;

-- name: MarkNotificationFailed :exec
UPDATE notifications
SET status = 'failed',
    attempts = attempts + 1,
    last_error = $2,
    updated_at = now()
WHERE id = $1;
