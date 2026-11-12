-- +goose Up
CREATE TABLE outbox (
  id            UUID PRIMARY KEY,
  aggregatetype VARCHAR(255) NOT NULL,
  aggregateid   VARCHAR(255) NOT NULL,
  type          VARCHAR(255) NOT NULL,
  payload       JSONB NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE PUBLICATION auth_outbox FOR TABLE outbox;

-- +goose Down
DROP PUBLICATION IF EXISTS auth_outbox;
DROP TABLE IF EXISTS outbox;
