# STAQ

STAQ (Smart Task Automation Queue) is a scheduled-task backend written in Go. It provides an HTTP API, a PostgreSQL-backed task and execution model, a Redis Stream for scheduled jobs, and separate scheduler and worker processes.

## Current capabilities

- Email/password registration, email verification, JWT access and refresh tokens, and Google sign-in.
- Google Gmail connection for sending email actions.
- Task, queue, trigger, and action CRUD endpoints.
- One-time, daily, weekly, monthly, yearly, and standard cron triggers with IANA time zones.
- Reminder, email, HTTP, and restricted shell actions.
- Execution records, heartbeats, retry counts, and execution logs.
- Docker Compose services for PostgreSQL, Redis, migrations, API, scheduler, and worker.

The checked-in frontend is a Next.js application scaffold; the backend docs describe implemented backend behavior, not a complete end-user product. Current limitations and unimplemented features are called out in the docs.

## Tech stack

- Backend: Go 1.26.2, `net/http`, PostgreSQL, Redis Streams
- Frontend: Next.js, React, TypeScript, Tailwind CSS
- Migrations: `golang-migrate` migration files under `backend/migrations`

## Run locally

1. Create `backend/.env` with at least `JWT_SECRET` and a base64 `TOKEN_ENCRYPTION_KEY`. See [docs/deployment.md](docs/deployment.md) for all settings.
2. From `backend/`, start the local services and all backend processes:

   ```sh
   docker compose up --build
   ```

   Compose runs migrations before starting the API, scheduler, and worker. The API listens on `http://localhost:8080`.

## Documentation

| Document | Contents |
|---|---|
| [docs/SRS.md](docs/SRS.md) | Implemented scope, requirements, and current limitations |
| [docs/architecture.md](docs/architecture.md) | Backend processes and execution flow |
| [docs/database-design.md](docs/database-design.md) | Migration-derived PostgreSQL and Redis design |
| [docs/er-diagram.md](docs/er-diagram.md) | Current database relationships |
| [docs/api-specification.md](docs/api-specification.md) | Registered HTTP routes and payload contracts |
| [docs/deployment.md](docs/deployment.md) | Configuration, local run, and Compose deployment |
| [docs/development-roadmap.md](docs/development-roadmap.md) | Implemented work and outstanding backend work |

## License

MIT
