# Expenses Management

## 🎯 Project Overview
**NeuroSpend** is a full-stack expense management application designed specifically for Indian users. It features automated bank statement parsing for major Indian banks (Axis, SBI, HDFC) and provides comprehensive expense tracking, categorization, and reporting capabilities.

### Tech Stack
- **Backend**: Go 1.24+ with Gin framework
- **Frontend**: Next.js 15+ with React 19, TypeScript, Tailwind CSS
- **Database**: PostgreSQL with Goose migrations
- **Authentication**: JWT-based auth with refresh tokens
- **Development**: Just command runner, Air for hot-reload, Ginkgo for testing

## 🏗️ Architecture Overview

### Project Structure
```
expenses/
├── server/           # Go backend application
│   ├── cmd/neurospend/     # Application entry point
│   ├── internal/           # Private application code
│   │   ├── api/           # HTTP handlers and routes
│   │   ├── config/        # Configuration management
│   │   ├── database/      # Database migrations and management
│   │   ├── models/        # Data models and DTOs
│   │   ├── repository/    # Data access layer
│   │   ├── service/       # Business logic layer
│   │   ├── wire/          # Dependency injection
│   │   ├── validator/     # Input validation
│   │   ├── parser/        # Bank statement parsers
│   │   └── errors/        # Error handling
│   └── pkg/              # Public packages
├── frontend/         # Next.js frontend application
│   ├── app/              # App router pages
│   ├── components/       # React components
│   │   ├── ui/           # Reusable UI components (shadcn/ui)
│   │   └── custom/       # Application-specific components
│   └── lib/              # Utilities and API clients
└── justfile          # Command automation
```

## 🔧 Development Setup

### Prerequisites
- Go 1.24+
- Node.js (latest)
- PostgreSQL
- Just command runner
- Mise (optional, for tool management)

### Environment Configuration
1. Copy `server/.env.example` to `server/.env` and configure:
   - Database connection details
   - JWT secrets
   - Server port
   - Logging level

2. Set up PostgreSQL database and schema as described in README.md

### Key Commands
```bash
# Install dependencies
just install

# Run backend server (with hot-reload)
just server

# Database operations
just db-upgrade        # Apply migrations
just db-seed          # Apply seed data
just db-status        # Check migration status
just db-downgrade     # Rollback last migration

# Testing
just test             # Run all tests
just test "focus"     # Run focused tests
```

## 🏛️ Backend Architecture

### Core Components

#### 1. **Entry Point** (`cmd/neurospend/main.go`)
- Simple main function that calls `server.Start()`
- Minimal bootstrap code

#### 2. **Server Setup** (`internal/server/server.go`)
- HTTP server configuration with timeouts
- Graceful shutdown handling
- Uses Wire for dependency injection
- Configurable port (default: 8080)

#### 3. **Dependency Injection** (`internal/wire/`)
- Google Wire for compile-time DI
- Provider pattern for resource management
- Clean separation of concerns

#### 4. **API Layer** (`internal/api/`)
- **Routes** (`routes.go`): RESTful API endpoints with CORS
- **Controllers**: Handle HTTP requests/responses
- **Middleware**: Authentication, logging, CORS

#### 5. **Business Logic** (`internal/service/`)
- Service interfaces and implementations
- Core business rules
- Transaction management
- Comprehensive test coverage

#### 6. **Data Access** (`internal/repository/`)
- Repository pattern for data access
- PostgreSQL integration with pgx/v5
- Query builders and data mapping

#### 7. **Models** (`internal/models/`)
- Request/Response DTOs
- Database models
- Validation tags
- Pagination structures

### Key Features

#### Authentication & Authorization
- JWT-based authentication with refresh tokens
- Protected routes with middleware
- User management (CRUD operations)
- Password hashing with bcrypt

#### Core Entities
1. **Users**: Authentication and profile management
2. **Accounts**: Bank accounts and financial accounts
3. **Categories**: Expense categorization system
4. **Transactions**: Core expense/income tracking
5. **Rules**: Automated transaction categorization

#### Database Design
- PostgreSQL with schema-based organization
- Goose migrations with environment variable substitution
- Soft deletes for users
- Audit trails with created_at/updated_at
- Foreign key relationships with proper indexing

#### Bank Statement Parsing
- Support for major Indian banks (Axis, SBI, HDFC)
- CSV parsing capabilities
- Automated transaction import
- Rule-based categorization

## 🎨 Frontend Architecture

### Core Structure

#### 1. **App Router** (`app/`)
- Next.js 15 App Router
- Server-side rendering
- Nested layouts
- Route groups for organization

#### 2. **Component Architecture**
- **UI Components** (`components/ui/`): shadcn/ui based reusable components
- **Custom Components** (`components/custom/`): Application-specific components
- **Providers**: Context providers for state management

#### 3. **State Management**
- React Context for global state
- Custom providers for:
  - Session management
  - User data
  - Account data
  - Category data
  - Theme management

#### 4. **API Integration** (`lib/api/`)
- Centralized API client
- Type-safe request/response handling
- Error handling with toast notifications
- Automatic cookie-based authentication

### Key Features

#### UI/UX
- Modern, responsive design with Tailwind CSS
- Dark/light theme support
- Toast notifications (Sonner)
- Loading states and skeletons
- Accessible components (Radix UI)

#### Data Management
- Optimistic updates
- Client-side caching
- Real-time data synchronization
- Form validation

## 🧪 Testing Strategy

### Backend Testing
- **Ginkgo/Gomega**: BDD-style testing framework
- **E2E Tests**: Full integration testing with database
- **Unit Tests**: Service and repository layer testing
- **Test Database**: Isolated test schema
- **Coverage**: Comprehensive test coverage tracking

### Test Organization
- Service layer tests for business logic
- Repository tests for data access
- Integration tests for API endpoints
- Mock implementations for external dependencies

## 🚀 Deployment & Operations

### Development Workflow
1. Use `just server` for backend development with hot-reload
2. Use `npm run dev` for frontend development
3. Database migrations are handled automatically
4. Tests run in isolated environment

### Production Considerations
- Environment-based configuration
- Database connection pooling
- Graceful shutdown handling
- CORS configuration for production domains
- JWT token security

## 🔍 Key Decision Points

### Architecture Decisions
1. **Monorepo Structure**: Frontend and backend in same repository for easier development
2. **Clean Architecture**: Clear separation between layers (API → Service → Repository)
3. **Dependency Injection**: Wire for compile-time DI, avoiding runtime reflection
4. **Database First**: Schema-driven development with migrations

### Technology Choices
1. **Go + Gin**: High performance, simple HTTP framework
2. **Next.js**: Full-stack React framework with SSR capabilities
3. **PostgreSQL**: Robust relational database with JSON support
4. **JWT**: Stateless authentication suitable for distributed systems

## 🛠️ Common Development Tasks

### Adding New Features
1. **Backend**:
   - Add model in `internal/models/`
   - Create repository interface and implementation
   - Implement service layer with business logic
   - Add controller for HTTP handling
   - Update routes in `internal/api/routes.go`
   - Add tests for all layers

2. **Frontend**:
   - Create API client in `lib/api/`
   - Add TypeScript types in `lib/models/`
   - Create components in `components/custom/`
   - Add pages in `app/`
   - Update providers if needed

### Database Changes
1. Create migration: `just db-create migration_name`
2. Edit migration file in `internal/database/migrations/`
3. Apply migration: `just db-upgrade`
4. Update models and repositories accordingly

### Debugging Tips
1. **Backend**: Check logs, use debugger, run specific tests
2. **Frontend**: Browser dev tools, React dev tools, network tab
3. **Database**: Use `just db-status` to check migrations
4. **API**: Test endpoints with curl or Postman

## 📚 Important Files to Know

### Configuration
- `justfile`: Command automation and development workflows
- `server/.env.example`: Environment variables template
- `mise.toml`: Tool version management
- `server/go.mod`: Go dependencies
- `frontend/package.json`: Node.js dependencies

### Entry Points
- `server/cmd/neurospend/main.go`: Backend entry point
- `frontend/app/layout.tsx`: Frontend root layout
- `frontend/app/page.tsx`: Homepage

### Core Logic
- `server/internal/api/routes.go`: API route definitions
- `server/internal/wire/wire.go`: Dependency injection setup
- `frontend/lib/api/request.ts`: API client base

## 🎯 Next Steps for New Developers

1. **Setup**: Follow README.md setup instructions
2. **Explore**: Run the application and explore the UI
3. **Code Reading**: Start with `main.go` and follow the flow
4. **Testing**: Run tests to understand expected behavior
5. **Small Changes**: Make a small feature addition to understand the workflow

This codebase follows clean architecture principles with clear separation of concerns, comprehensive testing, and modern development practices. The Indian bank statement parsing feature is a unique selling point that sets it apart from generic expense trackers.
