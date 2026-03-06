# project.v1 - Initial requirements (delta)

## Goal
Build a web application where users can register/login and submit project ideas that can later be implemented by an AI agent.

## Functional requirements
1. Users can register with email + password.
2. Users can log in with existing credentials.
3. Authenticated users can create project ideas (title + description).
4. Users can view a list of submitted ideas (newest first).
5. Each idea stores creator and creation timestamp.

## Non-functional requirements
1. Frontend and backend run in separate containers via docker-compose.
2. Backend uses Go + PostgreSQL + GraphQL.
3. Frontend uses TypeScript + React (+ Tailwind scaffold acceptable).
4. Domain/host values must not be hardcoded in application runtime code.
