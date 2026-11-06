piority:

- setup swagger docs

1. in auth servce have a cron job that clears the expired/revoked refersh tokens at intervals

```sql
-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_cron;

SELECT cron.schedule(
    'purge-expired-tokens',
    '0 0 * * *', -- Every day at midnight
    $$ DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked = TRUE; $$
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT cron.unschedule('purge-expired-tokens');
-- +goose StatementEnd
```

```sql
-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_refresh_tokens_cleanup ON refresh_tokens (expires_at, revoked);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_refresh_tokens_cleanup;
-- +goose StatementEnd
```

dont use pg cron like in the example, just build soeothing out

2. i should have some sort of pagination pkg and use it for the lists like list users and etc
