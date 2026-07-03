# Deployment Guide

## STAQ
### Intelligent Task Scheduling & Workflow Automation Platform

Version: 1.0

Prepared By:
Ahsan Ahmed Saif

---

# 1. Introduction

## 1.1 Purpose

This document describes how to configure, run, and deploy STAQ.

It includes the required software, environment variables, local development setup, Docker deployment, and production deployment recommendations.

---

# 2. System Requirements

## Backend

- Go 1.25 or later
- PostgreSQL 16+
- Redis 8+
- Docker (optional)
- Docker Compose (optional)

---

## Frontend

- Node.js 22+
- npm or pnpm

---

## Development Tools

- Git
- Visual Studio Code
- Postman or Bruno
- Make (optional)

---

# 3. Project Structure

```text
STAQ/
│
├── backend/
│   ├── cmd/
│   ├── configs/
│   ├── internal/
│   ├── migrations/
│   ├── pkg/
│   ├── scripts/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   ├── go.mod
│   └── .env
│
├── frontend/
│   ├── src/
│   ├── public/
│   ├── package.json
│   └── .env
│
└── docs/
```

---

# 4. Environment Variables

## Backend

Example `.env`

```env
APP_ENV=development
PORT=8080

DB_HOST=
DB_PORT=5432
DB_USER=
DB_PASSWORD=
DB_NAME=
DATABASE_URL=

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

JWT_SECRET=

SMTP_HOST=
SMTP_PORT=587
SMTP_USERNAME=
SMTP_PASSWORD=
SMTP_FROM=

FRONTEND_URL=http://localhost:5173
BACKEND_URL=http://localhost:8080
```

---

## Frontend

Example `.env`

```env
VITE_API_URL=http://localhost:8080/api/v1
```

---

# 5. Local Development

## Step 1

Clone the repository.

```bash
git clone <repository-url>
```

---

## Step 2

Backend

```bash
cd backend
```

Install dependencies.

```bash
go mod tidy
```

---

## Step 3

Frontend

```bash
cd frontend
npm install
```

---

## Step 4

Configure the environment variables.

---

## Step 5

Start PostgreSQL.

---

## Step 6

Start Redis.

---

## Step 7

Run database migrations.

```bash
make migrate-up
```

or

```bash
golang-migrate ...
```

---

## Step 8

Run the backend.

```bash
go run cmd/api/main.go
```

---

## Step 9

Run the frontend.

```bash
npm run dev
```

---

# 6. Docker Deployment

Build the backend image.

```bash
docker build -t staq-backend .
```

Run the container.

```bash
docker run ...
```

Alternatively,

```bash
docker compose up --build
```

---

# 7. Production Deployment

Recommended production services.

Backend

- Railway
- Render
- DigitalOcean
- VPS

Database

- Neon PostgreSQL

Redis

- Upstash Redis

Frontend

- Vercel
- Netlify

---

# 8. Health Check

Health endpoint.

```
GET /health
```

Example response.

```json
{
    "status":"ok"
}
```

---

# 9. Backup Strategy

Recommended backup schedule.

Database

- Daily backup
- Weekly snapshot

Redis

Redis stores temporary data.

Routine backups are optional.

---

# 10. Monitoring

Recommended monitoring tools.

- Prometheus
- Grafana
- Loki

Version 1 only provides application logs.

---

# 11. Logging

Logs should include.

- Startup
- Shutdown
- HTTP requests
- Scheduler
- Worker
- Errors

---

# 12. Security

Production recommendations.

- HTTPS
- Strong JWT secret
- Secure SMTP credentials
- Environment variables
- Firewall
- Database SSL
- Regular backups

---

# 13. Summary

Following this deployment guide allows STAQ to run consistently in both development and production environments.

The deployment process is designed to be reproducible, secure, and scalable.