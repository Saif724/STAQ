# Development Roadmap

This roadmap is aligned to the current backend source. “Implemented” means an implementation exists in the repository; it does not imply comprehensive test coverage or production readiness.

## Current implementation status

| Area | Status | Current scope |
|---|---|---|
| Go backend structure and configuration | Implemented | API, scheduler, and worker binaries; shared environment configuration |
| PostgreSQL and migrations | Implemented | Sixteen up/down migration pairs define the current schema |
| Redis broker | Implemented | One task Redis Stream and consumer group |
| Authentication | Implemented | Password registration/login, email verification, refresh/logout, Google sign-in |
| Gmail connection | Implemented | Google authorization, encrypted token persistence, Gmail send action support |
| Queue and task APIs | Implemented | Queue CRUD; task CRUD with archive behavior |
| Trigger scheduling | Implemented | One-time, calendar recurrence, and standard cron with time zones |
| Action execution | Implemented with limits | Reminder, email, HTTP, and `echo`-only shell executor |
| Execution tracking | Implemented | Individual execution lookup, logs, heartbeats, retry count |
| API contract documentation | In progress | OpenAPI and the hand-maintained API specification should remain synchronized |
| Frontend product workflows | Not established by backend | The checked-in frontend is a Next.js scaffold |
| Automated tests | Partial | Tests exist for selected validation, middleware, and response behavior; the full runtime flow is not covered |

## Recommended next work

1. **Verify access-control boundaries:** add authorization tests for every resource route and decide whether global queue management should be administrator-only.
2. **Make API documentation executable:** synchronize `backend/openapi.yaml` with the registered routes, DTOs, response envelopes, and actual status codes; consider generating API docs from the source contract.
3. **Expand end-to-end coverage:** test PostgreSQL migrations, scheduler dispatch, Redis recovery, worker retries/timeouts, OAuth state handling, and resource ownership.
4. **Complete operational visibility:** decide whether to add execution listing, queue metrics, and explicit scheduler coordination/leader election.
5. **Define product behavior before expanding executors:** specify notification delivery, shell security, webhook support, and any additional workflow steps before implementing them.
6. **Build and verify user workflows:** implement the frontend against the current API and test registration through scheduled execution.

## Future ideas (not implemented)

Potential later work includes natural-language scheduling, voice input, calendar integrations, team collaboration, additional notification channels, and plugins. These are proposals, not current release capabilities.
