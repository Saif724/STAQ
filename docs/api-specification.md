# HTTP API Specification

This reference reflects route registrations and handler DTOs in `backend/internal/router` and `backend/internal`. It is not generated from OpenAPI. The backend OpenAPI file should be checked against this contract before use.

## Base URL and conventions

Local API base URL: `http://localhost:8080`. Routes are registered at the root; there is no `/api/v1` prefix.

Protected routes use:

```http
Authorization: Bearer <access_token>
Content-Type: application/json
```

Most JSON success responses wrap a value as `{"data": ...}`. Error responses are `{"error":{"code":"...","message":"..."}}`. `GET /health`, redirects, and `204 No Content` responses do not use those wrappers. Not all endpoints share the same status codes; the tables below list the principal outcomes.

## Route index

| Method | Path | Auth | Behavior |
|---|---|---|---|
| GET | `/health` | No | PostgreSQL and Redis readiness |
| POST | `/auth/register` | No | Register and send verification email |
| POST | `/auth/login` | No | Issue access and refresh tokens |
| POST | `/auth/refresh` | No | Issue a new access token from a refresh token |
| POST | `/auth/logout` | No | Revoke supplied refresh token |
| POST | `/auth/verify-email` | No | Verify email using email and code |
| POST | `/auth/resend-verification` | No | Resend verification email |
| GET | `/auth/google` | No | Redirect to Google sign-in |
| GET | `/auth/google/callback` | No | Complete Google sign-in |
| GET | `/users/me` | Bearer | Current user's profile and saved timezone |
| PUT | `/users/me` | Bearer | Update full name and saved IANA timezone preference |
| GET | `/dashboard/summary` | Bearer | User-scoped execution metrics, recent executions/failures, and upcoming runs |
| GET | `/connections` | Bearer | List current user's connection metadata without credentials |
| DELETE | `/connections/{id}` | Bearer | Disconnect an owned external account |
| GET | `/connections/google/authorization` | Bearer | Return a Google authorization URL for browser navigation |
| GET | `/connections/google` | Bearer | Begin Gmail authorization; redirects to Google |
| GET | `/connections/google/callback` | OAuth state | Complete Gmail connection |
| POST | `/tasks` | Bearer | Create task |
| POST | `/task-setups` | Bearer | Atomically create a task, one trigger, and ordered actions |
| GET | `/tasks` | Bearer | List current user's tasks |
| GET | `/tasks/{id}` | Bearer | Fetch owned task |
| PUT | `/tasks/{id}` | Bearer | Update task name, description, status, timeout, and retries; queue assignment is preserved |
| DELETE | `/tasks/{id}` | Bearer | Archive task (not physical delete) |
| GET | `/queues` | Bearer | List queues |
| GET | `/queues/{id}` | Bearer | Fetch queue |
| POST | `/queues` | Bearer | Create queue |
| PUT | `/queues/{id}` | Bearer | Update queue |
| DELETE | `/queues/{id}` | Bearer | Delete queue |
| POST | `/triggers` | Bearer | Create trigger |
| GET | `/tasks/{taskID}/triggers` | Bearer | List triggers for an owned task |
| GET | `/triggers/{id}` | Bearer | Fetch trigger for an owned task |
| PUT | `/triggers/{id}` | Bearer | Update trigger |
| DELETE | `/triggers/{id}` | Bearer | Delete trigger; returns no content |
| POST | `/actions` | Bearer | Create action |
| GET | `/tasks/{taskID}/actions` | Bearer | List actions for an owned task |
| GET | `/actions/{id}` | Bearer | Fetch action for an owned task |
| PUT | `/actions/{id}` | Bearer | Update action |
| DELETE | `/actions/{id}` | Bearer | Delete action; returns no content |
| GET | `/executions` | Bearer | List owned executions with optional task/status filters and pagination |
| GET | `/executions/{id}` | Bearer | Fetch execution for a task owned by caller |
| GET | `/executions/{id}/logs` | Bearer | Fetch logs for an execution owned by caller |

There is no task search/filter/pagination route. Task creation assigns the authenticated user's active default queue; clients cannot select a queue.

## Health

`GET /health` returns `200` if both PostgreSQL and Redis respond, or `503` if either check fails.

```json
{"status":"ok","database":"ok","redis":"ok"}
```

## Authentication

### Register

`POST /auth/register`

```json
{"full_name":"Ada Lovelace","email":"ada@example.com","password":"correct-horse"}
```

Full name is 2–100 characters, email must parse as an address and be at most 255 characters, and password is 8–120 characters. Success is `201` with `data` containing `id`, `full_name`, `email`, and `email_verified`. Duplicate verified email returns `409`; validation/send failures return `400`.

### Login

`POST /auth/login` with `{"email":"ada@example.com","password":"correct-horse"}`. A verified active account receives `200` and `data` with `access_token` and `refresh_token`. Invalid credentials return `401`; inactive or unverified accounts return `403`.

### Refresh and logout

Both accept `{"refresh_token":"..."}` at `POST /auth/refresh` and `POST /auth/logout`. Refresh returns `data.access_token`; logout returns `data` containing the string `logged out successfully`. Refresh/logout endpoints do not require an access JWT because the refresh token is the credential.

### Verify and resend

`POST /auth/verify-email` accepts `{"email":"ada@example.com","code":"..."}`. Success returns `data.message`. `POST /auth/resend-verification` accepts `{"email":"ada@example.com"}` and returns `data.message`. Verification failures are represented by error codes such as `INVALID_VERIFICATION_CODE` and `VERIFICATION_EXPIRED`.

### Google sign-in

`GET /auth/google` sets an OAuth state cookie and redirects to Google. Google returns to `GET /auth/google/callback`; the callback validates the cookie and Redis state before completing sign-in, then redirects to the configured frontend with a one-time login code. Errors also redirect using fixed error codes. Configure Google's redirect URL to the backend callback. This is a browser redirect flow, not an email/password JSON login request.

## User profile

`GET /users/me` requires bearer auth and returns `data` with `id`, `full_name`, `email`, `email_verified`, and the saved IANA `timezone`. `PUT /users/me` accepts `{"full_name":"Ada Lovelace","timezone":"Asia/Dhaka"}`; full name must be 2–100 characters and timezone must be a valid IANA name. Email and password are not editable through this endpoint.

## Gmail connection

`GET /connections/google/authorization` requires bearer auth and returns `data.authorization_url`; the browser then navigates to that URL. `GET /connections/google` is also available for server-side redirect clients. The callback `GET /connections/google/callback` receives OAuth `state` and `code`; it uses the stored state to associate the resulting Google account with the initiating user. It redirects only to the configured frontend's fixed `/dashboard/connections` path with a fixed result code; it does not include OAuth codes, tokens, account email, or internal error details. Tokens are encrypted before persistence. `GET /connections` returns caller-owned metadata only (ID, provider, account email, scopes, expiry, and timestamps). `DELETE /connections/{id}` deletes only a connection owned by the caller and returns `204`.

## Queues

Queue endpoints require bearer auth. Queue create body: `{"name":"default","description":"Optional"}`. Update body: `{"name":"default","description":"Optional","is_active":true}`. Names are required and at most 100 characters; names are unique. Create returns `201`; list/get/update return `200`; delete returns `204`. Queue records are global, rather than owned by an individual user.

## Tasks

Create body:

```json
{
  "name":"Daily check",
  "description":"Optional",
  "timeout_seconds":300,
  "max_retries":2
}
```

The authenticated user ID is taken from the JWT, not the body. The backend assigns the user's active default queue; clients do not supply a queue ID. A non-empty name (maximum 150 characters) is required. Non-positive create timeout defaults to 300; max retries cannot be negative. New tasks start as `Active`.

For the complete automation wizard, prefer `POST /task-setups`. It accepts the task fields, trigger fields (`trigger_type`, optional `cron_expression`, `timezone`, `start_at`), and an `actions` array containing each action's type, order, JSON configuration, and continue-on-failure flag. User and queue assignment are derived by the backend. It validates the full setup before writing and creates the task, trigger, and up to 50 actions in one database transaction; a failed write rolls back the complete setup. Success is `201` with `data.task`, `data.trigger`, and `data.actions`.

Update body requires `name`, `status`, `timeout_seconds`, and `max_retries` (description may be null). Queue assignment is preserved and is not accepted in the update body:

```json
{
  "name":"Daily check",
  "description":null,
  "status":"Paused",
  "timeout_seconds":300,
  "max_retries":2
}
```

Allowed status strings are exactly `Active`, `Paused`, and `Archived`. `DELETE /tasks/{id}` archives the task and returns `204`. List returns all tasks belonging to the caller; it accepts no paging/filter parameters.

Pausing a task prevents the scheduler from publishing occurrences while the task is paused. Due recurring occurrences are advanced without dispatch, rather than replayed on resume; a due one-time trigger is deactivated without dispatch. The worker checks current task status before starting a queued job, so jobs already in Redis are skipped if the task is still paused or archived when the worker checks it.

Task success bodies use `data` containing the task object (list uses an array). Fields include `id`, `user_id`, `queue_id`, `name`, `description`, `status`, `timeout_seconds`, `max_retries`, `created_at`, and `updated_at`.

## Triggers

Create body:

```json
{
  "task_id":"<task-uuid>",
  "trigger_type":"CRON",
  "cron_expression":"0 9 * * 1-5",
  "timezone":"America/New_York",
  "start_at":"2026-10-01T09:00:00-04:00"
}
```

`start_at` and a valid IANA timezone are required. Types are `ONCE`, `DAILY`, `WEEKLY`, `MONTHLY`, `YEARLY`, and `CRON`; cron uses standard five-field syntax and is required only for `CRON`. A one-time start must be in the future. Update accepts `trigger_type`, `cron_expression`, `timezone`, `start_at`, and `is_active`. Create/get/update/list return trigger data; delete returns `204`.

The scheduler polls every 30 seconds. A recurring occurrence more than two minutes overdue is skipped and rescheduled to its next occurrence. Paused-task occurrences are also skipped and advanced, regardless of lateness.

## Actions

Create body:

```json
{
  "task_id":"<task-uuid>",
  "action_type":"HTTP",
  "execution_order":1,
  "configuration":{"method":"POST","url":"https://example.com/hook","headers":{"Content-Type":"application/json"},"body":"{}"},
  "continue_on_failure":false
}
```

Update uses the same fields except `task_id`. Types are `REMINDER`, `EMAIL`, `HTTP`, and `SHELL`; order must be at least 1 and unique within the task; configuration must be valid JSON. Configuration is stored as JSONB. **Serialization note:** the response model currently declares configuration as `[]byte`, so Go JSON encoding represents this field as base64 text rather than as the original JSON object.

Configuration shapes accepted by the current executors:

| Type | Configuration fields |
|---|---|
| `REMINDER` | `{"message":"..."}` |
| `EMAIL` | `{"sender":"STAQ" or "USER_GMAIL","connection_id":"...","to":"...","subject":"...","body":"..."}`; `connection_id` is required for `USER_GMAIL` |
| `HTTP` | `{"method":"GET|POST|PUT|PATCH|DELETE|HEAD","url":"...","headers":{},"body":"..."}`; method defaults to GET |
| `SHELL` | `{"command":"...","args":[]}`; current worker only allows the `echo` executable |

## Executions

`GET /executions` returns `data` with `items`, `total`, `page`, and `page_size`. Optional query parameters are `task_id` and `status`; status is one of `PENDING`, `RUNNING`, `SUCCESS`, `FAILED`, `CANCELLED`, or `TIMED_OUT`. `page` defaults to `1`; `page_size` defaults to `20` and must be between `1` and `100`. Results are ordered by `created_at DESC, id DESC`. Both items and total count are scoped to the authenticated user and the selected filters.

`GET /executions/{id}` returns one execution in `data`, including task/trigger IDs, scheduled/start/completion times, status, heartbeat, duration, retry count, and optional error. `GET /executions/{id}/logs` returns log entries with `INFO`, `WARNING`, or `ERROR` level. All three endpoints enforce task ownership.

## Dashboard summary

`GET /dashboard/summary` returns counts over the rolling 30 days beginning at `window_starts_at`. `success_rate` is `successful_count / (successful_count + failed_count + timed_out_count) * 100`; pending, running, and cancelled executions are excluded from this denominator. A zero denominator returns `0`. `recent_executions` and `recent_failures` contain up to five latest rows across all time, ordered by `created_at DESC, id DESC`; failures include `FAILED` and `TIMED_OUT`. `upcoming_executions` contains up to five future active triggers whose tasks are currently `Active`, ordered by next run time. The response is read from one repeatable-read database snapshot and contains only the authenticated user's data.

## Error format

```json
{"error":{"code":"TASK_NOT_FOUND","message":"task not found"}}
```

Error code and HTTP status depend on the handler; malformed JSON generally returns `400`, missing/invalid auth returns `401`, missing resources commonly return `404`, and unexpected failures may return `500`. The implementation does not use one uniform status mapping across every resource.
