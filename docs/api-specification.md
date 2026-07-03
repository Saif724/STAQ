# API Specification

## STAQ
### Intelligent Task Scheduling & Workflow Automation Platform

Version: 1.0

Prepared By:
Ahsan Ahmed Saif

---

# 1. Introduction

## 1.1 Purpose

This document defines all REST APIs provided by the STAQ backend.

It serves as the reference for frontend development, backend implementation, and API testing.

Every endpoint includes:

- HTTP Method
- Endpoint
- Authentication Requirement
- Request Body
- Response Body
- Status Codes
- Validation Rules

---

## 1.2 Base URL

Development

```
http://localhost:8080/api/v1
```

Production

```
https://your-domain.com/api/v1
```

---

## 1.3 Authentication

Protected endpoints require a valid JWT access token.

The token must be sent using the Authorization header.

```
Authorization: Bearer <access_token>
```

Endpoints related to registration, login, and email verification do not require authentication.

---

## 1.4 Response Format

### Success Response

```json
{
    "success": true,
    "message": "Operation completed successfully.",
    "data": {}
}
```

---

### Error Response

```json
{
    "success": false,
    "message": "Validation failed.",
    "errors": {
        "email": "Email already exists."
    }
}
```

---

# 2. Authentication APIs

Authentication APIs manage user accounts.

---

## 2.1 Register

### Endpoint

```
POST /auth/register
```

### Authentication

Not Required

### Request

```json
{
    "full_name": "Ahsan Ahmed Saif",
    "email": "example@email.com",
    "password": "StrongPassword123"
}
```

### Success Response

```json
{
    "success": true,
    "message": "Registration successful. Please verify your email."
}
```

### Validation

- Full name is required.
- Email must be unique.
- Password must meet the minimum security requirements.

### Status Codes

- 201 Created
- 400 Bad Request
- 409 Conflict

---

## 2.2 Login

### Endpoint

```
POST /auth/login
```

### Authentication

Not Required

### Request

```json
{
    "email": "example@email.com",
    "password": "StrongPassword123"
}
```

### Success Response

```json
{
    "success": true,
    "data": {
        "access_token": "...",
        "expires_in": 3600
    }
}
```

### Status Codes

- 200 OK
- 400 Bad Request
- 401 Unauthorized

---

## 2.3 Verify Email

### Endpoint

```
GET /auth/verify-email?token=xxxxx
```

### Authentication

Not Required

### Success Response

```json
{
    "success": true,
    "message": "Email verified successfully."
}
```

### Status Codes

- 200 OK
- 400 Bad Request
- 404 Not Found

---

## 2.4 Resend Verification Email

### Endpoint

```
POST /auth/resend-verification
```

### Authentication

Not Required

### Request

```json
{
    "email": "example@email.com"
}
```

### Status Codes

- 200 OK
- 400 Bad Request

---

## 2.5 Get Current User

### Endpoint

```
GET /auth/me
```

### Authentication

Required

### Success Response

```json
{
    "success": true,
    "data": {
        "id": "...",
        "full_name": "...",
        "email": "...",
        "email_verified": true
    }
}
```

---

## 2.6 Logout

### Endpoint

```
POST /auth/logout
```

### Authentication

Required

### Success Response

```json
{
    "success": true,
    "message": "Logged out successfully."
}
```

---

# 3. Task APIs

Task APIs allow authenticated users to create, manage, and organize automation tasks.

All Task APIs require authentication.

---

## 3.1 Create Task

### Endpoint

```
POST /api/v1/tasks
```

### Authentication

Required

### Request

```json
{
    "name": "Daily Study Reminder",
    "description": "Practice DSA every evening",
    "queue_id": "queue_uuid",
    "status": "ACTIVE",
    "timeout_seconds": 300,
    "max_retries": 3
}
```

### Success Response

```json
{
    "success": true,
    "message": "Task created successfully.",
    "data": {
        "id": "task_uuid"
    }
}
```

### Validation

- Name is required.
- Name must not exceed 150 characters.
- Queue must exist.
- Timeout must be greater than zero.
- Retry count cannot be negative.

### Status Codes

- 201 Created
- 400 Bad Request
- 401 Unauthorized
- 404 Not Found

---

## 3.2 Get All Tasks

### Endpoint

```
GET /api/v1/tasks
```

### Authentication

Required

### Query Parameters

| Parameter | Description |
|-----------|-------------|
| page | Page number |
| limit | Number of records |
| status | Filter by status |
| queue | Filter by queue |
| search | Search by task name |

Example

```
GET /api/v1/tasks?page=1&limit=20&status=ACTIVE
```

### Success Response

```json
{
    "success": true,
    "data": [
        {
            "id": "...",
            "name": "Daily Backup",
            "status": "ACTIVE"
        }
    ]
}
```

### Status Codes

- 200 OK
- 401 Unauthorized

---

## 3.3 Get Task

### Endpoint

```
GET /api/v1/tasks/{id}
```

### Authentication

Required

### Success Response

```json
{
    "success": true,
    "data": {
        "id": "...",
        "name": "...",
        "description": "...",
        "status": "ACTIVE"
    }
}
```

### Status Codes

- 200 OK
- 401 Unauthorized
- 404 Not Found

---

## 3.4 Update Task

### Endpoint

```
PUT /api/v1/tasks/{id}
```

### Authentication

Required

### Request

```json
{
    "name": "Updated Task",
    "description": "Updated Description",
    "queue_id": "queue_uuid",
    "status": "ACTIVE",
    "timeout_seconds": 600,
    "max_retries": 5
}
```

### Success Response

```json
{
    "success": true,
    "message": "Task updated successfully."
}
```

### Status Codes

- 200 OK
- 400 Bad Request
- 401 Unauthorized
- 404 Not Found

---

## 3.5 Delete Task

### Endpoint

```
DELETE /api/v1/tasks/{id}
```

### Authentication

Required

### Success Response

```json
{
    "success": true,
    "message": "Task deleted successfully."
}
```

### Status Codes

- 200 OK
- 401 Unauthorized
- 404 Not Found

---

## 3.6 Pause Task

### Endpoint

```
PATCH /api/v1/tasks/{id}/pause
```

### Authentication

Required

### Success Response

```json
{
    "success": true,
    "message": "Task paused successfully."
}
```

### Status Codes

- 200 OK
- 401 Unauthorized
- 404 Not Found

---

## 3.7 Resume Task

### Endpoint

```
PATCH /api/v1/tasks/{id}/resume
```

### Authentication

Required

### Success Response

```json
{
    "success": true,
    "message": "Task resumed successfully."
}
```

### Status Codes

- 200 OK
- 401 Unauthorized
- 404 Not Found

---

## 3.8 Archive Task

### Endpoint

```
PATCH /api/v1/tasks/{id}/archive
```

### Authentication

Required

### Success Response

```json
{
    "success": true,
    "message": "Task archived successfully."
}
```

### Status Codes

- 200 OK
- 401 Unauthorized
- 404 Not Found

---

# 4. Trigger APIs

Triggers define when a task should execute.

All Trigger APIs require authentication.

---

## 4.1 Create Trigger

### Endpoint

```
POST /api/v1/tasks/{taskId}/triggers
```

### Request

```json
{
    "trigger_type": "CRON",
    "cron_expression": "0 8 * * *",
    "timezone": "Asia/Dhaka"
}
```

### Success Response

```json
{
    "success": true,
    "message": "Trigger created successfully."
}
```

---

## 4.2 Get Triggers

### Endpoint

```
GET /api/v1/tasks/{taskId}/triggers
```

Returns every trigger belonging to a task.

---

## 4.3 Get Trigger

### Endpoint

```
GET /api/v1/triggers/{id}
```

Returns a specific trigger.

---

## 4.4 Update Trigger

### Endpoint

```
PUT /api/v1/triggers/{id}
```

Updates the scheduling configuration.

---

## 4.5 Delete Trigger

### Endpoint

```
DELETE /api/v1/triggers/{id}
```

Deletes the trigger.

---

# 5. Action APIs

Actions define what a task performs after it is triggered.

---

## 5.1 Create Action

### Endpoint

```
POST /api/v1/tasks/{taskId}/actions
```

### Request

```json
{
    "action_type": "HTTP",
    "execution_order": 1,
    "configuration": {
        "method": "POST",
        "url": "https://example.com/api"
    }
}
```

---

## 5.2 Get Actions

### Endpoint

```
GET /api/v1/tasks/{taskId}/actions
```

Returns all actions ordered by `execution_order`.

---

## 5.3 Get Action

### Endpoint

```
GET /api/v1/actions/{id}
```

Returns a single action.

---

## 5.4 Update Action

### Endpoint

```
PUT /api/v1/actions/{id}
```

Updates the action configuration.

---

## 5.5 Delete Action

### Endpoint

```
DELETE /api/v1/actions/{id}
```

Deletes the action.

---

## 5.6 Reorder Actions

### Endpoint

```
PATCH /api/v1/tasks/{taskId}/actions/reorder
```

### Request

```json
{
    "actions": [
        {
            "id": "uuid1",
            "execution_order": 1
        },
        {
            "id": "uuid2",
            "execution_order": 2
        }
    ]
}
```

Allows drag-and-drop ordering in the frontend.

---

# 6. Execution APIs

Execution APIs provide execution history and monitoring.

Executions cannot be edited or deleted.

---

## 6.1 Get Executions

### Endpoint

```
GET /api/v1/executions
```

Supports pagination and filtering.

Optional filters:

- task_id
- status
- start_date
- end_date

---

## 6.2 Get Execution

### Endpoint

```
GET /api/v1/executions/{id}
```

Returns complete execution details.

---

## 6.3 Get Execution Logs

### Endpoint

```
GET /api/v1/executions/{id}/logs
```

Returns logs generated during execution.

---

## 6.4 Retry Execution

### Endpoint

```
POST /api/v1/executions/{id}/retry
```

Creates a new execution using the same task configuration.

---

# 7. Queue APIs

Queue APIs manage execution queues.

---

## 7.1 Get Queues

### Endpoint

```
GET /api/v1/queues
```

Returns all queues.

---

## 7.2 Create Queue

### Endpoint

```
POST /api/v1/queues
```

Creates a queue.

---

## 7.3 Update Queue

### Endpoint

```
PUT /api/v1/queues/{id}
```

Updates queue information.

---

## 7.4 Delete Queue

### Endpoint

```
DELETE /api/v1/queues/{id}
```

Queues with assigned tasks cannot be deleted.

Inactive queues should be archived instead.

---

# 8. User Settings APIs

Stores user preferences.

---

## 8.1 Get User Settings

### Endpoint

```
GET /api/v1/settings
```

Returns the authenticated user's settings.

---

## 8.2 Update User Settings

### Endpoint

```
PUT /api/v1/settings
```

Example request:

```json
{
    "timezone": "Asia/Dhaka",
    "default_queue_id": "queue_uuid"
}
```

---

# 9. HTTP Status Codes

| Status Code | Meaning |
|-------------|---------|
| 200 | OK |
| 201 | Created |
| 204 | No Content |
| 400 | Bad Request |
| 401 | Unauthorized |
| 403 | Forbidden |
| 404 | Not Found |
| 409 | Conflict |
| 422 | Validation Error |
| 500 | Internal Server Error |

---

# 10. API Versioning

All endpoints are versioned.

Current version:

```
/api/v1
```

Future breaking changes will be introduced under:

```
/api/v2
```

allowing older clients to continue functioning without modification.

---

# 11. API Security

The API follows these security practices:

- JWT Authentication
- Password hashing using bcrypt
- Email verification before account activation
- Request validation
- SQL injection prevention through parameterized queries
- HTTPS in production
- CORS configuration
- Secure HTTP headers
- Rate limiting (future enhancement)

---

# 12. Summary

The STAQ REST API is organized around RESTful principles.

Resources are grouped logically into Authentication, Tasks, Triggers, Actions, Executions, Queues, and User Settings.

The API has been designed to remain stable, extensible, and versioned, providing a solid foundation for frontend integration and future feature expansion.
