# 1. Introduction

## 1.1 Purpose

STAQ is an intelligent task scheduling and workflow automation system designed to execute user-defined tasks based on time-based or event-based triggers.

It provides:
- Task scheduling
- Workflow automation
- Background execution
- Scalable worker processing

---

## 1.2 Scope

Version 1 focuses on:
- Task creation and management
- Trigger-based scheduling
- Action-based workflows
- Redis-based queue execution
- Worker-based processing engine

Out of scope:
- AI-based scheduling
- Voice control
- Browser automation
- External integrations (advanced plugins)

---

## 1.3 Goals

- Reliable task execution system
- Scalable background job processing
- Modular architecture
- Clean separation of concerns

---

# 2. System Overview

STAQ follows a distributed backend architecture:

- Client → React frontend
- Backend API → Go (Gin framework)
- Database → PostgreSQL (source of truth)
- Queue → Redis (job distribution)
- Workers → concurrent execution engine

---

## Core Flow

```
User → Task API → PostgreSQL
                ↓
           Trigger Engine
                ↓
           Scheduler
                ↓
             Redis Queue
                ↓
             Workers
                ↓
          Execution Logs
```

---

# 3. High-Level Architecture

## 3.1 Architecture Style

- Layered architecture
- Event-driven execution model
- Queue-based processing system

---

## 3.2 Components

- API Layer (Gin HTTP Server)
- Service Layer (Business Logic)
- Scheduler (Trigger evaluation)
- Queue (Redis)
- Worker Pool
- Database (PostgreSQL)

---

## 3.3 System Diagram

```
Frontend (React)
      |
      v
REST API (Go Gin)
      |
      v
Service Layer
      |
      +-------------------+
      |                   |
 PostgreSQL            Scheduler
      |                   |
      |                Redis Queue
      |                   |
      +-------- Worker Pool --------+
                     |
                Execution Engine
```

---

# 4. Core Components

## 4.1 API Layer

Handles:
- Authentication
- Task CRUD
- Trigger management
- Action management
- Execution queries

---

## 4.2 Service Layer

Responsible for:
- Business logic
- Validation
- Data transformation
- Orchestration between components

---

## 4.3 Database Layer

PostgreSQL stores:
- Users
- Tasks
- Triggers
- Actions
- Executions
- Logs

---

## 4.4 Redis Layer

Used for:
- Job queue
- Worker communication
- Retry handling
- Temporary execution state

---

## 4.5 Worker Pool

Responsible for:
- Consuming jobs from Redis
- Executing actions
- Handling retries
- Logging execution results

---

# 5. Authentication Flow

## Flow

```
Register → Store user → Send verification email
        → Verify email → Enable login
        → Login → JWT issued
```

---

## Key Components

- JWT authentication
- Password hashing (bcrypt)
- Email verification system

---

# 6. Task System

## 6.1 Task Definition

A Task is a user-defined automation unit containing:
- Metadata (name, description)
- Status
- Queue assignment
- Triggers
- Actions

---

## 6.2 Lifecycle

```
Created → Active → Scheduled → Executing → Completed
                     ↓
                  Paused
                     ↓
                 Archived
```

---

## 6.3 Responsibilities

- Define automation unit
- Link triggers and actions
- Manage execution lifecycle

---

## 6.4 Storage

Stored in PostgreSQL as the core entity of the system.

---

# 7. Trigger System

## 7.1 Overview

Triggers define when a task should be executed. They act as scheduling rules attached to tasks.

---

## 7.2 Types of Triggers

- CRON-based triggers
- Interval-based triggers
- One-time triggers
- Recurring triggers (daily, weekly, monthly)

---

## 7.3 Execution Logic

```
Trigger evaluated → If due → Create execution job → Send to scheduler
```

---

## 7.4 Responsibilities

- Store scheduling rules
- Calculate next execution time
- Support timezone-aware scheduling
- Provide execution metadata to scheduler

---

# 8. Action System

## 8.1 Overview

Actions define the actual work executed when a trigger fires.

Each task can have multiple actions executed in sequence.

---

## 8.2 Action Types

- HTTP Request
- Email Action
- Shell Command
- Internal function execution

---

## 8.3 Execution Order

Actions are executed sequentially based on `execution_order`.

```
Action 1 → Action 2 → Action 3 → Completion
```

---

## 8.4 Configuration

Each action stores configuration as JSONB:

Example:
```json
{
  "method": "POST",
  "url": "https://api.example.com"
}
```

---

# 9. Scheduler System

## 9.1 Overview

The Scheduler continuously evaluates triggers and generates execution jobs.

---

## 9.2 Responsibilities

- Scan active triggers
- Identify due executions
- Generate job payloads
- Push jobs to Redis queue

---

## 9.3 Execution Flow

```
Database Triggers
        ↓
Scheduler Engine
        ↓
Job Creation
        ↓
Redis Queue
```

---

## 9.4 Design Notes

- Runs as a background service
- Stateless design preferred
- Can scale horizontally if needed

---

# 10. Queue System (Redis)

## 10.1 Overview

Redis acts as a message broker between scheduler and workers.

---

## 10.2 Responsibilities

- Store execution jobs
- Handle retry queues
- Manage dead-letter queue
- Provide fast in-memory job access

---

## 10.3 Queue Types

- Main Queue
- Retry Queue
- Dead Letter Queue (DLQ)

---

## 10.4 Job Lifecycle

```
Scheduled Job → Redis Queue → Worker → Execution Result
```

---

# 11. Worker Pool

## 11.1 Overview

Workers consume jobs from Redis and execute task actions.

---

## 11.2 Responsibilities

- Consume queue jobs
- Execute actions sequentially
- Handle retries on failure
- Enforce timeout limits
- Log execution results

---

## 11.3 Concurrency Model

- Multiple workers run in parallel
- Each worker handles one job at a time

---

## 11.4 Failure Handling

- Retry mechanism with exponential backoff
- Move failed jobs to DLQ after max retries

---

## 11.5 Execution Flow

```
Job pulled from Redis
        ↓
Worker executes actions
        ↓
Success → Store execution log
Failure → Retry / DLQ
```

---

# 12. Execution System

## 12.1 Overview

The Execution System manages the lifecycle of a task run from start to finish.

It connects:
Scheduler → Worker → Database → Logs

---

## 12.2 Execution Lifecycle

```
Job received → Execution started → Actions executed → Result stored
```

---

## 12.3 Execution States

- PENDING
- RUNNING
- SUCCESS
- FAILED
- RETRYING

---

## 12.4 Responsibilities

- Track execution lifecycle
- Store execution metadata
- Capture logs and errors
- Manage retry state

---

## 12.5 Data Flow

```
Worker → Execution record → Logs → Database storage
```

---

# 13. Database Design Summary

## 13.1 Core Entities

- users
- email_verifications
- user_settings
- queues
- tasks
- triggers
- actions
- executions
- execution_logs

---

## 13.2 Design Principles

- Relational consistency (PostgreSQL)
- Foreign key enforcement
- Normalized schema
- Indexed for performance

---

## 13.3 Redis Role

- Temporary job storage
- Queue processing
- Retry handling
- DLQ management

---

# 14. Security Overview

## 14.1 Authentication

- JWT-based authentication
- Password hashing (bcrypt)
- Email verification required before login

---

## 14.2 Authorization

- Role-based access (future extension)
- User-scoped data isolation

---

## 14.3 API Security

- Input validation
- Rate limiting (future)
- Secure headers
- CORS configuration

---

## 14.4 Data Security

- Encrypted passwords
- Secure environment variables
- No sensitive data in logs

---

# 15. Deployment Architecture

## 15.1 Environments

- Development
- Staging (optional)
- Production

---

## 15.2 Infrastructure

- Backend: Go (Gin)
- Frontend: React
- Database: PostgreSQL (Neon)
- Queue: Redis (Upstash)
- Hosting: Vercel / Render / Railway

---

## 15.3 Deployment Flow

```
Git Push → CI/CD → Build → Deploy Backend + Frontend
```

---

## 15.4 Containerization

- Docker used for local development
- Docker Compose optional for full stack setup

---

# 16. Conclusion

STAQ is designed as a modular, scalable task automation platform.

Its architecture ensures:

- Separation of concerns
- Scalable execution engine
- Reliable task scheduling
- Fault-tolerant worker system

This architecture provides a strong foundation for implementing Version 1 and extending into future intelligent automation features.