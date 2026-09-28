# Entity Relationship Diagram

This diagram is derived from the SQL migrations in `backend/migrations`. Redis streams and consumer groups are runtime structures, not relational entities.

## Entities

The current schema has twelve tables: `users`, `email_verifications`, `refresh_tokens`, `oauth_accounts`, `connections`, `user_settings`, `queues`, `tasks`, `triggers`, `actions`, `executions`, and `execution_logs`. There is no `workflows` table.

## Relationships

```text
users 1 ── N email_verifications
users 1 ── N refresh_tokens
users 1 ── N oauth_accounts
users 1 ── N connections
users 1 ── N user_settings (no unique constraint on user_settings.user_id)
queues 1 ── N tasks
queues 1 ── N user_settings (optional default_queue_id)
users 1 ── N tasks
tasks 1 ── N triggers
tasks 1 ── N actions
tasks 1 ── N executions
triggers 1 ── N executions
executions 1 ── N execution_logs
```

The user-settings relationship is not enforced as one-to-one: the migration adds an index on `user_id` but no unique constraint. Queue names and `(task_id, execution_order)` for actions are unique. A unique `(trigger_id, scheduled_at)` constraint prevents duplicate execution records for a scheduled trigger occurrence.

## Referential actions and notes

- Deleting a user cascades to email verification records, refresh tokens, OAuth accounts, and connections.
- The migrations do not declare `ON DELETE CASCADE` for task children, task owners, task queues, or execution logs. Those relationships must not be documented as cascading deletes.
- Task `DELETE` API behavior is archive/status update, not a physical delete.
- Action configuration is PostgreSQL JSONB.
- OAuth accounts (Google sign-in identity) and connections (authorized Google/Gmail account and encrypted credentials) are separate records.

See [database-design.md](database-design.md) for columns, constraints, indexes, and migration history.
