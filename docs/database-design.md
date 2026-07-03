# Database Design

## STAQ
### Intelligent Task Scheduling & Workflow Automation Platform

Version: 1.0

Prepared By:
Ahsan Ahmed Saif

---

# 1. Introduction

## 1.1 Purpose

This document describes the database design of STAQ. It defines how application data is stored, organized, and managed using PostgreSQL and Redis.

The database is designed to ensure data integrity, scalability, reliability, and efficient query performance while supporting the scheduling and execution of automated tasks.

This document serves as the primary reference for implementing the database schema, migrations, indexes, and relationships.

---

## 1.2 Database Technologies

STAQ uses two different storage systems.

### PostgreSQL

PostgreSQL is the primary relational database used to store persistent application data.

It stores:

- User accounts
- Tasks
- Triggers
- Workflows
- Actions
- Executions
- Logs
- Queues
- Settings

PostgreSQL guarantees data consistency through ACID transactions and relational constraints.

---

### Redis

Redis is used as an in-memory data store for high-speed operations.

Redis is responsible for:

- Background job queues
- Worker communication
- Temporary locks
- Caching
- Rate limiting (future)
- Session storage (future)

Redis data is temporary and can be recreated from PostgreSQL when necessary.

---

# 2. Design Goals

The STAQ database is designed with the following goals:

- Maintain data integrity using relational constraints.
- Support concurrent task execution.
- Minimize redundant data.
- Optimize frequently executed queries.
- Separate persistent data from temporary queue data.
- Allow future expansion without major schema changes.
- Support horizontal worker scaling.

---

# 3. Database Overview

The system consists of two storage layers.

```

                    +----------------+
                    |   PostgreSQL   |
                    +----------------+
                            |
      ----------------------------------------------
      |        |        |        |        |         |
    Users    Tasks   Triggers Workflows Executions Logs
                            |
                            |
                    +----------------+
                    |     Redis      |
                    +----------------+
                            |
      ----------------------------------------------
      |              |               |
    Job Queue     Worker Queue     Cache

```

PostgreSQL stores permanent business data, while Redis handles temporary execution data and asynchronous communication.

---

# 4. Design Principles

The database follows several important design principles.

## Normalization

The schema follows relational normalization to minimize duplicate data and maintain consistency.

---

## Referential Integrity

Relationships between tables are enforced using foreign key constraints.

---

## Scalability

The schema supports future features such as multiple workers, distributed queues, plugins, and AI-assisted workflows without requiring major redesign.

---

## Performance

Indexes are added to frequently queried columns to improve read performance.

---

## Reliability

Critical operations use database transactions to ensure consistency even if failures occur during execution.

---

## Security

Sensitive information such as passwords and SMTP credentials are never stored in plain text.

---

# 5. Database Components

The STAQ database is divided into the following logical components.

| Component | Purpose |
|------------|---------|
| Authentication | User accounts and verification |
| Task Management | Tasks and scheduling |
| Workflow Engine | Actions and workflows |
| Execution Engine | Execution history and logs |
| Queue System | Worker queues |
| Configuration | User and system settings |

---

# 6. Summary

The STAQ database combines PostgreSQL for persistent relational data and Redis for high-performance asynchronous processing.

This hybrid architecture provides reliability, scalability, and performance while keeping the system modular and easy to maintain.

The following chapters describe the complete schema, relationships, constraints, indexes, and migration strategy.

---

# 2. PostgreSQL Schema

## 2.1 Overview

The PostgreSQL database stores all persistent application data.

The schema is organized into several logical domains to keep related data together and simplify maintenance.

The main domains are:

- Authentication
- Task Management
- Workflow Management
- Execution Tracking
- Queue Management
- User Preferences

---

## 2.2 Tables

Version 1 of STAQ contains the following tables.

| Table | Purpose |
|--------|---------|
| users | Stores user accounts and authentication information. |
| email_verifications | Stores email verification tokens. |
| tasks | Stores user-created tasks. |
| triggers | Stores task scheduling information. |
| workflows | Groups actions belonging to a task. |
| actions | Stores workflow actions. |
| executions | Stores every task execution. |
| execution_logs | Stores logs generated during execution. |
| queues | Stores logical execution queues. |
| user_settings | Stores user preferences and settings. |

---

## 2.3 Authentication Domain

Authentication is responsible for managing user accounts.

Tables:

- users
- email_verifications

Responsibilities:

- User registration
- User login
- Password management
- Email verification

---

## 2.4 Task Management Domain

Task Management stores scheduled tasks created by users.

Tables:

- tasks
- triggers

Responsibilities:

- Create tasks
- Schedule tasks
- Pause tasks
- Resume tasks
- Archive tasks

---

## 2.5 Workflow Domain

Every task owns exactly one workflow.

A workflow contains one or more actions that execute sequentially.

Tables:

- workflows
- actions

Responsibilities:

- Define automation steps
- Store action configurations
- Maintain execution order

---

## 2.6 Execution Domain

Every task execution is recorded for monitoring and debugging.

Tables:

- executions
- execution_logs

Responsibilities:

- Store execution history
- Store execution status
- Store execution duration
- Store error messages
- Store execution logs

---

## 2.7 Queue Domain

The Queue domain organizes task execution.

Table:

- queues

Responsibilities:

- Logical queue separation
- Worker assignment
- Queue monitoring

---

## 2.8 User Settings Domain

Stores user-specific configuration.

Table:

- user_settings

Examples:

- Preferred timezone
- Default queue
- Notification preferences

---

## 2.9 Schema Overview

The database schema is intentionally modular.

Each table has a single responsibility and communicates with other tables through foreign key relationships.

This design improves maintainability, scalability, and future extensibility.

The detailed structure of each table is described in the following chapters.

---

# 3. Redis Design

## 3.1 Overview

Redis serves as the high-performance, in-memory component of STAQ.

Unlike PostgreSQL, Redis does not store permanent business data. Instead, it is used for fast communication between the Scheduler, Broker, and Worker Pool.

If Redis is restarted, no critical business data is lost because the authoritative data remains in PostgreSQL.

---

## 3.2 Responsibilities

Redis is responsible for:

- Job queues
- Worker communication
- Temporary locks
- Execution retries
- Dead Letter Queue (DLQ)

Future versions may also use Redis for:

- Session storage
- Response caching
- Rate limiting

---

## 3.3 Job Queue

When the Scheduler determines that a task should execute, the Job Service creates a Job Payload.

The Broker publishes the Job Payload to a Redis queue.

Workers continuously listen to these queues and process jobs as they become available.

Example flow:

Scheduler

↓

Job Service

↓

Broker

↓

Redis Queue

↓

Worker

---

## 3.4 Dead Letter Queue (DLQ)

If a job fails repeatedly and reaches its maximum retry count, it is moved to the Dead Letter Queue.

The DLQ stores failed jobs for debugging and possible manual reprocessing.

Each failed job contains:

- Execution ID
- Task ID
- Failure reason
- Retry count
- Failed timestamp

---

## 3.5 Temporary Locks

Redis is used to prevent duplicate execution of the same scheduled task.

When a Worker begins processing a job, a temporary execution lock is created.

The lock is automatically removed after execution completes or expires.

This helps ensure that the same task is not executed multiple times concurrently.

---

## 3.6 Future Redis Usage

Redis has been integrated in a way that allows future expansion without changing the architecture.

Possible future additions include:

- User session storage
- API response caching
- Distributed rate limiting
- Distributed scheduler coordination

These features are outside the scope of Version 1.

---

## 3.7 Summary

Redis provides fast, temporary storage for asynchronous processing while PostgreSQL remains the source of truth for persistent application data.

Separating these responsibilities keeps the system reliable, scalable, and efficient.

---

# 4. Table Specifications

This chapter defines every table used by STAQ.

For each table, the following information is provided:

- Purpose
- Columns
- Constraints
- Relationships
- Indexes

---

# 4.1 users

## Purpose

The `users` table stores user account information and authentication details.

Each user can own multiple tasks but has only one account.

---

## Columns

| Column | Type | Constraints | Description |
|---------|------|-------------|-------------|
| id | UUID | Primary Key | Unique identifier for the user |
| full_name | VARCHAR(100) | NOT NULL | User's full name |
| email | VARCHAR(255) | UNIQUE, NOT NULL | User email address |
| password_hash | TEXT | NOT NULL | Encrypted password |
| email_verified | BOOLEAN | DEFAULT FALSE | Email verification status |
| is_active | BOOLEAN | DEFAULT TRUE | Account status |
| created_at | TIMESTAMP | NOT NULL | Account creation time |
| updated_at | TIMESTAMP | NOT NULL | Last update time |

---

## Relationships

One User

↓

Many Tasks

One User

↓

One User Settings

One User

↓

Many Email Verification Records

---

## Constraints

Primary Key

- id

Unique Constraints

- email

Not Null

- full_name
- email
- password_hash
- created_at
- updated_at

---

## Indexes

| Index | Purpose |
|--------|----------|
| idx_users_email | Fast login lookup |
| idx_users_created_at | Sorting users |

---

# 4.2 email_verifications

## Purpose

Stores email verification tokens generated during user registration.

Verification records expire automatically after a predefined period.

---

## Columns

| Column | Type | Constraints | Description |
|---------|------|-------------|-------------|
| id | UUID | Primary Key | Verification record ID |
| user_id | UUID | Foreign Key | Associated user |
| token | VARCHAR(255) | UNIQUE, NOT NULL | Verification token |
| expires_at | TIMESTAMP | NOT NULL | Expiration time |
| verified_at | TIMESTAMP | NULL | Verification completion time |
| created_at | TIMESTAMP | NOT NULL | Record creation time |

---

## Relationships

One User

↓

Many Email Verification Records

---

## Constraints

Primary Key

- id

Foreign Key

- user_id → users(id)

Unique

- token

---

## Indexes

| Index | Purpose |
|--------|----------|
| idx_email_verifications_token | Token lookup |
| idx_email_verifications_user | User lookup |
| idx_email_verifications_expires | Cleanup expired tokens |

---

# 4.3 user_settings

## Purpose

Stores user-specific preferences and application settings.

Each user has exactly one settings record.

---

## Columns

| Column | Type | Constraints | Description |
|---------|------|-------------|-------------|
| id | UUID | Primary Key | Unique settings identifier |
| user_id | UUID | Foreign Key, UNIQUE | Owner of the settings |
| timezone | VARCHAR(50) | NOT NULL | User's preferred timezone |
| default_queue_id | UUID | Foreign Key, NULL | Default execution queue |
| created_at | TIMESTAMP | NOT NULL | Record creation time |
| updated_at | TIMESTAMP | NOT NULL | Last update time |

---

## Relationships

One User

↓

One User Settings

One Queue

↓

Many User Settings

---

## Constraints

Primary Key

- id

Foreign Keys

- user_id → users(id)
- default_queue_id → queues(id)

Unique

- user_id

---

## Indexes

| Index | Purpose |
|--------|----------|
| idx_user_settings_user | Fast lookup by user |

---

# 4.4 queues

## Purpose

Stores logical execution queues.

Queues allow tasks to be separated into different execution groups.

Examples include:

- Default
- High Priority
- Low Priority

---

## Columns

| Column | Type | Constraints | Description |
|---------|------|-------------|-------------|
| id | UUID | Primary Key | Queue identifier |
| name | VARCHAR(100) | UNIQUE, NOT NULL | Queue name |
| description | TEXT | NULL | Queue description |
| is_active | BOOLEAN | DEFAULT TRUE | Queue status |
| created_at | TIMESTAMP | NOT NULL | Creation timestamp |

---

## Relationships

One Queue

↓

Many Tasks

One Queue

↓

Many User Settings

---

## Constraints

Primary Key

- id

Unique

- name

---

## Indexes

| Index | Purpose |
|--------|----------|
| idx_queues_name | Queue lookup |

---

# 4.5 tasks

## Purpose

Stores all tasks created by users.

A task represents an automation that will execute according to one or more triggers.

---

## Columns

| Column | Type | Constraints | Description |
|---------|------|-------------|-------------|
| id | UUID | Primary Key | Task identifier |
| user_id | UUID | Foreign Key | Task owner |
| queue_id | UUID | Foreign Key | Execution queue |
| name | VARCHAR(150) | NOT NULL | Task name |
| description | TEXT | NULL | Optional description |
| status | VARCHAR(20) | NOT NULL | Active, Paused, Archived |
| timeout_seconds | INTEGER | DEFAULT 300 | Maximum execution time |
| max_retries | INTEGER | DEFAULT 3 | Maximum retry attempts |
| created_at | TIMESTAMP | NOT NULL | Creation time |
| updated_at | TIMESTAMP | NOT NULL | Last update time |

---

## Relationships

One User

↓

Many Tasks

One Queue

↓

Many Tasks

One Task

↓

Many Triggers

One Task

↓

Many Actions

One Task

↓

Many Executions

---

## Constraints

Primary Key

- id

Foreign Keys

- user_id → users(id)
- queue_id → queues(id)

Check Constraints

status IN

- Active
- Paused
- Archived

---

## Indexes

| Index | Purpose |
|--------|----------|
| idx_tasks_user | User lookup |
| idx_tasks_queue | Queue lookup |
| idx_tasks_status | Status filtering |
| idx_tasks_created | Sort by creation date |

---

# 4.6 triggers

## Purpose

The `triggers` table defines when a task should be executed.

A task may have one or more triggers, allowing the same task to run according to multiple schedules.

Examples:

- Every day at 8:00 AM
- Every Monday at 9:00 AM
- Every month on the 1st
- Every year on January 1
- Custom Cron Expression

---

## Columns

| Column | Type | Constraints | Description |
|---------|------|-------------|-------------|
| id | UUID | Primary Key | Trigger identifier |
| task_id | UUID | Foreign Key | Associated task |
| trigger_type | VARCHAR(30) | NOT NULL | Type of trigger |
| cron_expression | VARCHAR(100) | NULL | Cron schedule |
| timezone | VARCHAR(50) | NOT NULL | Timezone used for scheduling |
| next_run_at | TIMESTAMP | NOT NULL | Next scheduled execution |
| last_run_at | TIMESTAMP | NULL | Last successful execution |
| is_active | BOOLEAN | DEFAULT TRUE | Trigger status |
| created_at | TIMESTAMP | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |

---

## Relationships

One Task

↓

Many Triggers

---

## Constraints

Primary Key

- id

Foreign Key

- task_id → tasks(id)

Check Constraints

trigger_type IN

- ONCE
- DAILY
- WEEKLY
- MONTHLY
- YEARLY
- CRON

---

## Indexes

| Index | Purpose |
|--------|----------|
| idx_triggers_task | Find triggers for a task |
| idx_triggers_next_run | Scheduler lookup |
| idx_triggers_active | Active trigger lookup |

---

# 4.7 actions

## Purpose

The `actions` table stores every action belonging to a task.

Actions are executed sequentially according to their execution order.

Examples include:

- Reminder
- Email
- HTTP Request
- Shell Command

---

## Columns

| Column | Type | Constraints | Description |
|---------|------|-------------|-------------|
| id | UUID | Primary Key | Action identifier |
| task_id | UUID | Foreign Key | Associated task |
| action_type | VARCHAR(30) | NOT NULL | Action type |
| execution_order | INTEGER | NOT NULL | Position in workflow |
| configuration | JSONB | NOT NULL | Action configuration |
| continue_on_failure | BOOLEAN | DEFAULT FALSE | Continue workflow if action fails (reserved for future use) |
| created_at | TIMESTAMP | NOT NULL | Creation timestamp |
| updated_at | TIMESTAMP | NOT NULL | Last update timestamp |

---

## Relationships

One Task

↓

Many Actions

---

## Constraints

Primary Key

- id

Foreign Key

- task_id → tasks(id)

Unique Constraint

(task_id, execution_order)

Check Constraints

action_type IN

- REMINDER
- EMAIL
- HTTP
- SHELL

---

## Indexes

| Index | Purpose |
|--------|----------|
| idx_actions_task | Lookup task actions |
| idx_actions_order | Execute in order |

---

# 4.8 executions

## Purpose

The `executions` table records every execution attempt performed by the Worker Pool.

Each execution represents one attempt to execute a task.

Execution history enables monitoring, debugging, retry handling, and performance analysis.

---

## Columns

| Column | Type | Constraints | Description |
|---------|------|-------------|-------------|
| id | UUID | Primary Key | Execution identifier |
| task_id | UUID | Foreign Key | Executed task |
| trigger_id | UUID | Foreign Key | Trigger that started execution |
| status | VARCHAR(30) | NOT NULL | Execution status |
| started_at | TIMESTAMP | NOT NULL | Start time |
| completed_at | TIMESTAMP | NULL | Completion time |
| duration_ms | BIGINT | NULL | Execution duration |
| retry_count | INTEGER | DEFAULT 0 | Retry number |
| error_message | TEXT | NULL | Failure reason |
| created_at | TIMESTAMP | NOT NULL | Record creation |

---

## Relationships

One Task

↓

Many Executions

One Trigger

↓

Many Executions

One Execution

↓

Many Execution Logs

---

## Constraints

Primary Key

- id

Foreign Keys

- task_id → tasks(id)
- trigger_id → triggers(id)

Check Constraints

status IN

- PENDING
- RUNNING
- SUCCESS
- FAILED
- CANCELLED
- TIMED_OUT

---

## Indexes

| Index | Purpose |
|--------|----------|
| idx_executions_task | Task history |
| idx_executions_status | Filter by status |
| idx_executions_started | Recent executions |
| idx_executions_trigger | Trigger history |

---

# 4.9 execution_logs

## Purpose

Stores detailed logs generated during task execution.

Logs help developers understand exactly what occurred during execution.

These logs are especially useful for debugging failed workflows.

---

## Columns

| Column | Type | Constraints | Description |
|---------|------|-------------|-------------|
| id | UUID | Primary Key | Log identifier |
| execution_id | UUID | Foreign Key | Associated execution |
| log_level | VARCHAR(20) | NOT NULL | Log severity |
| message | TEXT | NOT NULL | Log message |
| created_at | TIMESTAMP | NOT NULL | Log timestamp |

---

## Relationships

One Execution

↓

Many Execution Logs

---

## Constraints

Primary Key

- id

Foreign Key

- execution_id → executions(id)

Check Constraints

log_level IN

- INFO
- WARNING
- ERROR

---

## Indexes

| Index | Purpose |
|--------|----------|
| idx_logs_execution | Execution lookup |
| idx_logs_level | Filter by log level |
| idx_logs_created | Sort logs |

---

# 4.10 Table Summary

The following tables make up the Version 1 PostgreSQL schema.

| Table | Purpose |
|--------|---------|
| users | User accounts |
| email_verifications | Email verification records |
| user_settings | User preferences |
| queues | Logical execution queues |
| tasks | User-created automation tasks |
| triggers | Scheduling rules |
| actions | Workflow actions |
| executions | Execution history |
| execution_logs | Execution logs |

The schema is normalized to reduce redundancy while maintaining clear relationships between entities.

Each table has a single responsibility, making the database easier to maintain and extend in future versions.

---

# 5. Relationships & Constraints

This chapter defines the relationships between tables and the constraints used to maintain data integrity.

---

# 5.1 Entity Relationships

The following relationships exist between the primary entities.

## User → Tasks

One user can create multiple tasks.

One task belongs to exactly one user.

Relationship:

One-to-Many (1:N)

---

## User → User Settings

Each user has exactly one settings record.

Each settings record belongs to exactly one user.

Relationship:

One-to-One (1:1)

---

## User → Email Verifications

A user may receive multiple email verification tokens.

Each verification token belongs to one user.

Relationship:

One-to-Many (1:N)

---

## Queue → Tasks

One queue may contain many tasks.

Each task belongs to one queue.

Relationship:

One-to-Many (1:N)

---

## Queue → User Settings

A queue may be selected as the default queue by multiple users.

Each user may have one default queue.

Relationship:

One-to-Many (1:N)

---

## Task → Triggers

A task may have multiple triggers.

Each trigger belongs to one task.

Relationship:

One-to-Many (1:N)

---

## Task → Actions

A task may contain multiple actions.

Each action belongs to one task.

Relationship:

One-to-Many (1:N)

---

## Task → Executions

A task may execute many times.

Each execution belongs to one task.

Relationship:

One-to-Many (1:N)

---

## Trigger → Executions

One trigger may generate multiple executions.

Each execution is associated with the trigger that initiated it.

Relationship:

One-to-Many (1:N)

---

## Execution → Execution Logs

Each execution may produce multiple log entries.

Each log entry belongs to exactly one execution.

Relationship:

One-to-Many (1:N)

---

# 5.2 Foreign Keys

The following foreign key relationships enforce referential integrity.

| Child Table | Foreign Key | Parent Table |
|--------------|-------------|--------------|
| email_verifications | user_id | users |
| user_settings | user_id | users |
| user_settings | default_queue_id | queues |
| tasks | user_id | users |
| tasks | queue_id | queues |
| triggers | task_id | tasks |
| actions | task_id | tasks |
| executions | task_id | tasks |
| executions | trigger_id | triggers |
| execution_logs | execution_id | executions |

---

# 5.3 Primary Keys

Every table uses a UUID as its primary key.

| Table | Primary Key |
|---------|-------------|
| users | id |
| email_verifications | id |
| user_settings | id |
| queues | id |
| tasks | id |
| triggers | id |
| actions | id |
| executions | id |
| execution_logs | id |

Using UUIDs allows globally unique identifiers and simplifies future horizontal scaling.

---

# 5.4 Unique Constraints

The following fields must contain unique values.

| Table | Column |
|---------|---------|
| users | email |
| email_verifications | token |
| queues | name |
| user_settings | user_id |
| actions | (task_id, execution_order) |

These constraints prevent duplicate data and maintain consistency.

---

# 5.5 Check Constraints

Check constraints restrict columns to valid values.

## Task Status

Allowed values:

- ACTIVE
- PAUSED
- ARCHIVED

---

## Trigger Type

Allowed values:

- ONCE
- DAILY
- WEEKLY
- MONTHLY
- YEARLY
- CRON

---

## Action Type

Allowed values:

- REMINDER
- EMAIL
- HTTP
- SHELL

---

## Execution Status

Allowed values:

- PENDING
- RUNNING
- SUCCESS
- FAILED
- CANCELLED
- TIMED_OUT

---

## Log Level

Allowed values:

- INFO
- WARNING
- ERROR

---

# 5.6 Cascading Rules

The following cascading rules are recommended.

## User

Deleting a user deletes:

- Tasks
- User Settings
- Email Verification Records

---

## Task

Deleting a task deletes:

- Triggers
- Actions
- Executions

---

## Execution

Deleting an execution deletes:

- Execution Logs

---

Queue records should not be deleted if tasks reference them.

Instead, queues should be marked as inactive.

---

# 5.7 Data Integrity Rules

The database enforces the following integrity rules.

- Every task must belong to a valid user.
- Every trigger must belong to a valid task.
- Every action must belong to a valid task.
- Every execution must reference an existing task.
- Every execution log must reference an existing execution.
- User email addresses must be unique.
- Action execution order must be unique within each task.
- Verification tokens must be unique.
- Queue names must be unique.

---

# 5.8 Summary

Foreign keys, unique constraints, primary keys, and check constraints ensure that invalid or inconsistent data cannot be inserted into the database.

These rules provide a strong foundation for maintaining reliability and consistency throughout the STAQ platform.

---

# 6. Index Strategy

Indexes improve query performance by reducing the amount of data scanned during searches.

The following indexes are recommended for Version 1.

---

## 6.1 Users

| Index | Purpose |
|--------|---------|
| idx_users_email | User login |
| idx_users_created_at | Sort users by creation date |

---

## 6.2 Email Verifications

| Index | Purpose |
|--------|---------|
| idx_email_verifications_token | Token lookup |
| idx_email_verifications_user | User lookup |
| idx_email_verifications_expires | Cleanup expired tokens |

---

## 6.3 User Settings

| Index | Purpose |
|--------|---------|
| idx_user_settings_user | Fast user settings lookup |

---

## 6.4 Queues

| Index | Purpose |
|--------|---------|
| idx_queues_name | Queue lookup |

---

## 6.5 Tasks

| Index | Purpose |
|--------|---------|
| idx_tasks_user | User tasks |
| idx_tasks_queue | Queue filtering |
| idx_tasks_status | Status filtering |
| idx_tasks_created | Recent tasks |

---

## 6.6 Triggers

| Index | Purpose |
|--------|---------|
| idx_triggers_task | Task lookup |
| idx_triggers_next_run | Scheduler lookup |
| idx_triggers_active | Active triggers |

---

## 6.7 Actions

| Index | Purpose |
|--------|---------|
| idx_actions_task | Find task actions |
| idx_actions_order | Ordered execution |

---

## 6.8 Executions

| Index | Purpose |
|--------|---------|
| idx_executions_task | Task history |
| idx_executions_status | Status filtering |
| idx_executions_started | Recent executions |
| idx_executions_trigger | Trigger history |

---

## 6.9 Execution Logs

| Index | Purpose |
|--------|---------|
| idx_logs_execution | Execution lookup |
| idx_logs_level | Log filtering |
| idx_logs_created | Recent logs |

---

## 6.10 Indexing Guidelines

Indexes improve read performance but increase storage usage and write overhead.

Only columns that are frequently searched, filtered, sorted, or joined should be indexed.

Additional indexes can be introduced later based on production usage and performance analysis.

---

# 7. Migration Strategy

Database migrations provide a version-controlled method of creating and modifying the database schema.

All schema changes should be implemented through migration files rather than manual SQL execution.

---

## Migration Directory

```text
backend/
└── migrations/
```

---

## Migration Naming

Migration files should follow sequential numbering.

Example:

```text
000001_create_users.up.sql
000001_create_users.down.sql

000002_create_queues.up.sql
000002_create_queues.down.sql

000003_create_tasks.up.sql
000003_create_tasks.down.sql
```

---

## Migration Rules

- Every migration must have an UP and DOWN script.
- Migrations should perform a single logical change.
- Existing migrations must never be modified after being committed.
- New schema changes should always be introduced through new migration files.

---

## Migration Order

The recommended migration order is:

1. Users
2. Email Verifications
3. Queues
4. User Settings
5. Tasks
6. Triggers
7. Actions
8. Executions
9. Execution Logs

This order ensures that all foreign key dependencies are satisfied.

---

# 8. Transaction Strategy

Transactions ensure that related database operations either complete successfully together or are rolled back entirely in case of failure.

---

## When Transactions Should Be Used

Transactions are recommended for operations involving multiple related database updates.

Examples include:

- User registration
- Task creation
- Task deletion
- Task updates
- Execution recording

---

## Rollback Policy

If any operation within a transaction fails, all previous operations within that transaction should be rolled back.

This prevents partial updates and maintains data consistency.

---

## Isolation

STAQ relies on PostgreSQL's transaction management to ensure consistency during concurrent operations.

The default isolation level is sufficient for Version 1.

Higher isolation levels may be considered in future versions if required.

---

# 9. Summary

The STAQ database uses PostgreSQL for persistent relational data and Redis for high-speed background processing.

The schema is normalized, protected by foreign key relationships, and optimized with carefully selected indexes.

Migration files provide a repeatable and version-controlled mechanism for evolving the database schema.

This design provides a strong foundation for reliable scheduling, workflow execution, execution tracking, and future scalability while remaining simple enough for efficient development and maintenance.

