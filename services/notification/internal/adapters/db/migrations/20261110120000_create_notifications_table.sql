-- +goose Up
CREATE TABLE notifications (
    id                  UUID PRIMARY KEY,
    event_id            VARCHAR(64) NOT NULL UNIQUE,
    channel             VARCHAR(16) NOT NULL,
    template            VARCHAR(64) NOT NULL,
    recipient           VARCHAR(255) NOT NULL,
    status              VARCHAR(16) NOT NULL,
    attempts            INT NOT NULL DEFAULT 0,
    last_error          TEXT,
    provider_message_id VARCHAR(255),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    sent_at             TIMESTAMPTZ
);

CREATE INDEX idx_notifications_status ON notifications (status);

-- +goose Down
DROP TABLE IF EXISTS notifications;
