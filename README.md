# NeuroSpend - Intelligent Expense Management

## 🎯 Project Overview
**NeuroSpend** is a modern, full-stack expense management application designed specifically for Indian users. It features automated bank statement parsing for major Indian banks (Axis, SBI, HDFC), comprehensive expense tracking, intelligent categorization, and advanced reporting capabilities with a focus on type safety and robust error handling.

### ✨ Key Features
- 🏦 **Bank Statement Parsing** - Automated import from major Indian banks
- 📊 **Smart Categorization** - AI-powered expense categorization with custom rules
- 🔐 **Secure Authentication** - JWT-based auth with automatic token refresh
- 📱 **Responsive Design** - Modern UI with dark/light theme support
- 🚀 **Real-time Updates** - Optimistic updates with React Query
- 🛡️ **Type Safety** - Full TypeScript coverage with comprehensive error handling
- 🔄 **Offline Support** - Graceful handling of network issues
- 📈 **Analytics** - Detailed spending insights and trends

### 🛠️ Tech Stack
- **Backend**: Go 1.24+ with Gin framework
- **Frontend**: Next.js 15+ with React 19, TypeScript, Tailwind CSS
- **State Management**: React Query (TanStack Query) v5
- **Database**: PostgreSQL with Goose migrations
- **Authentication**: JWT-based auth with refresh tokens
- **UI Components**: shadcn/ui with Radix UI primitives
- **Development**: Just command runner, Air for hot-reload, Ginkgo for testing
- **Error Handling**: Comprehensive error boundaries and type-safe error handling

## 🏗️ Architecture Overview

### Project Structure
```
expenses/
├── server/                 # Go backend application
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
│   └── pkg/               # Public packages
├── frontend/              # Next.js frontend application
│   ├── app/               # App router pages and layouts
│   ├── components/        # React components
│   │   ├── ui/           # Reusable UI components (shadcn/ui)
│   │   ├── custom/       # Application-specific components
│   │   └── providers/    # Context providers
│   ├── hooks/            # Custom React hooks (React Query)
│   ├── lib/              # Utilities and configurations
│   │   ├── api/          # API client functions
│   │   ├── types/        # TypeScript type definitions
│   │   └── utils/        # Utility functions
│   └── docs/             # Frontend documentation
└── justfile              # Command automation
```

## 🔧 Development Setup

### Prerequisites
- **Go 1.24+** - Backend development
- **Node.js 18+** - Frontend development
- **PostgreSQL 14+** - Database
- **Just** - Command runner (`cargo install just` or `brew install just`)
- **Mise** (optional) - Tool version management

### Quick Start
```bash
# Clone the repository
git clone <repository-url>
cd expenses

# Install dependencies
just install

# Set up environment
cp server/.env.example server/.env
# Edit server/.env with your database credentials

# Set up database
just db-upgrade
just db-seed

# Start development servers
just dev  # Starts both backend and frontend
```

### Environment Configuration
Create `server/.env` with the following variables:
```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_username
DB_PASSWORD=your_password
DB_NAME=neurospend_dev

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key
JWT_REFRESH_SECRET=your-super-secret-refresh-key
JWT_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d

# Server Configuration
PORT=8080
GIN_MODE=debug
LOG_LEVEL=debug

# CORS Configuration
CORS_ORIGINS=http://localhost:3000
```

### Key Commands
```bash
# Development
just dev              # Start both backend and frontend
just server           # Start backend only (with hot-reload)
just frontend         # Start frontend only
just install          # Install all dependencies

# Database Operations
just db-upgrade       # Apply migrations
just db-downgrade     # Rollback last migration
just db-seed          # Apply seed data
just db-status        # Check migration status
just db-create <name> # Create new migration

# Testing
just test             # Run all tests
just test-backend     # Run backend tests only
just test-frontend    # Run frontend tests only
just test-e2e         # Run end-to-end tests

# Code Quality
just lint             # Run linters
just format           # Format code
just type-check       # TypeScript type checking

# Build & Deploy
just build            # Build for production
just docker-build     # Build Docker images
```

## 🏛️ Backend Architecture

### Core Components

#### 1. **Clean Architecture Layers**
```
┌─────────────────┐
│   API Layer     │ ← HTTP handlers, middleware, routing
├─────────────────┤
│ Service Layer   │ ← Business logic, validation
├─────────────────┤
│Repository Layer │ ← Data access, database queries
├─────────────────┤
│  Models Layer   │ ← Data structures, DTOs
└─────────────────┘
```

#### 2. **Dependency Injection** (`internal/wire/`)
- Google Wire for compile-time DI
- Provider pattern for resource management
- Clean separation of concerns
- Testable architecture

#### 3. **API Layer** (`internal/api/`)
- RESTful API endpoints with proper HTTP methods
- CORS configuration for frontend integration
- JWT middleware for authentication
- Request validation and error handling
- Rate limiting and security headers

#### 4. **Service Layer** (`internal/service/`)
- Business logic implementation
- Transaction management
- Data validation and transformation
- Integration with external services (bank APIs)

#### 5. **Repository Layer** (`internal/repository/`)
- Database abstraction layer
- PostgreSQL integration with pgx/v5
- Query optimization and connection pooling
- Migration management with Goose

### Key Features

#### Authentication & Authorization
- JWT-based stateless authentication
- Automatic token refresh mechanism
- Role-based access control (future)
- Secure password hashing with bcrypt
- Session management and logout

#### Bank Statement Processing
- Support for major Indian banks (Axis, SBI, HDFC)
- CSV/Excel file parsing
- Automated transaction categorization
- Duplicate detection and handling
- Error reporting and validation

#### Core Entities
1. **Users** - Authentication and profile management
2. **Accounts** - Bank accounts and financial accounts
3. **Categories** - Hierarchical expense categorization
4. **Transactions** - Core expense/income tracking with metadata
5. **Rules** - Automated categorization rules with conditions
6. **Budgets** - Spending limits and tracking (future)

## 🎨 Frontend Architecture

### Modern React Architecture

#### 1. **Next.js 15 App Router**
- Server-side rendering (SSR)
- Static site generation (SSG) where appropriate
- Nested layouts and route groups
- Middleware for authentication
- API routes for server-side logic

#### 2. **State Management with React Query**
```typescript
// Modern data fetching with React Query
const { data: transactions, isLoading, error } = useTransactions({
  page: 1,
  limit: 10,
  category: selectedCategory
});

// Optimistic updates for better UX
const createTransactionMutation = useCreateTransaction({
  onSuccess: () => {
    queryClient.invalidateQueries(['transactions']);
    toast.success('Transaction created!');
  }
});
```

#### 3. **Component Architecture**
```
components/
├── ui/              # Reusable UI primitives (shadcn/ui)
│   ├── button.tsx
│   ├── input.tsx
│   ├── card.tsx
│   └── ...
├── custom/          # Application-specific components
│   ├── Dashboard/
│   ├── Transaction/
│   ├── Modal/
│   └── ...
└── providers/       # Context providers
    ├── QueryProvider.tsx
    ├── ThemeProvider.tsx
    └── ...
```

#### 4. **Custom Hooks** (`hooks/`)
- **useUser** - User authentication and profile management
- **useTransactions** - Transaction CRUD operations with pagination
- **useAccounts** - Account management with optimistic updates
- **useCategories** - Category management and hierarchical data
- **useSession** - Authentication state and token management

#### 5. **Type-Safe API Integration**
```typescript
// Comprehensive error types
interface ApiError extends Error {
  status?: number;
  data?: {
    message?: string;
    errors?: Record<string, string[]>;
  };
}

// Type-safe API calls
export async function getTransactions(params: TransactionParams): Promise<PaginatedResponse<Transaction>> {
  const response = await apiClient.get('/transactions', { params });
  return response.data;
}
```

### Advanced Features

#### 1. **Error Boundaries & Error Handling**
- Global error boundary for unhandled errors
- Query-specific error boundaries for data fetching
- Type-safe error handling throughout the application
- User-friendly error messages with recovery options
- Automatic error reporting (configurable)

#### 2. **Authentication System**
```typescript
// Automatic token refresh
const { isAuthenticated, isLoading } = useSession();

// Protected routes with AuthGuard
<AuthGuard>
  <Dashboard />
</AuthGuard>
```

#### 3. **Theme System**
- Dark/light mode with system preference detection
- Consistent design tokens with CSS variables
- Smooth theme transitions
- Persistent theme selection

#### 4. **Performance Optimizations**
- React Query caching and background updates
- Optimistic updates for better perceived performance
- Code splitting with Next.js dynamic imports
- Image optimization with Next.js Image component
- Bundle analysis and optimization

## 🛡️ Error Handling & Type Safety

### Comprehensive Error System
```typescript
// Type-safe error handling
export type ApiErrorType = 
  | AuthError          // 401 - Authentication errors
  | ValidationError    // 400 - Validation errors
  | ForbiddenError     // 403 - Authorization errors
  | NotFoundError      // 404 - Resource not found
  | ServerError        // 5xx - Server errors
  | NetworkError;      // Network/connection errors

// Error boundaries for graceful error handling
<ErrorBoundary onError={reportError}>
  <QueryErrorBoundary>
    <YourComponent />
  </QueryErrorBoundary>
</ErrorBoundary>
```

### Type Safety Features
- **Zero `any` types** - Complete TypeScript coverage
- **API response typing** - All API responses are typed
- **Form validation** - Type-safe form handling with validation
- **Error type guards** - Safe error property access
- **Compile-time checks** - Catch errors before runtime

## 🧪 Testing Strategy

### Backend Testing
- **Unit Tests** - Service and repository layer testing with Ginkgo/Gomega
- **Integration Tests** - API endpoint testing with test database
- **E2E Tests** - Full workflow testing
- **Test Coverage** - Comprehensive coverage reporting
- **Mock Testing** - External service mocking

### Frontend Testing
- **Component Tests** - React component testing with React Testing Library
- **Hook Tests** - Custom hook testing with React Query testing utilities
- **Integration Tests** - User workflow testing
- **Visual Regression** - UI consistency testing
- **Accessibility Tests** - WCAG compliance testing

### Test Organization
```
server/
├── internal/
│   ├── service/
│   │   ├── user_service.go
│   │   └── user_service_test.go
│   └── repository/
│       ├── user_repository.go
│       └── user_repository_test.go

frontend/
├── __tests__/
│   ├── components/
│   ├── hooks/
│   └── utils/
└── e2e/
    └── specs/
```

## 🚀 Deployment & Operations

### Development Workflow
1. **Backend Development**: `just server` for hot-reload with Air
2. **Frontend Development**: `just frontend` for Next.js dev server
3. **Full Stack**: `just dev` to run both simultaneously
4. **Database**: Automatic migrations on server start
5. **Testing**: `just test` for comprehensive test suite

### Production Deployment
```bash
# Build production assets
just build

# Docker deployment
just docker-build
docker-compose up -d

# Environment-specific configuration
cp .env.production .env
```

### Production Considerations
- **Environment Configuration** - Separate configs for dev/staging/prod
- **Database Connection Pooling** - Optimized connection management
- **Graceful Shutdown** - Proper cleanup on server termination
- **CORS Configuration** - Production domain whitelisting
- **Security Headers** - Comprehensive security middleware
- **Monitoring & Logging** - Structured logging with error tracking
- **Performance Monitoring** - API response time tracking
- **Health Checks** - Kubernetes/Docker health endpoints

## 🔍 Key Architectural Decisions

### Backend Decisions
1. **Clean Architecture** - Clear separation of concerns with dependency inversion
2. **Dependency Injection** - Wire for compile-time DI, avoiding runtime reflection
3. **Repository Pattern** - Database abstraction for testability
4. **Service Layer** - Business logic isolation from HTTP concerns
5. **Database First** - Schema-driven development with migrations

### Frontend Decisions
1. **React Query** - Server state management with caching and synchronization
2. **Component Composition** - Reusable UI components with shadcn/ui
3. **Type-First Development** - TypeScript-first approach with strict typing
4. **Error Boundaries** - Graceful error handling and recovery
5. **Modern React Patterns** - Hooks, context, and functional components

### Technology Choices
1. **Go + Gin** - High performance, simple HTTP framework with excellent concurrency
2. **Next.js 15** - Full-stack React framework with SSR and modern features
3. **PostgreSQL** - Robust relational database with JSON support and ACID compliance
4. **React Query** - Powerful data synchronization and caching library
5. **TypeScript** - Type safety and better developer experience

## 🛠️ Development Guide

### Adding New Features

#### Backend Feature Development
1. **Define Models** - Add data structures in `internal/models/`
2. **Create Repository** - Implement data access in `internal/repository/`
3. **Implement Service** - Add business logic in `internal/service/`
4. **Add API Handlers** - Create HTTP handlers in `internal/api/`
5. **Update Routes** - Register routes in `internal/api/routes.go`
6. **Write Tests** - Add comprehensive tests for all layers
7. **Update Documentation** - Document API endpoints and usage

#### Frontend Feature Development
1. **Define Types** - Add TypeScript interfaces in `lib/types/`
2. **Create API Client** - Add API functions in `lib/api/`
3. **Build Components** - Create UI components in `components/custom/`
4. **Add Hooks** - Implement React Query hooks in `hooks/`
5. **Create Pages** - Add routes in `app/` directory
6. **Handle Errors** - Add error boundaries and error handling
7. **Write Tests** - Add component and integration tests

### Database Management
```bash
# Create new migration
just db-create add_user_preferences

# Apply migrations
just db-upgrade

# Rollback if needed
just db-downgrade

# Check migration status
just db-status
```

### Code Quality Standards
- **Linting** - ESLint for JavaScript/TypeScript, golangci-lint for Go
- **Formatting** - Prettier for frontend, gofmt for backend
- **Type Checking** - Strict TypeScript configuration
- **Testing** - Minimum 80% code coverage
- **Documentation** - Comprehensive README and inline documentation

## 📚 Important Files & Directories

### Configuration Files
- `justfile` - Command automation and development workflows
- `server/.env.example` - Environment variables template
- `mise.toml` - Tool version management
- `frontend/package.json` - Node.js dependencies and scripts
- `server/go.mod` - Go dependencies and module definition

### Entry Points
- `server/cmd/neurospend/main.go` - Backend application entry point
- `frontend/app/layout.tsx` - Frontend root layout with providers
- `frontend/app/page.tsx` - Homepage/dashboard
- `server/internal/server/server.go` - HTTP server setup

### Core Logic
- `server/internal/api/routes.go` - API route definitions
- `server/internal/wire/wire.go` - Dependency injection configuration
- `frontend/lib/query-client.ts` - React Query configuration
- `frontend/hooks/` - Custom React hooks for data management

### Documentation
- `frontend/AUTHENTICATION-FIX.md` - Authentication flow documentation
- `frontend/ERROR-TYPES-FIX.md` - Error handling improvements
- `frontend/HOOKS-MIGRATION-COMPLETE.md` - React Query migration guide

## 🎯 Getting Started for New Developers

### 1. **Environment Setup**
```bash
# Install prerequisites
brew install go node postgresql just

# Clone and setup
git clone <repo-url>
cd expenses
just install
```

### 2. **Understanding the Codebase**
1. **Start with the README** - This document provides the overview
2. **Explore the Backend** - Begin with `cmd/neurospend/main.go`
3. **Understand the Frontend** - Start with `app/layout.tsx` and `app/page.tsx`
4. **Review the API** - Check `internal/api/routes.go` for available endpoints
5. **Study the Hooks** - Examine `hooks/` directory for data management patterns

### 3. **Making Your First Change**
1. **Pick a small feature** - Start with a simple UI improvement
2. **Follow the patterns** - Use existing code as a template
3. **Write tests** - Add tests for your changes
4. **Test thoroughly** - Use `just test` to run the full test suite
5. **Submit for review** - Create a pull request with clear description

### 4. **Development Best Practices**
- **Type Safety First** - Always use TypeScript types, avoid `any`
- **Error Handling** - Implement proper error boundaries and handling
- **Testing** - Write tests for new features and bug fixes
- **Documentation** - Update documentation for significant changes
- **Performance** - Consider performance implications of changes

## 🌟 Unique Features

### Indian Banking Integration
- **Multi-bank Support** - Axis Bank, SBI, HDFC Bank statement parsing
- **Smart Categorization** - AI-powered expense categorization for Indian spending patterns
- **Currency Handling** - INR-first design with proper formatting
- **Regional Customization** - Indian financial year, tax categories, etc.

### Advanced Expense Management
- **Rule-based Categorization** - Custom rules for automatic transaction categorization
- **Recurring Transaction Detection** - Identify and manage subscription payments
- **Budget Tracking** - Set and monitor spending limits by category
- **Expense Analytics** - Detailed insights and spending trends
- **Export Capabilities** - CSV/PDF export for tax filing and record keeping

### Developer Experience
- **Type-Safe Development** - Complete TypeScript coverage with strict typing
- **Modern Tooling** - Latest versions of React, Next.js, and Go
- **Comprehensive Testing** - Unit, integration, and E2E testing
- **Error Boundaries** - Graceful error handling and recovery
- **Performance Optimized** - React Query caching and optimistic updates

---

## 🤝 Contributing

We welcome contributions! Please read our contributing guidelines and code of conduct before submitting pull requests.

### Development Process
1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Ensure all tests pass
5. Submit a pull request

### Code Standards
- Follow existing code patterns
- Write comprehensive tests
- Update documentation
- Use TypeScript strictly
- Handle errors gracefully

---

**NeuroSpend** represents a modern approach to expense management with a focus on Indian users, type safety, and exceptional developer experience. The architecture is designed to be scalable, maintainable, and performant while providing a delightful user experience.
