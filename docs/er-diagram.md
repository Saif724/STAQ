# Entity Relationship Diagram

## STAQ
### Intelligent Task Scheduling & Workflow Automation Platform

Version: 1.0

Prepared By:
Ahsan Ahmed Saif

---

# 1. Introduction

## 1.1 Purpose

This document describes the Entity Relationship (ER) model of STAQ.

The ER model illustrates how entities within the system relate to one another and provides a visual representation of the database structure.

Detailed table definitions, constraints, and indexes are documented in `database-design.md`.

---

# 2. Entities

The Version 1 database contains the following entities.

| Entity | Description |
|----------|-------------|
| users | Stores user accounts |
| email_verifications | Stores email verification tokens |
| user_settings | Stores user preferences |
| queues | Stores execution queues |
| tasks | Stores user-created tasks |
| triggers | Stores task schedules |
| actions | Stores task actions |
| executions | Stores task execution history |
| execution_logs | Stores execution logs |

---

# 3. Entity Relationship Diagram

```text
                         +------------------+
                         |      users       |
                         +------------------+
                           |      |      |
                 1:N       |      |1:1   |1:N
                           |      |      |
                           ▼      ▼      ▼
                +----------------+   +-------------------------+
                |     tasks      |   |     user_settings       |
                +----------------+   +-------------------------+
                     |      |                |
          1:N        |      |1:N             |N:1
                     |      |                |
                     ▼      ▼                ▼
              +-----------+  +-----------+  +-----------+
              | triggers  |  | actions   |  |  queues   |
              +-----------+  +-----------+  +-----------+
                     |
                     |1:N
                     ▼
              +---------------+
              | executions    |
              +---------------+
                     |
                     |1:N
                     ▼
             +------------------+
             | execution_logs   |
             +------------------+

users
   │
   └──────────────► email_verifications (1:N)
```

---

# 4. Relationship Details

## User → Tasks

One user can create multiple tasks.

Each task belongs to exactly one user.

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

A user may receive multiple verification tokens.

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

A queue may be selected as the default queue for multiple users.

Each user has at most one default queue.

Relationship:

One-to-Many (1:N)

---

## Task → Triggers

Each task may contain multiple scheduling rules.

Each trigger belongs to one task.

Relationship:

One-to-Many (1:N)

---

## Task → Actions

Each task contains one or more actions executed in order.

Each action belongs to one task.

Relationship:

One-to-Many (1:N)

---

## Task → Executions

A task may execute many times throughout its lifetime.

Each execution belongs to one task.

Relationship:

One-to-Many (1:N)

---

## Trigger → Executions

Each execution is initiated by a specific trigger.

A trigger may create many executions.

Relationship:

One-to-Many (1:N)

---

## Execution → Execution Logs

Each execution may generate multiple log entries.

Each log belongs to one execution.

Relationship:

One-to-Many (1:N)

---

# 5. Cardinality Summary

| Relationship | Cardinality |
|--------------|-------------|
| User → Tasks | 1:N |
| User → User Settings | 1:1 |
| User → Email Verifications | 1:N |
| Queue → Tasks | 1:N |
| Queue → User Settings | 1:N |
| Task → Triggers | 1:N |
| Task → Actions | 1:N |
| Task → Executions | 1:N |
| Trigger → Executions | 1:N |
| Execution → Execution Logs | 1:N |

---

# 6. Design Notes

The database follows a normalized relational design.

Key design decisions include:

- Every entity uses a UUID primary key.
- Tasks own actions directly; there is no separate workflow table in Version 1.
- Action configurations are stored using PostgreSQL JSONB.
- Redis is used for transient queue data and is not part of the ER model.
- Foreign key relationships enforce referential integrity.
- The schema is designed to support future expansion without significant structural changes.

---

# 7. Summary

The ER model provides a clear representation of the relationships between entities in STAQ.

It serves as the blueprint for the PostgreSQL schema and ensures consistency between the application architecture and the database implementation.