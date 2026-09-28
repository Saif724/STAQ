# Software Requirements and Current Scope

This document records backend behavior currently represented in the repository. It is an implementation snapshot, not a promise that every product idea or feature request is already available.

## Purpose

STAQ provides authenticated users with a backend for defining scheduled tasks, attaching ordered actions, dispatching due work through Redis, and inspecting individual executions and their logs.

## Implemented functional scope

### Accounts and authentication

- Register with full name, email, and password; email verification is required for password-based login.
- Login issues an access token and refresh token. Refresh and logout operate on the refresh token.
- Google OAuth sign-in is available, and authenticated users can authorize a Google account for Gmail sending.
- Password-based user endpoints use bearer JWT access tokens.

### Queues and tasks

- Queue records can be created, listed, fetched, updated, and deleted.
- A task belongs to a user and a queue. It has a name, optional description, status, timeout, and maximum retry count.
- Task endpoints support create, list-all-for-current-user, fetch, update, and archive through `DELETE`.
- Valid task statuses are `Active`, `Paused`, and `Archived`.

### Triggers

- Triggers support `ONCE`, `DAILY`, `WEEKLY`, `MONTHLY`, `YEARLY`, and standard five-field `CRON` schedules.
- Scheduling uses an IANA time-zone name and a required `start_at` timestamp.
- Triggers can be created, listed for a task, fetched, updated, and deleted.

### Actions

- Tasks can have ordered `REMINDER`, `EMAIL`, `HTTP`, and `SHELL` actions.
- Action configuration is stored as PostgreSQL JSONB. Actions support `continue_on_failure`.
- The actual executor limits, including restricted shell execution and reminder behavior, are documented in [architecture.md](architecture.md).

### Execution

- The scheduler publishes due jobs to a Redis Stream; workers process jobs and persist execution status, timing, retry count, errors, and logs.
- Execution states in the schema are `PENDING`, `RUNNING`, `SUCCESS`, `FAILED`, `CANCELLED`, and `TIMED_OUT`. The worker currently creates `RUNNING` executions and completes them as `SUCCESS`, `FAILED`, or `TIMED_OUT`.
- API clients can fetch an execution by ID and its logs. There is no execution-history list endpoint.

## Non-functional properties and current limits

- PostgreSQL is persistent storage; Redis carries job messages and transient OAuth/consumer state.
- Worker concurrency is four per worker process. Retry behavior is action-level and only applies to errors classified as retryable.
- The API has no `/api/v1` prefix. See [api-specification.md](api-specification.md) for exact routes and response shapes.
- Queue endpoints require bearer authentication, but queue records are global and are not scoped to an individual user.
- No role-based authorization, task search/filter/pagination, execution list/statistics, queue metrics, in-app notifications, arbitrary script execution, or workflow-specific actions are implemented.
- The `user_settings` table exists, but settings management is not exposed through the API.
- The frontend is a Next.js scaffold; this backend snapshot does not establish that the product workflows described above have a complete user interface.

## Outside current backend scope

AI or natural-language scheduling, voice control, browser automation, calendar sync, team collaboration, SMS/push notifications, and a plugin system are not implemented in the current backend.
