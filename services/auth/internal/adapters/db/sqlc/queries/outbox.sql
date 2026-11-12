-- name: InsertOutboxEvent :exec
INSERT INTO outbox (id, aggregatetype, aggregateid, type, payload)
VALUES ($1, $2, $3, $4, $5);

-- name: DeleteOutboxEvent :exec
DELETE FROM outbox WHERE id = $1;