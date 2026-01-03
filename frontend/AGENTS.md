# Frontend Development Guide

## Build, Lint, and Test Commands

### Development
- `npm run dev` - Start Next.js dev server with Turbopack (port 3000)
- `just frontend` - Start frontend dev server (from root)
- `npm run build` - Production build
- `npm start` - Start production server

### Formatting & Linting
- `npm run format` - Format all files with Prettier
- `npm run lint` - Run ESLint
- `npm run lint:fix` - Auto-fix ESLint issues

## Tech Stack & Frameworks

- **React 19** - Latest with Server Components
- **Next.js 15** - App Router, Turbopack
- **TypeScript 5** - Strict mode enabled
- **Tailwind CSS 4** - Utility-first styling
- **Shadcn UI** - Pre-built Radix UI components
- **TanStack Query** - Server state management
- **Zod** - Schema validation (via go-playground/validator)
- **Sonner** - Toast notifications
- **Date-fns** - Date utilities
- **Lucide React** - Icons

## Code Style Guidelines

### Component Structure
```tsx
"use client"; // Only when needed

import React from "react";
import { Button } from "@/components/ui/button";

interface Props {
  // prop types
}

export default function ComponentName({ ...props }: Props) {
  // State
  // Effects
  // Handlers
  // Render
}
```

### Naming Conventions
- **Files**: PascalCase (e.g., `UserProfile.tsx`, `useTransactions.ts`)
- **Components**: PascalCase (e.g., `UserProfile`, `TransactionTable`)
- **Hooks**: camelCase with `use` prefix (e.g., `useTransactions`, `useSession`)
- **Directories**: lowercase with dashes (e.g., `transaction-filters`, `user-auth`)
- **Constants**: UPPER_SNAKE_CASE (e.g., `PUBLIC_ROUTES`)
- **Interfaces**: PascalCase (e.g., `Transaction`, `ApiResponse`)
- **Event Handlers**: `handle` prefix (e.g., `handleClick`, `handleSubmit`)

### TypeScript Best Practices
- Use interfaces for object shapes, types for unions/primitives
- Avoid enums - use const objects instead
- Use `satisfies` operator for type validation
- Export types from `lib/models/` for shared types
- Prefer `readonly` for immutable arrays
- Use `unknown` over `any` when type is uncertain

### React 19 Patterns
- Use Server Components by default (no `"use client"`)
- Minimize client directives - only add when needed
- Use `useActionState` instead of deprecated `useFormState`
- Use enhanced `useFormStatus` with new properties (data, method, action)
- Use Suspense boundaries for async components

### State Management
- **Server State**: TanStack Query with hooks in `components/hooks/`
- **URL State**: Manage via Next.js `useSearchParams` and `router.push()`
- **Local State**: React `useState`, `useReducer` for small UI state
- **Global State**: React Context for session, theme (suspense-aware)

### Component Architecture
```
components/
├── ui/              # Shadcn UI components (no business logic)
├── custom/          # Business components (feature-specific)
│   ├── Dashboard/
│   ├── Transaction/
│   └── Modal/
├── hooks/           # Custom React hooks for data fetching
├── providers/       # Context providers
└── AuthGuard.tsx    # Authentication wrapper

lib/
├── api/             # API functions (one file per resource)
├── models/          # TypeScript interfaces/types
├── utils/           # Pure utility functions
├── constants/       # App constants
└── query-client.ts  # TanStack Query configuration
```

### Imports
- Group and sort imports alphabetically
- Use `@/` alias for absolute imports
- Order: external, internal components, internal hooks, internal lib
```tsx
import { Button } from "@/components/ui/button";
import { useTransactions } from "@/components/hooks/useTransactions";
import { formatCurrency } from "@/lib/utils";
```

### API Layer
- Use `lib/api/request.ts` `apiRequest()` for all fetch calls
- API functions in `lib/api/` correspond to backend endpoints
- Return `{ data: T }` pattern from API functions
- Use custom error handlers for specific status codes
- Always include credentials for cookies

### Data Fetching (TanStack Query)
- Create hooks in `components/hooks/` (e.g., `useTransactions.ts`)
- Use `queryKeys` object from `lib/query-client.ts` for cache keys
- Implement optimistic updates for mutations
- Use `enabled` to prevent unnecessary requests
- Set appropriate `staleTime` for caching

### Styling
- Use Shadcn UI components - do not create custom UI components
- Tailwind classes via `cn()` utility for conditional classes
- Follow design tokens (primary, secondary, accent, destructive)
- Use responsive prefixes (md:, lg:) for breakpoints
- Dark mode via `next-themes` - use `dark:` prefix

### Error Handling
- Use `ApiErrorType` from `lib/types/errors.ts`
- `getErrorMessage()` for user-friendly error messages
- Toast notifications via `sonner` for user feedback
- Console.error for debugging only
- Handle API errors in request layer, show toasts at component level

### Form Handling
- Use native HTML5 validation
- React Hook Form for complex forms
- Zod schemas for validation (via validator)
- `useFormStatus` for form submission state
- Show loading states during mutations

### File Organization
- Page routes in `app/` directory
- Reusable components in `components/`
- Shared utilities in `lib/`
- Models/interfaces in `lib/models/`
- Constants in `lib/constants/`

### Performance
- Use React.lazy for code splitting
- Implement proper loading states
- Use virtualization for long lists (`@tanstack/react-virtual`)
- Optimize images with Next.js Image component
- Debounce search inputs

### Accessibility
- Use semantic HTML
- ARIA labels on interactive elements
- Keyboard navigation support
- Focus management in modals
- Alt text for images

## Testing

No test framework is currently configured. Consider adding:
- Vitest for unit tests
- Testing Library for component tests
- Playwright for E2E tests

## Development Workflow

1. Create models in `lib/models/` if needed
2. Add API functions in `lib/api/<resource>.ts`
3. Create hooks in `components/hooks/use<Resource>.ts`
4. Build UI components in `components/custom/`
5. Use Shadcn UI components from `components/ui/`
6. Add routes in `app/` directory
7. Run `npm run lint:fix` and `npm run format` before commit

## Notes

- Always use TypeScript - no JavaScript files
- Server Components by default, add `"use client"` only when needed
- Use Shadcn UI for all UI components - don't reinvent the wheel
- Keep components focused and single-purpose
- Use absolute imports via `@/` alias
- Session management via `useSession()` hook
- Toast notifications via `sonner`
- Date formatting via `date-fns`
- Icons from `lucide-react`
- Currency formatting via `formatCurrency()` utility
