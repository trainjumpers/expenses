# NeuroSpend

Expense management application for Indian users with automated bank statement parsing.

## Features

- **Bank Statement Import**: Parse statements from HDFC, ICICI, Axis, SBI banks
- **Smart Categorization**: Rule-based automatic expense categorization
- **Analytics**: Spending insights and trends
- **JWT Authentication**: Secure auth with token refresh
- **Dark/Light Theme**: Responsive UI with theme switching

## Tech Stack

- **Backend**: Go 1.25, Gin, PostgreSQL, Ginkgo testing, Goose migrations
- **Frontend**: Next.js 15, React 19, TypeScript, Tailwind CSS 4, Shadcn UI, TanStack Query, Bun

## Quick Start

```bash
# Install dependencies
just install

# Set up database
cp server/.env.example server/.env
# Edit server/.env with your DB credentials
just db-upgrade
just db-seed

# Start development servers
just dev
```

The backend will run on `http://localhost:8080` and frontend on `http://localhost:3000`.

**Note**: Ensure Bun is installed and activated in your shell. If using mise, run: `eval "$(mise activate zsh)"`

## Environment Setup

Create `server/.env` with:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_username
DB_PASSWORD=your_password
DB_NAME=neurospend_dev
DB_SCHEMA=public

JWT_SECRET=your-jwt-secret
JWT_REFRESH_SECRET=your-refresh-secret
JWT_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

PORT=8080
GIN_MODE=debug
CORS_ORIGINS=http://localhost:3000
```

## Commands

```bash
# Development
just install          # Install Go tools and npm packages
just dev              # Start both servers
just server           # Backend with hot-reload
just frontend         # Frontend dev server

# Database
just db-create <name> # Create migration
just db-upgrade       # Apply migrations
just db-downgrade     # Rollback migration
just db-seed          # Seed database

# Testing
just test             # Run all tests
just test "Pattern"   # Run specific test

# Code quality
just format           # Format all code
```

## Project Structure

```
expenses/
├── server/                 # Go backend
│   ├── cmd/neurospend/     # Entry point
│   ├── internal/
│   │   ├── api/           # Controllers, routes
│   │   ├── service/       # Business logic
│   │   ├── repository/    # Database layer
│   │   ├── models/        # DTOs
│   │   ├── parser/        # Bank statement parsers
│   │   ├── errors/        # Custom errors
│   │   └── database/      # Migrations
│   └── e2e_test.sh       # Test runner
├── frontend/              # Next.js
│   ├── app/               # Pages
│   ├── components/        # React components
│   │   ├── ui/           # Shadcn UI
│   │   ├── custom/       # Feature components
│   │   └── hooks/        # TanStack Query hooks
│   └── lib/
│       ├── api/          # API clients
│       └── models/       # TypeScript types
└── justfile              # Command runner
```

## Architecture

**Backend**: Clean architecture with layered design
- Controller: HTTP handlers, error handling
- Service: Business logic with debug logging
- Repository: Database operations (pgx/v5)
- Models: Request/response DTOs
- Wire: Dependency injection (Google Wire)

**Frontend**: Next.js 15 App Router
- Server Components by default
- TanStack Query for server state
- Shadcn UI for components
- Context for global state

## Development

See detailed guides in:
- `AGENTS.md` - Project quick reference
- `server/AGENTS.md` - Backend conventions, testing, Go patterns
- `frontend/AGENTS.md` - React patterns, component architecture
- `README.md` - This file

### Adding a Feature

**Backend**:
1. Create model in `internal/models/`
2. Add repository in `internal/repository/`
3. Implement service in `internal/service/`
4. Create controller in `internal/api/controller/`
5. Add route in `internal/api/routes.go`
6. Write tests with Ginkgo
7. Run `just test "YourTestName"`

**Frontend**:
1. Add types in `lib/models/`
2. Create API function in `lib/api/`
3. Write hook in `components/hooks/`
4. Build component in `components/custom/`
5. Add page in `app/`
6. Use Shadcn UI from `components/ui/`

### Testing

**Backend** (Ginkgo):
- Real database with migrations (no external mocks)
- Run specific test: `just test "ServiceName.*"`
- Test files end with `_test.go`

**Frontend**:
- No test framework configured yet

## API

The backend REST API runs on port 8080. Key endpoints:
- `POST /api/auth/signup` - Create user
- `POST /api/auth/login` - Login
- `POST /api/auth/refresh` - Refresh token
- `GET /api/accounts` - List accounts
- `GET /api/transactions` - List transactions
- `POST /api/statements` - Import statement

See `server/internal/api/routes.go` for full API.
