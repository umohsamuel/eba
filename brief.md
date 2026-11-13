# EBA brief

Paste or reference this file at the start of a session so the assistant has the context.

## Working rules

- Production-grade advice only. No learner framing, no "easy option" shortcuts.
- When I ask a question, answer it. Don't change code unless I ask.
- Only change what I ask for. Report anything else you notice instead of fixing it.
- Re-read the current code before flagging or repeating a bug. No stale reports.
- No prose comments in code or config. Only functional ones (swag annotations, `//go:embed`).
- Never commit unless my message says commit. When I do:
  - one commit per logical change, in dependency order, each one building;
  - format: conventional subject (`feat:`, `fix:`, `refactor:`, `build:`, `chore:`, `docs:`) + short bullet body;
  - no AI attribution of any kind.
- Never commit secrets: `.env`, `*.pem`, `infra/connectors/auth-outbox.json`.
- Keep `pkg/` service-agnostic.

## Repo

Go monorepo with a committed `go.work`. Every module also has `require` + `replace` entries for the local `pkg` modules, so it builds without the workspace (Docker excludes `go.work`).

| Path | What |
|---|---|
| `pkg/database` | Postgres via `database/sql` + pgx; URL built with `net/url`; `SSLMode` defaults to `verify-full` |
| `pkg/errors` | `apperror`: `Kind`, `Wrap` (records the wrap site), `Origin` |
| `pkg/response`, `pkg/response/gin` | response envelope; `HandleError` / `HandleSuccess` / `HandleMessage`; `.Send()` / `.Abort()` |
| `pkg/logger` | slog setup (JSON in production, text otherwise) |
| `pkg/server`, `pkg/shutdown` | HTTP server and graceful shutdown |
| `pkg/authmw` | gin middleware verifying RS256 access tokens with a public key, for services other than auth |
| `pkg/events` | event contracts: `NotificationRequested`, topic, template and data key constants |
| `services/auth` | auth service (below) |
| `services/notification` | notification service (below) |
| `services/api` | gateway: only CORS + aggregated Swagger at `/docs` (`DOCS_SPECS=name=url,...`) so far |
| `infra/` | shared compose: Kafka, Debezium Connect, kafka-ui; owns the `eba_infra` network |
| `services/{cart,inventory,order,payment,product,user}` | not started |

**Layering (hexagonal):** handlers → services → repositories/adapters behind interfaces.
- Repositories wrap driver errors into domain errors.
- Services pass domain errors through.
- Handlers only wrap request-binding errors.

## Auth service

- **Endpoints:** `POST /api/v1/auth/{register,login,refresh,forgot-password,reset-password}`, `GET /.well-known/jwks.json`, `/docs`.
- **Migration endpoints** (`X-Migration-Token`, only registered if `MIGRATION_TOKEN` is set):
  - `GET /api/v1/migrations/status` and `/version`;
  - `POST /api/v1/migrations/up`, `/up-to/{version}` and `/down-to/{version}`.
- **Access token:** RS256 JWT, 15m, with a `kid` header. Key from `JWT_ACCESS_PRIVATE_KEY_FILE`; old public keys for rotation in `JWT_ACCESS_PREVIOUS_PUBLIC_KEY_FILES`. Audience: `JWT_AUDIENCE`, one entry per service.
- **Refresh token:** HS256 JWT, 168h, audience `JWT_REFRESH_AUDIENCE`. Only the SHA-256 hash is stored; `jti` is the row ID.
- **Client types:**
  - web (default) gets HttpOnly cookies only: `access_token` with Path `/`, `refresh_token` with Path `COOKIE_REFRESH_PATH` and SameSite=Strict;
  - `X-Client-Type: mobile` gets tokens in the body only.
- **Emails** are lowercased and trimmed, then validated with `net/mail`. Unique index on `lower(email)`.
- **Login** uses a dummy bcrypt hash for unknown emails, to keep timing uniform. bcrypt cost is 14.
- **Forgot password:**
  - same 200 response whether or not the account exists;
  - one transaction: revoke old reset tokens → insert the token hash (`user_tokens`, purpose `password_reset`) → insert the outbox row → delete the outbox row.
- **Reset password:** single-use token. One transaction: claim the token, update the password, revoke all refresh tokens.
- **Outbox + CDC:**
  - Postgres runs with `wal_level=logical`; publication `auth_outbox` and a dedicated `debezium` role (REPLICATION + SELECT on `outbox`);
  - Debezium reads slot `auth_outbox` and routes rows with the EventRouter: `aggregatetype` → topic, `aggregateid` → key, `payload` → value.
- **Own protected routes:** `middlewares.Auth` (Bearer, then the cookie). The `protected` group in `http.go` is still empty.

## Notification service

- **Consumes** `notifications.requested` (group `notification.main`).
- **Retries:** `.retry.1` → `.retry.2` → `.retry.3` (delays from `KAFKA_RETRY_DELAYS`) → `.dlq`. Permanent errors go straight to the DLQ.
- **Handle:** validate → render the embedded template → `CreateIfAbsent` by `event_id` (skip if already sent) → SMTP via go-mail → mark sent or failed. Email contents are never stored.
- **Templates:** `auth.welcome`, `auth.verify_email`, `auth.password_reset`, `auth.password_changed` (HTML + text, shared layout).
- **Topics:** created at startup, because broker auto-create is off.

## Local dev

| Thing | Where |
|---|---|
| auth | `:8080`, Swagger at `/docs/index.html` |
| Debezium Connect REST | `:8083` |
| kafka-ui | `:8090` |
| Mailpit | SMTP `:1025`, web inbox `:8025` |
| Kafka | `localhost:9092` from the host, `kafka:29092` from containers |
| auth Postgres | `:5432` (`auth_pg`) |
| notification Postgres | `:5433` (`notification_pg`) |

**Start order:**
1. `infra`: `make docker-run`. Auth's `make docker-run` also runs this first.
2. Auth: `make docker-run`, then `make watch`.
3. Notification: `docker compose up -d`, then `make watch`.
4. If Kafka data is new, register the connector: `curl -X POST localhost:8083/connectors -H "Content-Type: application/json" -d @infra/connectors/auth-outbox.json` (copy it from `auth-outbox.example.json` and set the password).

**Gotchas:**
- In Git Bash, prefix `docker exec ... /opt/...` with `MSYS_NO_PATHCONV=1`.
- goose auto-loads `.env`; use `-env=none` when you pass a DSN yourself.
- After GitHub "Rebase and merge", always start new branches from an updated `main`.

## Deferred: things I said I'd implement myself

**Gateway (`services/api`)**
- Reverse proxy `/api/v1/auth/*` to auth with `httputil.ReverseProxy`, passing `Set-Cookie` through.
- Protected routes with `pkg/authmw` (`CookieName: access_token`). Overwrite `X-User-ID` from the verified claims before proxying; strip the client's copy.
- Load `JWT_ACCESS_PUBLIC_KEY_FILE` (`openssl pkey -in keys/access_private.pem -pubout -out keys/access_public.pem`).
- Filter internal routes (e.g. `/migrations`) out of the aggregated docs. Disable docs in production.

**Rate limiting / Redis**
- Rate-limit login and forgot-password per IP and per account, at the gateway or with Redis.
- Redis: a running instance, `pkg/redis` (client constructor only), a per-service adapter behind a port, `github.com/redis/go-redis/v9`, key prefixes per service.
- Consider bcrypt cost 14 → 12.

**Auth features**
- Logout: revoke the refresh token and clear the cookies (`cookieWriter.Clear` exists).
- `/auth/me` on the protected group, plus `@Security BearerAuth` on protected routes.
- Email verification: reuse `user_tokens` with purpose `email_verification`; template already exists.
- Publish welcome and password-changed events through the outbox; templates already exist.
- Google login: an `identities` table (provider + provider subject, unique). Link to an existing account by email only when both sides are verified.
- Refresh token rotation + reuse detection.
- Credential version / `password_changed_at`: closes the login-vs-reset race and makes access tokens revocable.
- Async forgot-password (202 + background job): uniform response and timing even when writes fail.
- Cleanup job for expired/revoked refresh tokens and stale `user_tokens` (`DeleteAllStale`). Built in-app, not pg_cron (`todos.md`).
- Pagination package for list endpoints (`todos.md`); `users.List` ignores its filter today.
- Remove the `user.User` claim from the refresh token, and `pkg/token`'s dependency on the domain package.
- Delete the empty `internal/adapters/adapters.go` stub.

**Notification**
- Claim before send (`sending` status + `claimed_at`) once there are multiple instances or SMS.
- DLQ replay tool and alerting on the DLQ.
- Short `retention.ms` on `notifications.requested*` (reset links live in Kafka).
- More channels (SMS, push) as `EmailSender`-style adapters; user preferences.
- A real provider for production (SES / Postmark / Resend) over SMTP.

**Platform**
- Outbox + Debezium connector for each new service that publishes events.
- gRPC for internal service-to-service calls; REST stays at the edge.
- Asymmetric-key rollout: other services verify via JWKS (`MicahParks/keyfunc`), automatic key rotation.
- Production: Connect REST not public, connector secrets from files, monitor `pg_replication_slots` (drop unused slots), dedicated DB users, TLS to Postgres, Kafka Connect hosted properly (Strimzi / MSK Connect).
- CI: `GOWORK=off go build ./...` per module, tests, lint.
