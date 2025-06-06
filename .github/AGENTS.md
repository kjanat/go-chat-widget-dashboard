# Go Chat Widget Dashboard Knowledge Base

This file provides coding conventions, project structure details, and pull request guidelines for contributors and AI assistants.

## Project Overview
This application exposes a chat widget and admin dashboard built with Go. It supports OpenAI integration, WebSocket chat, and a 3D avatar system.

## Project Structure
```
/             - main entry point and configuration
/internal/    - application packages (database, services, handlers)
/scripts/     - helper scripts (database initialization)
/web/         - templates and static assets
/db/          - sqlite database files (ignored except for .gitkeep)
```

## Coding Conventions
- Go modules with standard formatting (`gofmt`)
- Keep handlers under `internal/handlers`
- Services reside in `internal/services`
- Templates live in `web/templates`

## Testing
Run `make test` to execute Go tests across packages.

## Pull Requests
Include a concise summary and list of test commands in the PR body.
