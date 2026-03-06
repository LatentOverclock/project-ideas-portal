# project-ideas-portal

Requirement-driven implementation using `agents-md`.

## Stack
- Frontend: React + TypeScript (Vite)
- Backend: Go + GraphQL
- Database: PostgreSQL
- Orchestration: Docker Compose

## Run (local)
```bash
docker compose -f docker-compose.yml -f docker-compose.local.yml up -d --build
```

Frontend: http://localhost:5173  
Backend GraphQL: http://localhost:8080/graphql

## Run (deployment)
```bash
docker compose -f docker-compose.yml -f docker-compose.deploy.yml up -d --build
```

## Deployment note
Deployment host/domain is configured externally via environment variables (e.g. `APP_HOST` in deployment compose), not in application code.
