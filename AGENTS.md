# NeuroSpend - Quick Reference

## Project Overview
Full-stack expense management app for Indian users with bank statement parsing and smart categorization.

**Tech Stack**: Go 1.25 + Gin (backend), Next.js 15 + React 19 (frontend), PostgreSQL, TanStack Query

## Essential Commands (use justfile)

```bash
just install          # Install all dependencies
just dev              # Start both servers (backend:8080, frontend:3000)
just server           # Backend with hot-reload
just frontend         # Next.js with Turbopack

just db-create <name> # Create migration
just db-upgrade       # Apply migrations
just db-seed          # Seed database

just test             # Run all tests
just test "TestName"  # Run single test (Ginkgo focus string)

just format           # Format all code (Go + TS)
```

## Project Structure

```
expenses/
├── server/     # Go backend → see server/AGENTS.md for details
├── frontend/   # Next.js frontend → see frontend/AGENTS.md for details
├── justfile    # Command automation (primary source of truth)
└── README.md   # Full project documentation
```

## Backend (Go)
- **Architecture**: Controller → Service → Repository → DB
- **Testing**: Ginkgo with real database (no external service mocks)
- **DI**: Google Wire (`wire gen ./internal/wire`)
- **DB**: PostgreSQL with Goose migrations

**See `server/AGENTS.md` for:** Build commands, code style, testing patterns, naming conventions

## Frontend (TypeScript/React)
- **Framework**: Next.js 15 App Router (Server Components default)
- **State**: TanStack Query (hooks in `components/hooks/`)
- **UI**: Shadcn UI components only
- **Styling**: Tailwind CSS 4

**See `frontend/AGENTS.md` for:** Component architecture, API patterns, TypeScript conventions

## Development Workflow

**Adding a Feature:**
1. Backend: Model → Repository → Service → Controller → Routes → Tests
2. Frontend: Types → API functions → Hooks → Components → Pages
3. Run `just format` and `just test` before committing

**For detailed guidelines, see:**
- `server/AGENTS.md` - Backend Go conventions, testing, architecture
- `frontend/AGENTS.md` - Frontend React/Next.js patterns, component structure
- `README.md` - Complete project documentation
