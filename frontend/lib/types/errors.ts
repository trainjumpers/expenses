// Base API Error interface
export interface ApiError extends Error {
  status?: number;
  statusText?: string;
  data?: unknown;
}

// HTTP Error with status code
export interface HttpError extends Error {
  status: number;
  statusText: string;
  data?: {
    message?: string;
    error?: string;
    details?: unknown;
  };
}

// Network Error (no response received)
export interface NetworkError extends Error {
  code?: string;
  errno?: number;
  syscall?: string;
}

// Validation Error (400 responses)
export interface ValidationError extends HttpError {
  status: 400;
  data: {
    message: string;
    errors?: Record<string, string[]>;
    field_errors?: Record<string, string>;
  };
}

// Authentication Error (401 responses)
export interface AuthError extends HttpError {
  status: 401;
  data?: {
    message?: string;
    error?: "invalid_token" | "token_expired" | "unauthorized";
  };
}

// Authorization Error (403 responses)
export interface ForbiddenError extends HttpError {
  status: 403;
  data?: {
    message?: string;
    error?: string;
  };
}

// Not Found Error (404 responses)
export interface NotFoundError extends HttpError {
  status: 404;
  data?: {
    message?: string;
    error?: string;
  };
}

// Conflict Error (409 responses)
export interface ConflictError extends HttpError {
  status: 409;
  data?: {
    message?: string;
    error?: string;
  };
}

// Server Error (5xx responses)
export interface ServerError extends HttpError {
  status: number; // 500-599
  data?: {
    message?: string;
    error?: string;
    trace?: string;
  };
}

// Union type for all possible API errors
export type ApiErrorType =
  | ApiError
  | HttpError
  | NetworkError
  | ValidationError
  | AuthError
  | ForbiddenError
  | NotFoundError
  | ConflictError
  | ServerError;

// Type guard functions
export function isHttpError(error: unknown): error is HttpError {
  return (
    error instanceof Error &&
    "status" in error &&
    typeof (error as HttpError).status === "number"
  );
}

export function isAuthError(error: unknown): error is AuthError {
  return isHttpError(error) && error.status === 401;
}

export function isValidationError(error: unknown): error is ValidationError {
  return isHttpError(error) && error.status === 400;
}

export function isForbiddenError(error: unknown): error is ForbiddenError {
  return isHttpError(error) && error.status === 403;
}

export function isNotFoundError(error: unknown): error is NotFoundError {
  return isHttpError(error) && error.status === 404;
}

export function isConflictError(error: unknown): error is ConflictError {
  return isHttpError(error) && error.status === 409;
}

export function isServerError(error: unknown): error is ServerError {
  return isHttpError(error) && error.status >= 500 && error.status < 600;
}

export function isNetworkError(error: unknown): error is NetworkError {
  return error instanceof Error && "code" in error && !("status" in error);
}

// Helper function to extract error message
export function getErrorMessage(error: unknown): string {
  if (isHttpError(error)) {
    return (
      error.data?.message ||
      error.message ||
      `HTTP ${error.status}: ${error.statusText}`
    );
  }

  if (error instanceof Error) {
    return error.message;
  }

  if (typeof error === "string") {
    return error;
  }

  return "An unknown error occurred";
}

// Helper function to get error status
export function getErrorStatus(error: unknown): number | undefined {
  if (isHttpError(error)) {
    return error.status;
  }
  return undefined;
}

// Helper function to check if error is statement password required
export function isStatementPasswordRequiredError(error: unknown): boolean {
  if (typeof error === "object" && error !== null) {
    const err = error as { message?: string; isPasswordRequired?: boolean };
    return (
      err.isPasswordRequired === true ||
      err.message === "statement password required"
    );
  }
  return false;
}
