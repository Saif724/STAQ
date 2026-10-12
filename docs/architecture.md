# Backend Architecture

This document describes the backend as currently implemented. It does not describe planned frontend or platform capabilities as shipped features.

## Runtime components

The backend is split into three Go programs under `backend/cmd`:

| Process | Entry point | Responsibility |
|---|---|---|
| API | `cmd/api` | HTTP routes, authentication, resource CRUD, PostgreSQL and Redis connections |
| Scheduler | `cmd/scheduler` | Polls due triggers every 30 seconds and publishes eligible task jobs |
| Worker | `cmd/worker` | Consumes jobs, runs task actions, and records execution state and logs |

All processes load the same environment configuration and connect to PostgreSQL and Redis. The HTTP API uses the Go standard-library `net/http` `ServeMux`; there is no Gin server or `/api/v1` route prefix. The API defaults to port `8080`, scheduler health to `8081`, and worker health to `8082`; each can be overridden with `PORT`.

## Dependencies and data ownership

- PostgreSQL is the persistent source of truth for users, queues, tasks, triggers, actions, connections, and execution history.
- Redis is used for the task-job stream, consumer-group state, temporary Google OAuth state, and transient coordination. The task stream is `staq:task-stream`, consumed through group `staq-workers`.
- The API manages records; the scheduler advances due trigger times; workers load current task/action data from PostgreSQL when processing a job.
- Queue records are task metadata/assignment. The current broker has one Redis stream, not a separate stream per database queue.

## Scheduled execution flow

```text
Client
  | HTTP API
  v
PostgreSQL: tasks, triggers, actions
  ^                         |
  |                         | Scheduler polls every 30 seconds
  |                         v
  |                   Redis Stream
  |                         |
  |                         v
  +--- Worker loads task/actions, executes actions in order
                            |
                            v
                 PostgreSQL: executions/logs
```

The scheduler selects due active triggers and their task status under a database transaction, publishes a job containing the task ID, trigger ID, and scheduled time, and advances the trigger schedule. A one-time trigger is deactivated after dispatch. For recurring triggers, an occurrence more than two minutes late is skipped and the next run is calculated; this catch-up rule prevents a long outage from replaying every missed interval. Paused and archived tasks are not dispatched: due recurring occurrences advance without publishing, and due one-time triggers are deactivated without publishing. Paused occurrences are not replayed when a task is resumed.

Workers use a Redis Stream consumer group and run four concurrent consumers per worker process. They periodically renew claimed jobs, write execution heartbeats every ten seconds, and attempt pending-job recovery every two minutes. Jobs idle in the pending list for five minutes can be claimed for recovery. Before creating an execution, the worker reloads the task and skips the job if its current status is paused or archived. A unique database constraint on `(trigger_id, scheduled_at)` prevents duplicate execution records for the same scheduled occurrence.

Actions in a task are loaded in execution order. Each action gets a timeout based on `tasks.timeout_seconds`; retryable action failures use increasing delays capped at thirty seconds and `tasks.max_retries`. Non-retryable errors are not retried. `continue_on_failure` controls whether a later action can proceed after an action failure.

## Action executors

| Type | Implemented behavior |
|---|---|
| `REMINDER` | Validates and records the configured message as an execution result; it does not send an in-app or push notification. |
| `EMAIL` | Sends using the configured STAQ sender or an authorized user's Gmail connection. |
| `HTTP` | Makes a request with configured method, URL, headers, and body. Supported methods are GET, POST, PUT, PATCH, DELETE, and HEAD; the HTTP client timeout is 30 seconds and response capture is capped at 64 KiB. |
| `SHELL` | Executes an allowlisted binary. The current worker registry allows only the system `echo` executable; arbitrary commands/scripts are not enabled. |

There are no separate workflow, wait/delay, webhook, or trigger-another-task action types in the current action registry.

## Authentication and request middleware

The API applies request logging, request IDs, CORS, and panic recovery. Protected endpoints use bearer JWT middleware. Registration/login/email verification and Google sign-in are public; Gmail connection initiation, user profile, tasks, queues, triggers, actions, and executions require authentication. The Gmail OAuth callback is public and uses OAuth state to associate the connection with the initiating user.

## Database and migrations

PostgreSQL schema is defined by the sequential SQL migrations in `backend/migrations`. The schema includes authentication, tasks, scheduling, actions, queue metadata, OAuth accounts, Gmail connections, executions, and logs. There is no `workflows` table. See [database-design.md](database-design.md) and [er-diagram.md](er-diagram.md).

## Current boundaries

- Execution history is listable per authenticated user with task/status filters and bounded pagination; there is no task-list pagination/search/filter API.
- Dashboard metrics cover a rolling 30-day execution window; recent executions/failures are all-time top-five lists, while upcoming runs are the next five future triggers for active tasks.
- Profile updates currently support only full name and IANA timezone. Email/password changes and broader user settings are not exposed.
- Gmail connections can be listed and disconnected by their owner; only metadata is returned, never stored credentials. Gmail OAuth redirects to a fixed path on the configured frontend.
- Queue CRUD remains registered, but queues are global metadata and are not user-selectable in task creation. Queue depth/statistics and worker assignment APIs are not implemented.
- Horizontal worker processes can share the Redis consumer group, but scheduler leader election is not implemented; run one scheduler instance unless duplicate scheduling has been independently controlled.
