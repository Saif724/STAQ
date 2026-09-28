# Database Design

This reference follows the SQL migrations in `backend/migrations`; migrations are authoritative if this document and schema ever differ.

## Storage

PostgreSQL stores persistent accounts, tasks, schedule definitions, action configuration, connections, execution history, and logs. Redis is used for the task-job stream and group, pending-job delivery/recovery, and temporary OAuth state. Redis is not a relational schema and does not provide the caching, rate-limiting, or session-storage features proposed in older drafts.

The API, scheduler, and worker connect to PostgreSQL through a `pgx` pool and to Redis through `go-redis`. The scheduler uses a transaction while selecting/updating a due trigger and publishing its job.

## PostgreSQL tables

The current schema contains twelve tables.

| Table | Purpose |
|---|---|
| `users` | Local user account and password hash |
| `email_verifications` | Verification token lifecycle |
| `refresh_tokens` | Hashed, expiring, revocable refresh tokens |
| `oauth_accounts` | External identity mapping for Google sign-in |
| `connections` | Google/Gmail account metadata and encrypted OAuth credentials |
| `user_settings` | Time zone and optional default queue settings |
| `queues` | Named task queue metadata |
| `tasks` | User-owned tasks and execution limits |
| `triggers` | One-time, recurrence, or cron schedule |
| `actions` | Ordered task actions and JSONB configuration |
| `executions` | A scheduled execution occurrence and current outcome |
| `execution_logs` | Execution event/result messages |

There is no `workflows` table. Actions belong directly to tasks. `user_settings` is present in migrations but has no current API/service route.

## Table definitions

Types below are the migration types; timestamps use PostgreSQL `TIMESTAMP` (without explicit `WITH TIME ZONE`). Unless called out, identifiers are UUID primary keys and timestamps do not have SQL defaults.

### `users`

| Column | Type | Nullability / constraint |
|---|---|---|
| `id` | UUID | Primary key |
| `full_name` | VARCHAR(100) | NOT NULL |
| `email` | VARCHAR(255) | NOT NULL, UNIQUE |
| `password_hash` | TEXT | NOT NULL |
| `email_verified` | BOOLEAN | DEFAULT FALSE |
| `is_active` | BOOLEAN | DEFAULT TRUE |
| `created_at`, `updated_at` | TIMESTAMP | NOT NULL |

Indexes: `idx_users_email`, `idx_users_created_at`.

### `email_verifications`

`id UUID` primary key; `user_id UUID NOT NULL` references `users(id)` with `ON DELETE CASCADE`; `token TEXT NOT NULL UNIQUE`; `expires_at TIMESTAMP NOT NULL`; `verified_at TIMESTAMP NULL`; `created_at TIMESTAMP NOT NULL`. Migration 000011 adds nullable `revoked_at TIMESTAMP`. Indexes cover token, user ID, and expiry.

### `refresh_tokens`

`id UUID` primary key; `user_id UUID NOT NULL` references users with `ON DELETE CASCADE`; `token_hash TEXT NOT NULL UNIQUE`; `expires_at TIMESTAMP NOT NULL`; `revoked_at TIMESTAMP NULL`; `created_at TIMESTAMP NOT NULL`; `last_used_at TIMESTAMP NULL`. Indexes cover user ID and expiry. Raw refresh tokens are not stored in this table.

### `oauth_accounts`

`id UUID` primary key; `user_id UUID NOT NULL` references users with `ON DELETE CASCADE`; `provider VARCHAR(255) NOT NULL`; `provider_user_id VARCHAR(255) NOT NULL`; `created_at`, `updated_at TIMESTAMP NOT NULL`. `(provider, provider_user_id)` is unique. Indexes cover user ID and provider.

### `connections`

`id UUID` primary key; `user_id UUID NOT NULL` references users with `ON DELETE CASCADE`; `provider VARCHAR(50) NOT NULL`; `provider_account_id VARCHAR(255) NOT NULL`; `account_email VARCHAR(255) NOT NULL`; `access_token_encrypted TEXT NOT NULL`; `refresh_token_encrypted TEXT NULL`; `token_expires_at TIMESTAMP NULL`; `scopes TEXT[] NOT NULL DEFAULT '{}'`; `created_at`, `updated_at TIMESTAMP NOT NULL`. `(provider, provider_account_id)` is unique. Indexes cover user ID and provider.

### `user_settings`

`id UUID` primary key; `user_id UUID NOT NULL` references users; `timezone VARCHAR(50) NOT NULL`; `default_queue_id UUID NULL` references queues; `created_at`, `updated_at TIMESTAMP NOT NULL`. There is an index on `user_id`, but no uniqueness constraint or delete cascade on either foreign key.

### `queues`

`id UUID` primary key; `name VARCHAR(100) NOT NULL UNIQUE`; `description TEXT NULL`; `is_active BOOLEAN DEFAULT TRUE`; `created_at TIMESTAMP NOT NULL`. Index: `idx_queues_name`.

### `tasks`

`id UUID` primary key; `user_id UUID NOT NULL` references users; `queue_id UUID NOT NULL` references queues; `name VARCHAR(150) NOT NULL`; `description TEXT NULL`; `status VARCHAR(20) NOT NULL`; `timeout_seconds INTEGER DEFAULT 300`; `max_retries INTEGER DEFAULT 3`; `created_at`, `updated_at TIMESTAMP NOT NULL`. Status check: `Active`, `Paused`, or `Archived`. Indexes: user, queue, status, and created time. Service behavior defaults non-positive create timeouts to 300; `max_retries` must be non-negative. Updates require positive timeout.

### `triggers`

`id UUID` primary key; `task_id UUID NOT NULL` references tasks; `trigger_type VARCHAR(30) NOT NULL`; `cron_expression VARCHAR(100) NULL`; `timezone VARCHAR(50) NOT NULL`; `next_run_at TIMESTAMP NOT NULL`; `last_run_at TIMESTAMP NULL`; `is_active BOOLEAN DEFAULT TRUE`; `created_at`, `updated_at TIMESTAMP NOT NULL`. Migration 000015 adds required `start_at TIMESTAMP`. Trigger types: `ONCE`, `DAILY`, `WEEKLY`, `MONTHLY`, `YEARLY`, `CRON`. Indexes: task, next run, active, and start time.

### `actions`

`id UUID` primary key; `task_id UUID NOT NULL` references tasks; `action_type VARCHAR(30) NOT NULL`; `execution_order INTEGER NOT NULL`; `configuration JSONB NOT NULL`; `continue_on_failure BOOLEAN DEFAULT FALSE`; `created_at`, `updated_at TIMESTAMP NOT NULL`. `(task_id, execution_order)` is unique. Types: `REMINDER`, `EMAIL`, `HTTP`, `SHELL`. Indexes: task and execution order.

### `executions`

`id UUID` primary key; `task_id UUID NOT NULL` references tasks; `trigger_id UUID NOT NULL` references triggers; `status VARCHAR(30) NOT NULL`; `started_at TIMESTAMP NOT NULL`; `completed_at TIMESTAMP NULL`; `duration_ms BIGINT NULL`; `retry_count INTEGER DEFAULT 0`; `error_message TEXT NULL`; `created_at TIMESTAMP NOT NULL`. Migration 000014 adds required `scheduled_at TIMESTAMP` and unique `(trigger_id, scheduled_at)`. Migration 000016 adds nullable `heartbeat_at TIMESTAMP`. Status check: `PENDING`, `RUNNING`, `SUCCESS`, `FAILED`, `CANCELLED`, `TIMED_OUT`. Indexes: task, status, start time, trigger, and a partial heartbeat index for running executions.

### `execution_logs`

`id UUID` primary key; `execution_id UUID NOT NULL` references executions; `log_level VARCHAR(20) NOT NULL`; `message TEXT NOT NULL`; `created_at TIMESTAMP NOT NULL`. Levels: `INFO`, `WARNING`, `ERROR`. Indexes: execution, level, and created time. The migration does not specify cascade deletion.

## Redis job representation

The stream `staq:task-stream` uses consumer group `staq-workers`. Each `job` field contains JSON with `id`, `task_id`, `trigger_id`, `scheduled_at`, and `created_at`. Workers acknowledge completed/ignored messages, renew active deliveries, and periodically claim stale pending messages. This stream is distinct from the SQL `queues` table; the current job payload does not select a Redis stream by queue ID.

## Migration history

The repository has 16 ordered up/down migration pairs:

1. Users
2. Email verifications
3. Queues
4. User settings
5. Tasks
6. Triggers
7. Actions
8. Executions
9. Execution logs
10. Refresh tokens
11. Verification revocation timestamp
12. OAuth accounts
13. Google connections
14. Scheduled execution timestamp and uniqueness
15. Trigger start time
16. Execution heartbeat timestamp and partial index

Apply schema changes through migrations. See [deployment.md](deployment.md) for the Compose and CLI commands.
