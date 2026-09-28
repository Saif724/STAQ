# Deployment Guide

This guide describes the current backend configuration and Docker Compose setup. Review the security notes before exposing any service publicly.

## Requirements

- Go `1.26.2` for local backend builds (as declared by `backend/go.mod`).
- PostgreSQL and Redis for all three backend processes.
- Docker Engine with the Docker Compose plugin for the bundled stack.
- Node.js/npm only when running the separate Next.js frontend.

## Configuration

The API, scheduler, and worker call the same configuration loader. It reads environment variables and optionally loads a `.env` file from the current working directory.

| Variable | Default | Required/used for |
|---|---|---|
| `APP_ENV` | `development` | Logger environment |
| `PORT` | `8080` | HTTP API port; scheduler/worker health server also reads this variable |
| `FRONTEND_URL` | `http://localhost:5173` | Exact allowed CORS origin for the API |
| `BACKEND_URL` | `http://localhost:8080` | Loaded by configuration, but not currently consumed elsewhere by backend runtime code |
| `DATABASE_URL` | none | Required PostgreSQL URL |
| `REDIS_ADDRESS` | `localhost:6379` | Redis address; Compose overrides it to `redis:6379` |
| `REDIS_PASSWORD` | empty | Redis password |
| `REDIS_DB` | `0` | Redis database number |
| `REDIS_TLS` | `false` | Enable Redis TLS |
| `JWT_SECRET` | none | Required signing secret for JWTs |
| `TOKEN_ENCRYPTION_KEY` | none | Required standard-base64 encoding of exactly 32 key bytes, used to encrypt stored Google tokens |
| `GOOGLE_CLIENT_ID` | empty | Google sign-in and Gmail authorization |
| `GOOGLE_CLIENT_SECRET` | empty | Google sign-in and Gmail authorization |
| `GOOGLE_REDIRECT_URL` | empty | Google sign-in callback URL |
| `GOOGLE_GMAIL_REDIRECT_URL` | empty | Gmail connection callback URL |
| `STAQ_GMAIL_REFRESH_TOKEN` | empty | Platform Gmail sender credentials |
| `STAQ_GMAIL_FROM` | empty | Platform sender address |

`DATABASE_URL`, `JWT_SECRET`, `TOKEN_ENCRYPTION_KEY`, `PORT`, and `APP_ENV` are validated at startup (the last two have defaults). OAuth and email credentials are needed for those integrations to work. Never commit real secrets.

The CORS middleware returns the configured `FRONTEND_URL` as `Access-Control-Allow-Origin`; the browser requires it to exactly match the page's origin. The frontend's `npm run dev` command uses Next.js's default port `3000`, while the backend's default CORS origin uses port `5173`; set `FRONTEND_URL` to the exact origin from which the browser runs. A mismatch prevents browser requests from being accepted by CORS.

Both `GOOGLE_REDIRECT_URL` and `GOOGLE_GMAIL_REDIRECT_URL` are used as OAuth callback URLs and must match the callback URLs registered with Google. In the current local environment, the configured backend URL is remote HTTPS while both OAuth callbacks target local HTTP. That is valid only for a deliberate local OAuth flow; for a remote deployment, point both callbacks at its externally reachable HTTPS routes and register the same URLs with Google. `BACKEND_URL` itself does not rewrite these callback URLs.

Create `backend/.env` for Compose services, for example:

```env
APP_ENV=development
PORT=8080
FRONTEND_URL=http://localhost:3000
BACKEND_URL=http://localhost:8080
JWT_SECRET=replace-with-a-long-random-secret
TOKEN_ENCRYPTION_KEY=replace-with-base64-of-32-random-bytes
GOOGLE_CLIENT_ID=
GOOGLE_CLIENT_SECRET=
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback
GOOGLE_GMAIL_REDIRECT_URL=http://localhost:8080/connections/google/callback
STAQ_GMAIL_REFRESH_TOKEN=
STAQ_GMAIL_FROM=
```

Compose overrides `DATABASE_URL` and `REDIS_ADDRESS` for containers, so the corresponding values in `.env` are ignored by the Compose API, scheduler, and worker. The checked-in Compose file uses local development credentials and must not be used as-is for a public/production deployment.

There is a database-name mismatch in the current Compose configuration: PostgreSQL is initialized with database `staq`, but the migration service and backend service URLs select database `postgres`. Migrations and the application therefore use `postgres`, not the configured `staq` database. Correct the Compose values together before relying on the intended database name; this guide does not change backend deployment files.

## Run the Compose stack

From `backend/`:

```sh
docker compose up --build
```

Compose starts PostgreSQL and Redis, runs the migrations, then starts the API, scheduler, and worker. The API is published at `http://localhost:8080`; PostgreSQL is published on host port `5433` and Redis on `6380`. The compose network uses service DNS names and container ports. Stop the stack with `Ctrl+C`; use `docker compose down` to stop and remove containers while retaining the named PostgreSQL volume.

The Compose migration service downloads `golang-migrate` at startup and applies `backend/migrations`. The backend `Makefile` currently has no targets; documented `make migrate-up` commands are not available.

## Run processes without Compose

Provide PostgreSQL and Redis, set `DATABASE_URL`, and use the correct host/port values. Apply migrations with the `migrate` CLI:

```sh
migrate -path migrations -database "$DATABASE_URL" up
```

Then, from `backend/`, run each process in a separate terminal:

```sh
go run ./cmd/api
go run ./cmd/scheduler
go run ./cmd/worker
```

All processes default to port `8080` for their HTTP listener. Set distinct `PORT` values for the scheduler and worker if you need their simple `/health` endpoints while the API is running. The worker and scheduler are long-running processes and must remain running for scheduled jobs to be dispatched and executed.

## Frontend

The frontend is a separate Next.js app under `frontend/`:

```sh
cd frontend
npm install
npm run dev
```

Next.js serves on `http://localhost:3000` by default. Set backend `FRONTEND_URL` to that origin for local browser access, unless you intentionally start the frontend on a different port.

The frontend package does not currently declare a `VITE_API_URL` setting; configure any API URL in frontend code according to the actual app implementation.

## Health and operations

- API `GET /health` checks both PostgreSQL and Redis and returns `200` when both respond, otherwise `503`.
- Scheduler and worker expose a simpler plain-text `ok` health endpoint on their configured `PORT`.
- The API uses structured application logging. No Prometheus/Grafana metrics endpoint is implemented.
- Back up PostgreSQL using the selected hosting provider's procedures. Redis contains the task stream and consumer state, so do not assume its contents can always be reconstructed without analyzing pending work.

## Production checklist

- Use managed PostgreSQL/Redis or securely operated equivalents and production-specific URLs.
- Replace the Compose database credentials; do not expose development ports unnecessarily.
- Use strong JWT and encryption keys, protect them in a secret manager, and configure TLS where supported.
- Set OAuth callback URLs to registered HTTPS endpoints.
- Restrict network access and review whether global queue management should be administrator-only before public exposure.
- Run one scheduler instance unless scheduler coordination is added and verified.
- Add backups, alerting, log retention, and recovery procedures appropriate to the deployment.
