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
| GET | `/users/me` | Bearer | Current user's basic profile |
| GET | `/connections/google` | Bearer | Begin Gmail authorization; redirects to Google |
| GET | `/connections/google/callback` | OAuth state | Complete Gmail connection |
| POST | `/tasks` | Bearer | Create task |
| GET | `/tasks` | Bearer | List current user's tasks |
| GET | `/tasks/{id}` | Bearer | Fetch owned task |
| PUT | `/tasks/{id}` | Bearer | Replace task fields |
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
| GET | `/executions/{id}` | Bearer | Fetch execution for a task owned by caller |
| GET | `/executions/{id}/logs` | Bearer | Fetch logs for an execution owned by caller |

There is no task search/filter/pagination route, no execution-list route, and no route to manage user settings.

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

`GET /auth/google` sets an OAuth state cookie and redirects to Google. Google returns to `GET /auth/google/callback`; the callback validates the cookie and Redis state before completing sign-in. Configure the Google redirect URL to this callback. The callback is a browser redirect flow, not an email/password JSON login request.

## User profile

`GET /users/me` requires bearer auth and returns `data` with `id`, `full_name`, `email`, and `email_verified`.

## Gmail connection

`GET /connections/google` requires bearer auth and redirects to Google. The callback `GET /connections/google/callback` receives OAuth `state` and `code`; it uses the stored state to associate the resulting Google account with the initiating user. Tokens are encrypted before persistence. The API does not expose a connection-list or disconnect route.

## Queues

Queue endpoints require bearer auth. Queue create body: `{"name":"default","description":"Optional"}`. Update body: `{"name":"default","description":"Optional","is_active":true}`. Names are required and at most 100 characters; names are unique. Create returns `201`; list/get/update return `200`; delete returns `204`. Queue records are global, rather than owned by an individual user.

## Tasks

Create body:

```json
{
  "queue_id":"<queue-uuid>",
  "name":"Daily check",
  "description":"Optional",
  "timeout_seconds":300,
  "max_retries":2
}
```

The authenticated user ID is taken from the JWT, not the body. `queue_id` is required; the backend has no default-queue lookup. The supplied queue must exist and be active. A non-empty name (maximum 150 characters) is required. Non-positive create timeout defaults to 300; max retries cannot be negative. New tasks start as `Active`.

Update body requires `queue_id`, `name`, `status`, `timeout_seconds`, and `max_retries` (description may be null):

```json
{
  "queue_id":"<queue-uuid>",
  "name":"Daily check",
  "description":null,
  "status":"Paused",
  "timeout_seconds":300,
  "max_retries":2
}
```

Allowed status strings are exactly `Active`, `Paused`, and `Archived`. `DELETE /tasks/{id}` archives the task and returns `204`. List returns all tasks belonging to the caller; it accepts no paging/filter parameters.

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

`GET /executions/{id}` returns one execution in `data`, including task/trigger IDs, scheduled/start/completion times, status, heartbeat, duration, retry count, and optional error. `GET /executions/{id}/logs` returns log entries with `INFO`, `WARNING`, or `ERROR` level. Both enforce task ownership. There is no endpoint to enumerate a user's executions.

## Error format

```json
{"error":{"code":"TASK_NOT_FOUND","message":"task not found"}}
```

Error code and HTTP status depend on the handler; malformed JSON generally returns `400`, missing/invalid auth returns `401`, missing resources commonly return `404`, and unexpected failures may return `500`. The implementation does not use one uniform status mapping across every resource.
