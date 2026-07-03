# 1. Introduction

## 1.1 Purpose

This document describes the software architecture of **STAQ (Smart Task Automation Queue)**.

Its purpose is to provide a complete technical overview of how the system is structured, how its components interact, and the design principles used during development.

While the Software Requirements Specification (SRS) defines **what** STAQ should accomplish, this document explains **how** those requirements are implemented through the system's architecture.

This document serves as the primary technical reference during implementation and should be consulted whenever new features are added or existing components are modified.

---

## 1.2 Scope

This architecture document applies to **STAQ Version 1.0**.

It covers the complete backend architecture, frontend interaction, execution pipeline, scheduling engine, worker system, data flow, authentication, configuration management, logging strategy, and scalability considerations.

Version 1 focuses on building a reliable and extensible task scheduling and workflow automation platform. Advanced capabilities such as natural language processing, voice interaction, AI-assisted task creation, browser automation, and third-party productivity integrations are intentionally excluded from this version and are considered future enhancements.

---

## 1.3 Intended Audience

This document is intended for:

* Developers implementing STAQ.
* Contributors who want to understand the project architecture.
* Reviewers evaluating the system design.
* Future maintainers extending the platform.

Although STAQ is currently a solo project, the architecture follows established software engineering practices to support future collaboration and long-term maintainability.

---

## 1.4 Architectural Philosophy

STAQ is designed around four core principles:

### Separation of Responsibilities

Each major component has a clearly defined responsibility.

For example:

* The Scheduler determines **when** tasks should execute.
* The Job Service prepares executable jobs.
* The Broker publishes jobs to the queue.
* Workers consume queued jobs.
* The Execution Engine performs task execution.
* Action Executors implement individual action types.

This separation reduces coupling and makes each component easier to understand and maintain.

---

### Modularity

The application is divided into independent modules based on business functionality.

Examples include:

* Authentication
* Tasks
* Triggers
* Jobs
* Workers
* Notifications

Each module owns its business logic, allowing features to evolve independently while minimizing the impact on other parts of the system.

---

### Extensibility

STAQ is designed so that new action types, scheduling strategies, and integrations can be added with minimal changes to existing components.

For example, adding support for a new action type should require implementing a new Action Executor without modifying the Scheduler or Worker logic.

This approach enables the system to evolve without frequent architectural changes.

---

### Reliability

Task automation systems must operate reliably over long periods.

The architecture therefore emphasizes:

* Persistent storage of business data.
* Reliable message delivery through queues.
* Retry mechanisms for temporary failures.
* Execution tracking.
* Failure logging.
* Recoverable background services.

These design decisions reduce the likelihood of lost tasks and improve operational stability.

---

## 1.5 Document Organization

The remainder of this document is organized as follows:

* **Chapter 2** explains the architectural goals and design objectives.
* **Chapter 3** introduces the technologies selected for STAQ and the reasons behind those choices.
* **Chapter 4** presents the high-level system architecture and component relationships.
* **Subsequent chapters** describe each major subsystem in detail, including the scheduler, job service, broker, worker pool, execution engine, authentication, logging, configuration, scalability, and security.

Together, these chapters provide a complete blueprint for implementing and maintaining STAQ Version 1.0.

---

# 2. Architecture Goals

The architecture of STAQ is designed to provide a reliable, scalable, and maintainable foundation for task scheduling and workflow automation.

Rather than optimizing for a single use case, the system is designed to support a wide range of automation scenarios while remaining simple to understand and extend.

The following architectural goals guided every major design decision.

---

## 2.1 Simplicity

The architecture should be easy to understand and easy to maintain.

Although STAQ consists of multiple independent components, each component performs a single well-defined responsibility.

For example:

* The Scheduler determines when a task should execute.
* The Job Service prepares executable jobs.
* The Broker publishes jobs to Redis.
* Workers process queued jobs.
* The Execution Engine coordinates action execution.

This separation keeps the overall system understandable while reducing unnecessary complexity.

Whenever multiple design options exist, the simpler solution is preferred unless a more complex solution provides significant long-term benefits.

---

## 2.2 Modularity

STAQ is organized into independent business modules.

Each module encapsulates its own logic and communicates with other modules through clearly defined interfaces.

Examples of modules include:

* Authentication
* Users
* Tasks
* Triggers
* Jobs
* Queues
* Workers
* Notifications
* Executions

This modular design allows individual features to evolve independently without affecting unrelated parts of the system.

Benefits include:

* Easier maintenance
* Better code organization
* Independent testing
* Faster feature development
* Lower coupling

---

## 2.3 Scalability

The architecture is designed to scale without requiring significant redesign.

Scalability is achieved by separating task scheduling from task execution.

Instead of executing scheduled tasks immediately, STAQ follows a queue-based architecture:

```text
Scheduler
    │
    ▼
Job Service
    │
    ▼
Broker
    │
    ▼
Redis Queue
    │
    ▼
Worker Pool
```

Because Workers operate independently, increasing system capacity only requires starting additional Worker instances.

Similarly, Scheduler and API services can be deployed independently if future workloads require horizontal scaling.

---

## 2.4 Reliability

A task scheduler must prioritize reliability over raw performance.

STAQ is designed to minimize the possibility of lost tasks or incomplete executions.

Reliability is achieved through:

* Persistent PostgreSQL storage
* Redis-based message queue
* Configurable retry policies
* Execution logging
* Worker isolation
* Error recovery
* Background processing

If temporary failures occur, the system retries execution according to the configured retry policy instead of immediately marking the task as failed.

---

## 2.5 Extensibility

One of the primary goals of STAQ is to support future expansion without requiring architectural redesign.

The Action Executor architecture allows new action types to be added independently.

For example, Version 1 supports:

* Email
* HTTP Requests
* Shell Commands
* Reminders

Future versions may introduce:

* Discord Messages
* Slack Notifications
* GitHub Actions
* Calendar Events
* File Uploads
* AI-powered Actions

Adding these features should require implementing a new executor while leaving the Scheduler, Worker, and Execution Engine unchanged.

---

## 2.6 Separation of Concerns

Every component in STAQ has a clearly defined responsibility.

Responsibilities are intentionally isolated.

| Component        | Primary Responsibility  |
| ---------------- | ----------------------- |
| REST API         | Handle client requests  |
| Scheduler        | Detect due triggers     |
| Job Service      | Build execution jobs    |
| Broker           | Publish jobs to Redis   |
| Worker           | Process queued jobs     |
| Execution Engine | Execute workflows       |
| Action Registry  | Resolve executors       |
| Action Executor  | Perform a single action |

This separation reduces coupling and improves maintainability.

---

## 2.7 Maintainability

The project is expected to grow over time.

To simplify maintenance, the architecture follows several design principles:

* Clear package boundaries
* Consistent naming conventions
* Dependency injection
* Repository pattern
* Standardized error handling
* Centralized configuration
* Structured logging

These practices reduce technical debt and make future modifications easier.

---

## 2.8 Testability

The architecture should support automated testing.

Components communicate through interfaces wherever appropriate, allowing implementations to be replaced during testing.

Examples include:

* Repository implementations
* Email providers
* Redis communication
* External APIs

This makes unit testing possible without requiring real infrastructure.

---

## 2.9 Security

Security is considered throughout the architecture rather than being added later.

Version 1 includes:

* JWT Authentication
* Refresh Tokens
* Password Hashing
* Email Verification
* Protected API Endpoints
* Input Validation
* Role-Based Authorization

Sensitive information such as database credentials, SMTP credentials, and JWT secrets are loaded from environment variables rather than hardcoded into the application.

---

## 2.10 Observability

Understanding system behavior is essential for debugging and monitoring.

STAQ therefore records important operational events.

Examples include:

* User authentication
* Task creation
* Job publication
* Worker activity
* Execution duration
* Execution failures
* Retry attempts

These records simplify debugging while providing visibility into system performance.

---

## 2.11 Future Readiness

Although Version 1 focuses on reliable task scheduling, the architecture is intentionally designed to accommodate future enhancements.

Planned capabilities include:

* Natural Language Task Creation
* Voice Commands
* AI Workflow Generation
* Calendar Integration
* Browser Automation
* Plugin System
* Distributed Worker Clusters

These features should integrate with the existing architecture without requiring major structural changes.

---

## 2.12 Summary

The architecture of STAQ is guided by a small set of principles:

* Keep responsibilities clearly separated.
* Design independent, modular components.
* Prioritize reliability over unnecessary complexity.
* Build for future extensibility.
* Maintain consistent code organization.
* Enable horizontal scalability.
* Ensure security from the beginning.

These principles provide the foundation for every architectural decision described throughout the remainder of this document.

---

# 3. Technology Stack

Selecting the right technologies is essential to building a reliable, maintainable, and scalable task scheduling platform.

The technologies chosen for STAQ were evaluated based on performance, ecosystem maturity, ease of development, long-term maintainability, and suitability for concurrent background processing.

This chapter explains the purpose of each technology and the reasoning behind its selection.

---

## 3.1 Technology Overview

| Layer               | Technology        |
| ------------------- | ----------------- |
| Frontend            | React             |
| Language            | TypeScript        |
| Styling             | Tailwind CSS      |
| Backend             | Go                |
| HTTP Framework      | Gin               |
| Database            | PostgreSQL        |
| Cache / Queue       | Redis             |
| Authentication      | JWT               |
| Password Hashing    | bcrypt            |
| Email               | SMTP              |
| API Documentation   | Swagger (OpenAPI) |
| Containerization    | Docker            |
| Local Orchestration | Docker Compose    |

---

# 3.2 Backend

## Go

Go is the primary programming language used to build the backend services of STAQ.

It was selected because it provides:

* Excellent concurrency through Goroutines
* Low memory usage
* Fast execution speed
* Strong standard library
* Simple deployment
* Easy cross-platform compilation

Since STAQ executes many background jobs simultaneously, Go's lightweight concurrency model is particularly well suited for worker pools and schedulers.

---

## Gin

Gin is used as the HTTP framework.

Responsibilities include:

* Routing
* Middleware
* Request parsing
* Response generation
* Authentication middleware
* API grouping

Gin was selected because it offers:

* High performance
* Minimal overhead
* Clean API design
* Large community support
* Easy middleware integration

---

# 3.3 Database

## PostgreSQL

PostgreSQL serves as the primary persistent data store.

It stores:

* Users
* Tasks
* Triggers
* Actions
* Queues
* Executions
* Notifications
* Refresh Tokens
* Email Verification Tokens

Reasons for choosing PostgreSQL include:

* ACID-compliant transactions
* Strong indexing capabilities
* Excellent query performance
* Mature ecosystem
* Advanced SQL features
* Reliable data integrity

Because STAQ manages highly relational data, PostgreSQL is a natural choice.

---

# 3.4 Queue System

## Redis

Redis is used exclusively as the message broker and execution queue.

It is **not** the primary database.

Responsibilities include:

* Holding pending jobs
* Delivering jobs to workers
* Buffering execution requests
* Decoupling scheduling from execution

Redis provides:

* Extremely low latency
* High throughput
* Lightweight deployment
* Excellent Go libraries
* Reliable queue operations

By separating temporary execution data from permanent business data, Redis and PostgreSQL each perform the tasks they are best suited for.

---

# 3.5 Frontend

## React

React is used to build the user interface.

Responsibilities include:

* Authentication pages
* Dashboard
* Task Management
* Trigger Management
* Queue Monitoring
* Execution History
* User Profile
* Notifications

React was selected because it offers:

* Component-based architecture
* Strong ecosystem
* Efficient rendering
* Excellent TypeScript support
* Large community

---

## TypeScript

TypeScript adds static typing to JavaScript.

Benefits include:

* Better IDE support
* Early error detection
* Improved maintainability
* Safer refactoring
* Better documentation through types

For medium and large applications, TypeScript significantly improves code quality.

---

## Tailwind CSS

Tailwind CSS is used for styling the frontend.

Advantages include:

* Utility-first workflow
* Rapid UI development
* Consistent design
* Minimal custom CSS
* Excellent responsiveness

---

# 3.6 Authentication

Authentication is based on JSON Web Tokens (JWT).

The authentication system includes:

* Access Tokens
* Refresh Tokens
* Email Verification
* Password Hashing
* Role-Based Authorization

Passwords are never stored in plain text.

Instead, bcrypt is used to securely hash passwords before they are stored in PostgreSQL.

---

# 3.7 Email Service

STAQ supports sending emails through SMTP.

Version 1 uses SMTP for:

* Email verification
* Password reset
* Scheduled email actions

Future versions may support additional providers such as SendGrid or Amazon SES without changing the overall architecture.

---

# 3.8 API Documentation

All REST APIs are documented using the OpenAPI Specification.

Swagger UI provides an interactive interface for testing endpoints during development.

Benefits include:

* Interactive API exploration
* Request and response examples
* Automatic documentation generation
* Improved frontend-backend collaboration

---

# 3.9 Containerization

## Docker

Docker provides a consistent runtime environment for STAQ.

Benefits include:

* Reproducible deployments
* Environment isolation
* Easier onboarding
* Consistent development and production environments

---

## Docker Compose

Docker Compose simplifies local development by orchestrating multiple services.

Version 1 includes:

* Backend
* Frontend
* PostgreSQL
* Redis

This allows developers to start the entire application with a single command.

---

# 3.10 Development Tools

The following tools support development:

| Tool            | Purpose               |
| --------------- | --------------------- |
| Git             | Version Control       |
| GitHub          | Source Code Hosting   |
| VS Code         | Primary IDE           |
| Postman / Bruno | API Testing           |
| pgAdmin         | PostgreSQL Management |
| Redis Insight   | Redis Inspection      |

These tools are not part of the production system but improve the development workflow.

---

# 3.11 Technology Selection Principles

The technology stack was chosen according to the following principles:

* Prefer stable technologies over experimental ones.
* Use specialized tools for specialized tasks.
* Minimize unnecessary dependencies.
* Keep deployment straightforward.
* Favor long-term maintainability over short-term convenience.

Each technology in STAQ has a clearly defined purpose, reducing overlap and simplifying future maintenance.

---

# 3.12 Summary

The selected technology stack provides a balance of performance, simplicity, and extensibility.

Go and Redis enable efficient background processing, PostgreSQL ensures reliable data persistence, React delivers a modern user experience, and Docker simplifies deployment.

Together, these technologies provide a strong foundation for implementing STAQ Version 1.0 while supporting future enhancements as the platform evolves.

---

# 4. High-Level Architecture

STAQ follows a modular, queue-driven architecture designed to separate task scheduling from task execution. Each major subsystem has a single responsibility and communicates with other components through well-defined interfaces.

This design improves maintainability, scalability, and reliability while allowing individual components to evolve independently.

---

# 4.1 System Overview

The architecture consists of four primary layers:

1. Presentation Layer
2. Application Layer
3. Processing Layer
4. Data Layer

Each layer has a specific responsibility and interacts only with adjacent layers whenever possible.

---

# 4.2 High-Level Architecture Diagram

```text
                        +----------------------+
                        |    React Frontend    |
                        +----------+-----------+
                                   |
                             REST API (HTTPS)
                                   |
                                   ▼
                    +-----------------------------+
                    |        Gin REST API         |
                    +-------------+---------------+
                                  |
             +--------------------+--------------------+
             |                    |                    |
             ▼                    ▼                    ▼
       Authentication         Task Service       User Service
             |                    |                    |
             +--------------------+--------------------+
                                  |
                                  ▼
                         PostgreSQL Database
                                  ▲
                                  |
                        Scheduler (Background)
                                  |
                                  ▼
                           Job Service
                                  |
                                  ▼
                               Broker
                                  |
                                  ▼
                            Redis Queue
                                  |
                    +-------------+-------------+
                    |             |             |
                    ▼             ▼             ▼
                 Worker 1      Worker 2      Worker N
                    |             |             |
                    +-------------+-------------+
                                  |
                                  ▼
                         Execution Engine
                                  |
                                  ▼
                          Action Registry
                                  |
                                  ▼
                      Appropriate Action Executor
                                  |
        +-----------+------------+------------+-------------+
        |           |            |            |             |
        ▼           ▼            ▼            ▼             ▼
   Email Action  HTTP Action  Reminder    Shell Action   Future Actions
                                  |
                                  ▼
                         Execution Result
                                  |
                                  ▼
                           PostgreSQL
                                  |
                                  ▼
                          Notifications
```

---

# 4.3 Architectural Layers

## Presentation Layer

The Presentation Layer consists of the React frontend.

Responsibilities include:

* User Authentication
* Dashboard
* Task Management
* Trigger Management
* Queue Monitoring
* Execution History
* Profile Management
* Notifications

This layer never communicates directly with the database.

All communication occurs through the REST API.

---

## Application Layer

The Application Layer is implemented using Go and the Gin framework.

Responsibilities include:

* Request Validation
* Business Logic
* Authentication
* Authorization
* Database Operations
* Scheduling
* Queue Management

This layer coordinates all business operations but delegates background execution to the processing layer.

---

## Processing Layer

The Processing Layer is responsible for asynchronous work.

Its components include:

* Scheduler
* Job Service
* Broker
* Redis Queue
* Worker Pool
* Execution Engine
* Action Registry
* Action Executors

Separating processing from the REST API ensures that long-running tasks do not block user requests.

---

## Data Layer

The Data Layer provides persistent and temporary storage.

### PostgreSQL

Stores:

* Users
* Tasks
* Triggers
* Queues
* Executions
* Notifications
* Email Verification Tokens
* Refresh Tokens

### Redis

Stores:

* Pending Jobs
* Processing Jobs
* Retry Jobs (if applicable)

Redis acts only as a fast message broker and is never used as the primary source of business data.

---

# 4.4 Design Principles

The high-level architecture follows several important principles.

### Separation of Concerns

Each subsystem has a single, clearly defined responsibility.

Examples include:

* Scheduler decides **when** work begins.
* Job Service prepares work.
* Broker distributes work.
* Worker processes work.
* Execution Engine coordinates execution.
* Action Executors perform the actual business action.

---

### Loose Coupling

Components communicate through interfaces rather than depending directly on implementations.

For example, the Worker does not know how an Email Action works. It only requests the appropriate executor from the Action Registry.

This allows new action types to be added without modifying worker logic.

---

### High Cohesion

Each package groups closely related functionality together.

For example:

* Authentication logic remains within the `auth` package.
* Scheduling logic remains within the `scheduler` package.
* Task management remains within the `tasks` package.

This organization improves readability and simplifies maintenance.

---

### Asynchronous Processing

Time-consuming work is performed asynchronously.

Instead of executing tasks during an HTTP request:

1. The Scheduler detects that a trigger is due.
2. The Job Service creates an execution job.
3. The Broker publishes the job to Redis.
4. Workers process the job independently.

Asynchronous execution keeps API responses fast and allows the system to process many jobs concurrently.

---

### Fault Isolation

Failures in one task should not affect other tasks.

If a Worker encounters an error while processing a job:

* The error is recorded.
* Retry policies are applied if configured.
* Other Workers continue processing unaffected.

This improves system stability and reliability.

---

# 4.5 Component Communication

Communication between components follows a predictable flow.

```text
Frontend
    │
    ▼
REST API
    │
    ▼
PostgreSQL
    ▲
    │
Scheduler
    │
    ▼
Job Service
    │
    ▼
Broker
    │
    ▼
Redis Queue
    │
    ▼
Worker
    │
    ▼
Execution Engine
    │
    ▼
Action Registry
    │
    ▼
Action Executor
    │
    ▼
Execution Stored
    │
    ▼
Notification Generated
```

Every component performs exactly one step before passing responsibility to the next component.

---

# 4.6 Summary

The architecture of STAQ separates user interaction, business logic, asynchronous processing, and data storage into independent layers.

This modular design allows the platform to:

* Process thousands of scheduled jobs efficiently.
* Scale Workers independently.
* Keep HTTP requests responsive.
* Add new action types without changing existing components.
* Maintain a clean separation of responsibilities.

The following chapters examine each subsystem in detail, beginning with the backend architecture and request lifecycle.

---

# 5. Core Components

STAQ is composed of multiple independent components that work together to schedule, queue, execute, and monitor automated tasks.

Each component has a single responsibility and communicates with other components through well-defined interfaces. This separation improves maintainability, scalability, and testability.

The following sections describe the purpose and responsibilities of each core component.

---

# 5.1 React Frontend

The React Frontend is the user-facing interface of STAQ.

It communicates exclusively with the REST API and never interacts directly with the database or Redis.

### Responsibilities

* User Registration
* User Login
* Email Verification
* Dashboard
* Task Management
* Trigger Management
* Queue Monitoring
* Execution History
* Notification Center
* User Profile

The frontend is responsible only for presentation and user interaction.

---

# 5.2 REST API

The REST API is the primary entry point for all client requests.

It validates incoming requests, authenticates users, executes business logic, and returns structured JSON responses.

### Responsibilities

* Routing
* Authentication
* Authorization
* Input Validation
* Business Logic Coordination
* Response Formatting
* Error Handling

The API is intentionally stateless, allowing multiple API instances to run simultaneously behind a load balancer if needed.

---

# 5.3 PostgreSQL Database

PostgreSQL is the primary source of persistent data.

All business entities are stored here.

### Stores

* Users
* Tasks
* Triggers
* Queues
* Executions
* Notifications
* Refresh Tokens
* Email Verification Tokens

PostgreSQL is considered the single source of truth for STAQ.

---

# 5.4 Scheduler

The Scheduler continuously checks for triggers that are due for execution.

It does not execute tasks.

Instead, it identifies work that needs to be performed.

### Responsibilities

* Load active triggers
* Detect due schedules
* Validate trigger state
* Prevent duplicate scheduling
* Forward due triggers to the Job Service

The Scheduler performs no business actions beyond deciding *when* work should begin.

---

# 5.5 Job Service

The Job Service transforms scheduled triggers into executable jobs.

It prepares all information required for background execution before publishing the job to the queue.

### Responsibilities

* Load task details
* Load associated actions
* Validate configuration
* Generate Execution ID
* Build Job Payload
* Initialize Retry Policy
* Pass the job to the Broker

This component separates scheduling logic from execution preparation.

---

# 5.6 Broker

The Broker publishes prepared jobs to Redis.

It acts as the communication layer between the Job Service and Worker Pool.

### Responsibilities

* Serialize jobs
* Select destination queue
* Publish jobs to Redis
* Report publishing failures

The Broker is not responsible for executing jobs.

---

# 5.7 Redis Queue

Redis temporarily stores executable jobs until a Worker retrieves them.

Redis acts only as a message broker and not as a permanent data store.

### Responsibilities

* Queue pending jobs
* Deliver jobs to Workers
* Buffer execution requests
* Improve asynchronous processing

Redis enables decoupling between task scheduling and task execution.

---

# 5.8 Worker Pool

Workers continuously listen for new jobs in Redis.

Each Worker processes one job at a time before requesting the next available job.

### Responsibilities

* Retrieve jobs
* Handle retries
* Apply execution timeouts
* Delegate execution to the Execution Engine
* Report execution results

Multiple Workers may run concurrently to increase processing capacity.

---

# 5.9 Execution Engine

The Execution Engine coordinates the execution of a job.

Rather than implementing action-specific logic, it orchestrates the execution process.

### Responsibilities

* Parse job payload
* Execute workflow actions
* Handle execution order
* Manage execution context
* Capture execution results
* Record execution status

The Execution Engine acts as the central coordinator for background execution.

---

# 5.10 Action Registry

The Action Registry maps an action type to its corresponding Action Executor.

For example:

```text
EMAIL  → Email Executor
HTTP   → HTTP Executor
SHELL  → Shell Executor
```

### Responsibilities

* Register supported executors
* Resolve executor implementations
* Return the appropriate executor
* Report unsupported action types

The registry allows new action types to be added without modifying the Worker or Execution Engine.

---

# 5.11 Action Executors

Action Executors implement the actual business operations performed by STAQ.

Each executor is responsible for one action type only.

### Version 1 Executors

* Reminder Executor
* Email Executor
* HTTP Executor
* Shell Executor

Each executor implements a common interface, ensuring consistent behavior across all action types.

---

# 5.12 Notification Service

The Notification Service informs users about important events occurring within STAQ.

### Responsibilities

* Task completion notifications
* Task failure notifications
* Reminder notifications
* System notifications

Version 1 supports in-application notifications. Additional delivery channels may be added in future versions.

---

# 5.13 Authentication Service

The Authentication Service manages user identity and access control.

### Responsibilities

* User Registration
* User Login
* Password Hashing
* JWT Generation
* Refresh Token Management
* Email Verification
* Password Reset

Only authenticated users may access protected resources.

---

# 5.14 Logging System

The Logging System records significant events throughout the application.

Examples include:

* API requests
* Authentication events
* Scheduler activity
* Worker lifecycle
* Job publication
* Execution failures
* Retry attempts

Structured logging simplifies debugging and operational monitoring.

---

# 5.15 Configuration System

The Configuration System loads application settings from environment variables and configuration files.

Examples include:

* Database Connection
* Redis Connection
* SMTP Credentials
* JWT Secret
* Worker Count
* Scheduler Interval

Centralized configuration simplifies deployment across different environments.

---

# 5.16 Component Relationships

The core components cooperate according to the following sequence:

```text
Frontend
      │
      ▼
REST API
      │
      ▼
PostgreSQL
      ▲
      │
Scheduler
      │
      ▼
Job Service
      │
      ▼
Broker
      │
      ▼
Redis Queue
      │
      ▼
Worker Pool
      │
      ▼
Execution Engine
      │
      ▼
Action Registry
      │
      ▼
Action Executor
      │
      ▼
Execution Stored
      │
      ▼
Notification Service
```

Each component performs a clearly defined responsibility before handing control to the next stage.

---

# 5.17 Summary

The architecture of STAQ is built around independent components with clearly defined responsibilities.

This modular approach provides several advantages:

* Easier maintenance
* Improved scalability
* Better testability
* Clear separation of concerns
* Simplified feature expansion

As the project evolves, new components can be introduced without disrupting the existing execution pipeline.

---

# 6. Backend Architecture

The STAQ backend follows a layered architecture that separates HTTP handling, business logic, data access, and background processing into distinct layers.

Each layer has a clearly defined responsibility and depends only on the layer directly beneath it.

This organization improves readability, maintainability, testability, and scalability.

---

# 6.1 Backend Overview

The backend consists of three executable applications:

* API Server
* Scheduler
* Worker

Although they share the same codebase, each executable performs a different responsibility.

```text
backend/

cmd/
├── api/
├── scheduler/
└── worker/
```

This separation allows each service to be deployed independently if needed.

---

# 6.2 Layered Architecture

The backend follows the following logical structure:

```text
                HTTP Request
                     │
                     ▼
              Gin Router
                     │
                     ▼
              Middleware
                     │
                     ▼
                Handler
                     │
                     ▼
                Service
                     │
                     ▼
              Repository
                     │
                     ▼
             PostgreSQL / Redis
```

Each layer communicates only with the layer immediately below it.

---

# 6.3 Router Layer

The Router is responsible for mapping incoming HTTP requests to the appropriate handlers.

Examples include:

* Authentication Routes
* User Routes
* Task Routes
* Trigger Routes
* Queue Routes
* Execution Routes
* Notification Routes

The Router contains no business logic.

Its only responsibility is request routing.

---

# 6.4 Middleware Layer

Middleware executes before a request reaches a handler.

Each middleware performs one specific task.

Version 1 includes:

* Request Logging
* Panic Recovery
* CORS
* JWT Authentication
* Role Authorization

Middleware should remain lightweight and avoid business logic.

---

# 6.5 Handler Layer

Handlers receive validated HTTP requests.

Responsibilities include:

* Reading request parameters
* Parsing JSON
* Returning HTTP responses
* Calling Services
* Mapping errors to status codes

Handlers should contain minimal logic.

Business decisions belong in the Service layer.

---

# 6.6 Service Layer

The Service layer contains the application's business logic.

Examples include:

* Register User
* Login User
* Create Task
* Update Task
* Pause Task
* Resume Task
* Delete Task
* Verify Email

Services coordinate repositories and other components to perform complete business operations.

---

# 6.7 Repository Layer

Repositories abstract database access.

Instead of writing SQL directly inside Services, Services communicate with repositories.

Responsibilities include:

* Create records
* Read records
* Update records
* Delete records
* Execute queries

This abstraction improves testability and isolates persistence logic.

---

# 6.8 Background Processing

Unlike traditional web applications, STAQ performs much of its work asynchronously.

Background execution follows this pipeline:

```text
Scheduler
      │
      ▼
Job Service
      │
      ▼
Broker
      │
      ▼
Redis Queue
      │
      ▼
Worker
      │
      ▼
Execution Engine
      │
      ▼
Action Executor
```

This processing occurs independently of incoming HTTP requests.

---

# 6.9 Package Responsibilities

Each package inside `internal/` owns a specific business capability.

Examples include:

| Package       | Responsibility                   |
| ------------- | -------------------------------- |
| auth          | Authentication and authorization |
| users         | User management                  |
| tasks         | Task management                  |
| triggers      | Scheduling rules                 |
| jobs          | Job preparation                  |
| broker        | Queue publishing                 |
| workers       | Background workers               |
| actions       | Action execution                 |
| executions    | Execution history                |
| notifications | User notifications               |

Packages should avoid depending directly on unrelated packages whenever possible.

---

# 6.10 Dependency Direction

Dependencies always point downward.

```text
Handler
    │
    ▼
Service
    │
    ▼
Repository
    │
    ▼
Database
```

A Repository must never call a Service.

A Handler must never call a Repository directly.

This rule prevents circular dependencies and keeps responsibilities clearly separated.

---

# 6.11 Error Handling

Errors propagate upward through the layers.

For example:

```text
Database Error
        │
        ▼
Repository
        │
        ▼
Service
        │
        ▼
Handler
        │
        ▼
HTTP Response
```

Each layer may add context to the error before passing it upward.

Sensitive implementation details should never be exposed to API clients.

---

# 6.12 Logging

Logging occurs throughout the backend.

Examples include:

* Incoming requests
* Authentication attempts
* Task creation
* Scheduler activity
* Job publication
* Worker execution
* Execution failures

Logs should be structured and include sufficient context to support debugging and monitoring.

---

# 6.13 Summary

The backend architecture separates responsibilities into well-defined layers.

This layered approach improves code organization, simplifies testing, reduces coupling, and provides a strong foundation for future growth.

The following chapters examine the execution flow of STAQ in greater detail, beginning with the lifecycle of a request through the system.

---

# 7. Request Lifecycle

This chapter describes how an HTTP request flows through the STAQ backend, from the moment a client sends a request until a response is returned.

Understanding this lifecycle ensures that every API endpoint follows a consistent processing pipeline and maintains a clear separation of responsibilities.

---

# 7.1 Overview

Every request received by the API follows the same sequence.

```text
Client
   │
   ▼
Gin Router
   │
   ▼
Middleware
   │
   ▼
Handler
   │
   ▼
Service
   │
   ▼
Repository
   │
   ▼
Database
   │
   ▲
Repository
   │
   ▲
Service
   │
   ▲
Handler
   │
   ▲
HTTP Response
```

Each layer has a single responsibility.

---

# 7.2 Step 1 – Client Request

The lifecycle begins when the frontend sends an HTTP request.

Example:

```http
POST /api/v1/tasks
```

The request contains:

* HTTP Method
* URL
* Headers
* Authentication Token
* JSON Body

Example JSON:

```json
{
    "name": "Daily Study",
    "description": "Practice Go",
    "queue_id": 1
}
```

---

# 7.3 Step 2 – Router

Gin receives the request and determines which Handler should process it.

Example:

```text
POST /tasks

↓

Task Handler
```

The Router performs no business logic.

---

# 7.4 Step 3 – Middleware

Before the Handler executes, middleware processes the request.

Typical order:

```text
Logging

↓

Recovery

↓

CORS

↓

JWT Authentication

↓

Role Authorization

↓

Handler
```

Each middleware performs one specific responsibility.

If any middleware fails, request processing stops immediately.

Example:

Missing JWT

↓

401 Unauthorized

No Handler execution

---

# 7.5 Step 4 – Handler

Handlers are responsible for HTTP concerns only.

Responsibilities include:

* Parse JSON
* Validate request format
* Read URL parameters
* Read query parameters
* Call Service
* Return HTTP response

Example:

```text
CreateTaskHandler

↓

taskService.CreateTask(...)
```

Handlers never access the database directly.

---

# 7.6 Step 5 – Service

Services contain the business logic.

Example:

Create Task

↓

Validate Queue

↓

Validate User

↓

Generate Task

↓

Call Repository

↓

Return Result

Services may coordinate multiple repositories when necessary.

Business rules belong here—not in Handlers.

---

# 7.7 Step 6 – Repository

Repositories communicate with PostgreSQL.

Responsibilities include:

* Insert records
* Update records
* Delete records
* Execute queries

Repositories never contain business rules.

Example:

```text
TaskRepository

↓

INSERT INTO tasks (...)
```

---

# 7.8 Step 7 – Database

PostgreSQL executes the SQL query.

Possible outcomes:

* Success
* Constraint Violation
* Duplicate Key
* Foreign Key Error
* Connection Failure

The database returns the result to the Repository.

---

# 7.9 Step 8 – Response

The response travels back through the layers.

```text
Database

↓

Repository

↓

Service

↓

Handler

↓

HTTP Response

↓

Frontend
```

Each layer may add context before returning the final response.

---

# 7.10 Example – User Login

```text
POST /login

↓

Router

↓

Handler

↓

Auth Service

↓

User Repository

↓

PostgreSQL

↓

Password Verification

↓

JWT Generation

↓

HTTP Response
```

---

# 7.11 Example – Create Task

```text
POST /tasks

↓

JWT Middleware

↓

Task Handler

↓

Task Service

↓

Task Repository

↓

PostgreSQL

↓

201 Created
```

Creating a task does **not** immediately execute it.

Execution occurs later through the Scheduler.

---

# 7.12 Example – Pause Task

```text
PATCH /tasks/{id}/pause

↓

Task Handler

↓

Task Service

↓

Repository

↓

Task Status Updated

↓

Scheduler ignores paused task
```

This illustrates how the API modifies system state while background services react later.

---

# 7.13 Error Flow

Errors propagate upward.

```text
Database Error

↓

Repository

↓

Service

↓

Handler

↓

HTTP Error Response
```

Example:

```http
409 Conflict
```

```json
{
    "error": "Task name already exists."
}
```

The client receives a meaningful error without exposing internal implementation details.

---

# 7.14 Success Response

A successful request follows the same path.

Example:

```http
201 Created
```

```json
{
    "id": 12,
    "name": "Daily Study",
    "status": "ACTIVE"
}
```

Consistent response formats simplify frontend development and API integration.

---

# 7.15 Summary

Every HTTP request in STAQ follows the same layered processing model:

1. Client sends request.
2. Router selects the Handler.
3. Middleware validates the request.
4. Handler parses HTTP data.
5. Service executes business logic.
6. Repository communicates with PostgreSQL.
7. Response returns to the client.

This predictable lifecycle improves maintainability, simplifies debugging, and ensures that business logic remains isolated from HTTP and persistence concerns.

---

# 8. Execution Architecture

The execution architecture is the core of STAQ.

It defines how scheduled tasks move through the system, from the moment a trigger becomes due until every configured action has been executed and the execution result has been stored.

Unlike synchronous web applications, STAQ performs task execution asynchronously through a queue-based processing pipeline.

This architecture ensures that long-running operations never block API requests while allowing multiple jobs to execute concurrently.

---

# 8.1 Execution Pipeline

Every execution follows the same sequence.

```text
                Trigger Due
                     │
                     ▼
                Scheduler
                     │
                     ▼
               Job Service
                     │
                     ▼
                  Broker
                     │
                     ▼
                Redis Queue
                     │
                     ▼
                 Worker Pool
                     │
                     ▼
             Execution Engine
                     │
                     ▼
            Execution Context
                     │
                     ▼
             Action Registry
                     │
                     ▼
             Action Executor
                     │
                     ▼
             Execution Result
                     │
                     ▼
               PostgreSQL
                     │
                     ▼
             Notification Service
```

Every component has a single responsibility.

No component skips another component.

---

# 8.2 Trigger Detection

The Scheduler continuously checks all active triggers.

For each trigger it determines:

* Is the trigger active?
* Is the scheduled time due?
* Is the task paused?
* Is another execution already running?
* Is the trigger valid?

Only valid triggers are forwarded for execution.

The Scheduler never performs the execution itself.

---

# 8.3 Job Creation

The Job Service converts a scheduled trigger into an executable job.

During this stage it performs:

* Task lookup
* Trigger lookup
* Queue lookup
* Action lookup
* Retry policy initialization
* Execution ID generation

The final result is a complete Job object that contains all information required by the Worker.

---

# 8.4 Queue Publication

The Broker serializes the Job and publishes it to Redis.

Once published, the Scheduler considers its work complete.

The Broker is responsible only for reliable job publication.

It never performs execution.

---

# 8.5 Queue Processing

Workers continuously monitor Redis for pending jobs.

When a Worker becomes available:

1. Retrieve one job.
2. Deserialize the payload.
3. Build the Execution Context.
4. Start execution.

Each Worker processes one job at a time.

Multiple Workers may execute simultaneously.

---

# 8.6 Execution Context

Before execution begins, the Worker creates an Execution Context.

The Execution Context carries all information required during execution.

Typical contents include:

* Execution ID
* Task ID
* User ID
* Queue ID
* Retry Count
* Start Time
* Logger
* Cancellation Context

Rather than passing many individual parameters through the execution pipeline, every component receives a single Execution Context.

This improves readability, maintainability, and testability.

---

# 8.7 Execution Engine

The Execution Engine coordinates the complete lifecycle of a job.

Responsibilities include:

* Validate execution state
* Execute workflow actions
* Handle action ordering
* Apply timeouts
* Record execution duration
* Capture errors
* Determine final execution status

The Execution Engine does not implement action-specific behavior.

Instead, it delegates work to Action Executors.

---

# 8.8 Action Resolution

For each action in a workflow, the Execution Engine requests the appropriate executor from the Action Registry.

Example:

```text
EMAIL

↓

Email Executor
```

```text
HTTP

↓

HTTP Executor
```

```text
SHELL

↓

Shell Executor
```

This mechanism allows new action types to be introduced without modifying the Execution Engine.

---

# 8.9 Action Execution

Each Action Executor performs one business operation.

Version 1 supports:

* Reminder
* Email
* HTTP Request
* Shell Command

Each executor follows the same interface, enabling the Execution Engine to execute any supported action consistently.

---

# 8.10 Execution Result

When execution completes, the Worker constructs an execution result.

Typical information includes:

* Execution ID
* Status
* Start Time
* End Time
* Duration
* Retry Count
* Error Message
* Output

The result is stored permanently in PostgreSQL.

---

# 8.11 Retry Handling

If execution fails and retries remain available:

```text
Execution Failed

↓

Retry Policy

↓

Delay

↓

Redis Queue

↓

Worker

↓

Retry Execution
```

Retries continue until:

* Execution succeeds.
* Maximum retry count is reached.
* Execution is cancelled.

---

# 8.12 Failure Isolation

A failed execution must never interrupt other jobs.

Each Worker operates independently.

If one execution fails:

* The failure is recorded.
* Retry logic is applied if configured.
* Other Workers continue processing normally.

This isolation improves overall system reliability.

---

# 8.13 Execution Completion

After successful execution:

1. Execution record updated.
2. Task statistics updated.
3. Notifications generated (if configured).
4. Worker requests the next job.

The Worker immediately returns to the queue for additional work.

---

# 8.14 Summary

The execution architecture is the heart of STAQ.

By separating scheduling, queueing, execution, and action processing into independent components, the platform achieves:

* High concurrency
* Reliable background execution
* Easy extensibility
* Fault isolation
* Efficient resource utilization
* Clear separation of responsibilities

This architecture provides the foundation for future capabilities such as workflow dependencies, distributed workers, AI-assisted task planning, and advanced automation without requiring significant structural changes.

---

# 9. Scheduler Architecture

The Scheduler is responsible for determining **when** a task should be executed.

It is the first component in the execution pipeline and continuously monitors active triggers to identify tasks whose scheduled execution time has arrived.

The Scheduler **never executes tasks directly**. Its only responsibility is to detect due triggers and create executable jobs.

---

# 9.1 Responsibilities

The Scheduler performs the following responsibilities:

* Load active triggers
* Calculate next execution times
* Detect due triggers
* Skip inactive or paused tasks
* Prevent duplicate scheduling
* Submit execution requests to the Job Service
* Update trigger execution metadata

The Scheduler does **not**:

* Execute business actions
* Send emails
* Call APIs
* Run shell commands
* Store execution results

These responsibilities belong to downstream components.

---

# 9.2 Scheduler Lifecycle

The Scheduler runs continuously in the background.

```text
Start Scheduler
       │
       ▼
Load Active Triggers
       │
       ▼
Check Current Time
       │
       ▼
Find Due Triggers
       │
       ▼
Create Execution Jobs
       │
       ▼
Send Jobs to Broker
       │
       ▼
Sleep Until Next Tick
       │
       └───────────────┐
                       ▼
               Repeat Forever
```

The Scheduler repeats this cycle at a configurable interval.

---

# 9.3 Scheduler Tick

Rather than checking continuously, the Scheduler wakes up periodically.

Example:

```text
Tick Interval = 1 second
```

Every tick:

1. Read current time.
2. Find triggers due before or at the current time.
3. Submit jobs.
4. Calculate the next execution time.
5. Return to sleep.

The tick interval should be configurable through the application configuration.

---

# 9.4 Trigger Evaluation

Each active trigger is evaluated independently.

The Scheduler verifies:

* Trigger is active.
* Parent task is active.
* Execution time has arrived.
* Trigger has not already been processed.
* Task is not currently paused.

Only triggers satisfying all conditions are eligible for execution.

---

# 9.5 Preventing Duplicate Execution

A trigger must never generate duplicate jobs for the same scheduled occurrence.

To prevent duplicates, the Scheduler:

* Marks a trigger as being processed.
* Creates an execution record.
* Updates the trigger's next run time before scheduling the next occurrence.

This ensures that even if the Scheduler restarts, the same occurrence is not processed twice.

---

# 9.6 Trigger Types

Version 1 supports multiple trigger types.

Supported trigger schedules include:

* One-Time
* Delayed
* Daily
* Weekly
* Monthly
* Yearly
* Cron Expression

Each trigger calculates its own next execution time according to its schedule.

---

# 9.7 Next Run Calculation

After scheduling a trigger, the Scheduler calculates its next execution time.

Examples:

Daily Trigger

```text
Today 09:00

↓

Tomorrow 09:00
```

Weekly Trigger

```text
Monday

↓

Next Monday
```

Cron Trigger

```text
*/15 * * * *

↓

Next matching cron time
```

If a trigger has no future executions (for example, a completed one-time trigger), it is marked as finished or disabled.

---

# 9.8 Scheduler Failure Recovery

The Scheduler must tolerate failures gracefully.

Possible failures include:

* Database temporarily unavailable
* Redis unavailable
* Job creation failure
* Unexpected runtime error

If an error occurs:

* Log the error.
* Continue processing remaining triggers.
* Retry during the next scheduling cycle if appropriate.

A failure affecting one trigger must not stop the Scheduler.

---

# 9.9 Time Zone Support

Every trigger stores its associated time zone.

The Scheduler always evaluates triggers using their configured time zone.

Examples:

* UTC
* Asia/Dhaka
* Europe/London

This ensures users in different regions receive correct execution times.

---

# 9.10 Scheduler Configuration

The Scheduler behavior is configurable.

Typical configuration options include:

* Tick Interval
* Maximum Triggers Per Tick
* Default Time Zone
* Maximum Concurrent Scheduling Operations
* Scheduler Startup Delay

Centralizing configuration simplifies tuning and deployment.

---

# 9.11 Summary

The Scheduler is responsible only for deciding **when** work should begin.

By separating scheduling from execution, STAQ maintains a clean architecture where timing logic remains independent of business logic.

This separation improves reliability, scalability, and future extensibility while allowing multiple Workers to execute jobs concurrently without affecting scheduling accuracy.

---

# 10. Job Service Architecture

The Job Service is responsible for transforming a scheduled trigger into an executable job.

It acts as the bridge between the Scheduler and the Broker by collecting all required information, validating it, and constructing a complete job payload for background execution.

The Job Service does **not** execute tasks. Its only responsibility is to prepare jobs for processing.

---

# 10.1 Responsibilities

The Job Service performs the following responsibilities:

* Receive execution requests from the Scheduler
* Load task information
* Load trigger information
* Load workflow actions
* Validate task configuration
* Create an execution record
* Generate an execution identifier
* Build the job payload
* Send the job to the Broker

The Job Service never:

* Executes actions
* Sends emails
* Runs shell commands
* Calls external APIs
* Processes Redis messages

---

# 10.2 Job Creation Flow

Every execution request follows the same sequence.

```text
Scheduler
      │
      ▼
Load Task
      │
      ▼
Load Trigger
      │
      ▼
Load Actions
      │
      ▼
Validate Configuration
      │
      ▼
Create Execution Record
      │
      ▼
Build Job Payload
      │
      ▼
Broker
```

Each step must complete successfully before the next begins.

---

# 10.3 Loading Task Information

The Job Service first retrieves the associated task.

Information loaded includes:

* Task ID
* Task Name
* Owner
* Queue
* Status
* Timeout
* Retry Policy

If the task is inactive or deleted, job creation stops immediately.

---

# 10.4 Loading Trigger Information

The Job Service retrieves the trigger that initiated execution.

Examples of trigger information include:

* Trigger ID
* Trigger Type
* Schedule
* Time Zone
* Current Run Time
* Next Run Time

This information becomes part of the execution context.

---

# 10.5 Loading Workflow Actions

Every task contains one or more actions.

The Job Service retrieves the complete ordered action list.

Example:

```text
Task

↓

Action 1 → Reminder

↓

Action 2 → HTTP Request

↓

Action 3 → Email
```

The action order is preserved exactly as configured.

---

# 10.6 Validation

Before creating a job, the Job Service validates:

* Task exists
* Task is active
* Trigger exists
* Trigger is active
* Queue exists
* At least one action exists
* Retry policy is valid
* Timeout configuration is valid

If any validation fails, no job is created.

The Scheduler records the failure and continues processing other triggers.

---

# 10.7 Execution Record Creation

Before publishing a job, the Job Service creates a new execution record in PostgreSQL.

Initial values include:

* Execution ID
* Task ID
* Trigger ID
* Status = Pending
* Retry Count = 0
* Created Time

This allows execution history to exist even before a Worker begins processing.

---

# 10.8 Job Payload

The Job Payload contains everything required by the Worker.

Typical fields include:

* Execution ID
* Task ID
* User ID
* Queue ID
* Trigger ID
* Ordered Actions
* Retry Policy
* Timeout
* Scheduled Time

The payload is self-contained so that Workers do not need to perform additional database lookups during normal execution whenever possible.

---

# 10.9 Publishing the Job

After the payload has been prepared, the Job Service forwards it to the Broker.

At this point:

* The Scheduler has completed its work.
* The Job Service has completed its work.
* Responsibility transfers entirely to the Broker.

---

# 10.10 Error Handling

Possible errors include:

* Missing task
* Missing trigger
* Invalid workflow
* Database failure
* Execution record creation failure

If an error occurs:

* Record the failure.
* Log sufficient diagnostic information.
* Return the error to the Scheduler.
* Continue processing other execution requests.

Errors affecting one job must not interrupt processing of unrelated jobs.

---

# 10.11 Design Principles

The Job Service follows several important principles.

### Single Responsibility

Its only purpose is to prepare executable jobs.

### Immutable Job Payload

Once published, the job payload should not change.

Workers should treat the payload as read-only.

### Separation of Concerns

Scheduling, job preparation, queue publication, and execution remain independent responsibilities.

---

# 10.12 Summary

The Job Service transforms scheduling decisions into executable work.

By gathering all required task information, validating configuration, creating execution records, and constructing immutable job payloads, it ensures that Workers receive complete and reliable execution requests while remaining independent of scheduling logic.

---

# 11. Broker Architecture

The Broker is responsible for delivering executable jobs from the Job Service to the background processing system.

It provides an abstraction over the underlying message queue, allowing the rest of the application to remain independent of the queue technology being used.

In Version 1, Redis serves as the message broker implementation.

---

# 11.1 Responsibilities

The Broker performs the following responsibilities:

* Receive job payloads from the Job Service
* Serialize job payloads
* Publish jobs to the appropriate queue
* Report publication success or failure
* Hide queue implementation details from the rest of the application

The Broker never:

* Executes jobs
* Validates business logic
* Stores execution history
* Performs scheduling

Its only responsibility is reliable message delivery.

---

# 11.2 Broker Flow

```text
Job Service
      │
      ▼
Broker Interface
      │
      ▼
Redis Broker
      │
      ▼
Redis Queue
```

The Job Service depends only on the Broker interface.

It has no knowledge of Redis-specific commands or implementation details.

---

# 11.3 Message Serialization

Before publication, each Job Payload is serialized into a transport format.

Version 1 uses JSON serialization.

Example fields include:

* Execution ID
* Task ID
* User ID
* Trigger ID
* Queue ID
* Actions
* Retry Policy
* Timeout
* Scheduled Time

The serialized payload is treated as immutable after publication.

---

# 11.4 Queue Selection

Each job belongs to a logical queue.

Examples include:

* default
* email
* http
* shell

The Broker determines the correct destination queue using the queue information contained in the Job Payload.

This allows different categories of work to be processed independently.

---

# 11.5 Publication

Publishing consists of:

1. Serialize the Job Payload.
2. Connect to Redis.
3. Push the message to the target queue.
4. Confirm successful publication.
5. Return the result to the Job Service.

If publication fails, the Job Service is notified immediately.

---

# 11.6 Failure Handling

Possible publication failures include:

* Redis unavailable
* Network interruption
* Serialization failure
* Queue not found

If publication fails:

* The error is logged.
* The Scheduler is informed through the Job Service.
* No Worker receives the job.
* The execution remains in a pending or failed scheduling state according to the retry policy.

The Broker never attempts to execute or modify jobs.

---

# 11.7 Broker Interface

The Broker exposes a technology-independent contract.

Typical operations include:

* Publish Job
* Check Broker Health
* Close Connection

Additional capabilities may be introduced in future versions without affecting the Scheduler or Job Service.

---

# 11.8 Redis Broker

The Redis Broker is the Version 1 implementation of the Broker interface.

Responsibilities include:

* Managing Redis connections
* Publishing messages
* Handling Redis-specific errors
* Reconnecting when necessary

All Redis-specific logic remains isolated within this implementation.

---

# 11.9 Design Principles

The Broker follows several architectural principles.

### Abstraction

Business logic depends on an interface rather than a concrete queue technology.

### Reliability

A job is considered queued only after successful publication.

### Transparency

The Scheduler and Job Service remain unaware of the underlying messaging implementation.

### Extensibility

Future queue technologies can be introduced by implementing the same Broker interface.

---

# 11.10 Summary

The Broker separates job creation from job delivery.

By abstracting the messaging system behind a common interface, STAQ remains flexible, testable, and ready for future expansion while maintaining a simple Redis-based implementation in Version 1.

---

# 12. Worker Pool Architecture

The Worker Pool is responsible for executing jobs retrieved from the queue.

Workers continuously listen for new jobs, process them independently, execute all configured actions, record execution results, and return to waiting for the next job.

The Worker Pool is designed for high concurrency while ensuring that each individual job is executed exactly once by a single Worker.

---

# 12.1 Responsibilities

The Worker Pool performs the following responsibilities:

* Listen for queued jobs
* Retrieve job payloads
* Build the Execution Context
* Execute jobs through the Execution Engine
* Handle retries
* Apply execution timeouts
* Record execution status
* Return to waiting for additional jobs

Workers never:

* Schedule jobs
* Create jobs
* Modify triggers
* Manage users
* Perform authentication

Their only responsibility is reliable background execution.

---

# 12.2 Worker Lifecycle

Each Worker follows the same execution cycle.

```text
Start Worker
      │
      ▼
Wait for Job
      │
      ▼
Receive Job
      │
      ▼
Build Execution Context
      │
      ▼
Execute Job
      │
      ▼
Store Result
      │
      ▼
Request Next Job
      │
      └──────────────┐
                     ▼
                 Repeat
```

Workers remain active until the application shuts down.

---

# 12.3 Worker Pool

Instead of one Worker, STAQ runs multiple Workers concurrently.

Example:

```text
Redis Queue
      │
      ├─────────────┬─────────────┬─────────────┐
      ▼             ▼             ▼             ▼
 Worker 1      Worker 2      Worker 3      Worker N
```

Each Worker processes one job at a time.

This allows many jobs to execute simultaneously while keeping each individual Worker simple.

---

# 12.4 Job Retrieval

Workers continuously request jobs from the Broker.

When a job is received:

1. Deserialize the Job Payload.
2. Validate the payload.
3. Build the Execution Context.
4. Begin execution.

If no jobs are available, the Worker waits without consuming excessive CPU resources.

---

# 12.5 Execution Context

Every job receives a new Execution Context.

The Execution Context contains runtime information that may change during execution.

Typical fields include:

* Execution ID
* Task ID
* User ID
* Queue ID
* Retry Count
* Current Action Index
* Start Time
* Deadline
* Logger
* Cancellation Context

The Execution Context is mutable and exists only for the lifetime of a single execution.

---

# 12.6 Execution

Workers delegate execution to the Execution Engine.

```text
Worker

↓

Execution Engine

↓

Action Registry

↓

Action Executor
```

Workers never execute business actions directly.

---

# 12.7 Timeout Handling

Each task may define an execution timeout.

Example:

```text
Task Timeout

↓

5 Minutes
```

If the timeout expires:

* Execution is cancelled.
* Current action stops if possible.
* Execution status becomes "Timed Out".
* Retry policy is evaluated.

This prevents indefinitely running jobs.

---

# 12.8 Retry Processing

When execution fails:

```text
Failure

↓

Retry Policy

↓

Retry Remaining?

↓

Yes

↓

Return Job to Queue

↓

Worker Executes Again
```

If no retries remain:

* Execution becomes Failed.
* Failure is recorded.
* Job moves to the Dead Letter Queue (if enabled).

---

# 12.9 Graceful Shutdown

Workers support graceful shutdown.

Shutdown sequence:

1. Stop accepting new jobs.
2. Finish the current execution.
3. Store execution result.
4. Close resources.
5. Exit safely.

This prevents partially processed jobs during deployments or application restarts.

---

# 12.10 Concurrency

Workers execute independently.

Example:

```text
Worker 1

↓

Email Task

----------------------

Worker 2

↓

HTTP Task

----------------------

Worker 3

↓

Shell Task
```

One slow task must never block unrelated tasks.

The number of Workers is configurable.

---

# 12.11 Error Isolation

Errors remain isolated within a single execution.

If one Worker fails:

* Other Workers continue processing.
* Remaining queued jobs remain unaffected.
* Scheduler continues operating normally.

This isolation improves overall reliability.

---

# 12.12 Logging

Workers generate structured logs during execution.

Examples include:

* Job received
* Execution started
* Action completed
* Retry scheduled
* Timeout occurred
* Execution finished

These logs support debugging and operational monitoring.

---

# 12.13 Metrics

Each Worker should expose operational metrics.

Examples include:

* Active Workers
* Idle Workers
* Running Jobs
* Successful Executions
* Failed Executions
* Retry Count
* Average Execution Time

These metrics help administrators monitor system health and performance.

---

# 12.14 Summary

The Worker Pool is responsible for all background execution within STAQ.

By using multiple independent Workers, immutable Job Payloads, mutable Execution Contexts, configurable retries, and graceful shutdown, the Worker Pool provides reliable, scalable, and fault-tolerant execution while remaining independent of scheduling and job creation.

---

# 13. Action Framework

The Action Framework is responsible for executing the business operations defined within a task.

Each task consists of one or more actions executed in a predefined order.

The framework is designed to be extensible, allowing new action types to be introduced without modifying the Execution Engine or Worker Pool.

---

# 13.1 Overview

Every task contains a workflow.

Example:

```text
Task: Daily Backup

↓

Reminder

↓

Shell Command

↓

HTTP Request

↓

Email

↓

Completed
```

The Execution Engine executes each action sequentially.

---

# 13.2 Supported Actions (Version 1)

STAQ Version 1 supports the following action types:

| Action   | Description                               |
| -------- | ----------------------------------------- |
| Reminder | Create an in-app reminder or notification |
| Email    | Send an email using SMTP                  |
| HTTP     | Send an HTTP request                      |
| Shell    | Execute an approved shell command         |

Future versions can introduce additional action types without changing the framework.

---

# 13.3 Action Lifecycle

Every action follows the same execution lifecycle.

```text
Execution Engine
        │
        ▼
Action Registry
        │
        ▼
Action Executor
        │
        ▼
Validate Configuration
        │
        ▼
Execute Action
        │
        ▼
Return Result
```

The Worker interacts only with the Execution Engine.

The Execution Engine interacts only with the Action Registry.

The Registry selects the appropriate Action Executor.

---

# 13.4 Action Interface

Every Action Executor implements the same interface.

Typical responsibilities include:

* Validate configuration
* Execute business logic
* Return execution result
* Return execution error

Because all executors follow a common contract, the Execution Engine treats every action uniformly.

This eliminates the need for action-specific logic inside the Worker or Scheduler.

---

# 13.5 Action Registry

The Action Registry maps an action type to its corresponding executor.

Example mapping:

```text
REMINDER → Reminder Executor

EMAIL → Email Executor

HTTP → HTTP Executor

SHELL → Shell Executor
```

The Registry is initialized during application startup.

When execution begins, the Execution Engine requests the appropriate executor based on the action type.

---

# 13.6 Reminder Executor

Responsibilities:

* Validate reminder configuration
* Create in-app notification
* Record execution result

The Reminder Executor does not interact with external services.

---

# 13.7 Email Executor

Responsibilities:

* Validate recipient
* Validate SMTP configuration
* Build email message
* Send email
* Record delivery result

Failures are returned to the Execution Engine for retry evaluation.

---

# 13.8 HTTP Executor

Responsibilities:

* Validate URL
* Validate HTTP method
* Attach headers
* Attach request body
* Send request
* Capture response
* Record result

Supported methods:

* GET
* POST
* PUT
* PATCH
* DELETE

HTTP timeouts are configurable.

---

# 13.9 Shell Executor

Responsibilities:

* Validate command
* Validate execution environment
* Execute approved command
* Capture stdout
* Capture stderr
* Return exit code

Only approved commands should be executed.

Direct execution of arbitrary user-provided commands is prohibited for security reasons.

---

# 13.10 Action Ordering

Actions execute sequentially.

Example:

```text
Action 1

↓

Action 2

↓

Action 3

↓

Action 4
```

The next action begins only after the previous action completes.

This guarantees deterministic workflow execution.

---

# 13.11 Action Failure

If an action fails:

```text
Action Failed

↓

Execution Engine

↓

Retry Policy

↓

Retry?

↓

Yes → Retry

No → Execution Failed
```

Subsequent actions are not executed unless the workflow eventually succeeds.

This prevents inconsistent workflow states.

---

# 13.12 Action Result

Every executor returns a standardized execution result.

Typical information includes:

* Status
* Start Time
* End Time
* Duration
* Output
* Error Message

The Execution Engine aggregates these results into the overall execution record.

---

# 13.13 Future Action Types

The framework is designed to support additional action types without architectural changes.

Examples include:

* Database Query
* File Operations
* GitHub Integration
* Browser Automation
* Calendar Events
* AI Processing
* Plugin Actions

Adding a new action requires:

1. Implement the Action interface.
2. Register the executor.
3. Define the action configuration.

No modifications are required to the Worker Pool or Execution Engine.

---

# 13.14 Design Principles

The Action Framework follows these principles:

### Extensibility

New action types can be added independently.

### Single Responsibility

Each executor performs one specific type of work.

### Loose Coupling

Executors are selected dynamically through the Action Registry.

### Reusability

Action Executors can be reused across multiple workflows.

### Consistency

Every action follows the same lifecycle and returns the same result structure.

---

# 13.15 Summary

The Action Framework provides a flexible and extensible mechanism for executing workflow actions.

By separating action-specific logic into dedicated executors and using a centralized Action Registry, STAQ remains easy to extend, test, and maintain while supporting a growing library of automation capabilities.
