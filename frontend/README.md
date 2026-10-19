# STAQ — Frontend

Next.js + React + TypeScript + Tailwind client for the STAQ backend.

## Status

This is an in-progress client. Authentication, task and trigger
management, the executions dashboard, and settings are implemented.
Some screens are still being wired to the API — see
[../docs/development-roadmap.md](../docs/development-roadmap.md) for
the current state.

## Run locally

1. Copy `.env.local.example` to `.env.local` and set `NEXT_PUBLIC_API_URL`
   to the backend (default: `http://localhost:8080`).
2. Install and start:

   ```sh
   npm install
   npm run dev