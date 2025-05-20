package errors

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

type AuthError struct {
	Message   string // User-friendly message
	Err       error  // Original error for debugging
	Stack     string // Stack trace
	ErrorType string // Type of error (e.g., "InvalidCredentials", "TokenGeneration")
	Status    int    // HTTP status code
}

func (e *AuthError) Error() string {
	return e.Message
}

// Unwrap returns the original error
func (e *AuthError) Unwrap() error {
	return e.Err
}

func GetStackTrace(err error) string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var stack strings.Builder
	for {
		frame, more := frames.Next()
		fmt.Fprintf(&stack, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
		if !more {
			break
		}
	}
	return stack.String()
}

// FormatError creates a new AuthError with stack trace
func FormatError(status int, message string, err error, errorType string) *AuthError {
	return &AuthError{
		Message:   message,
		Err:       err,
		Stack:     GetStackTrace(err),
		ErrorType: errorType,
		Status:    status,
	}
}

func IsErrorType(err error, errorType string) bool {
	if authErr, ok := err.(*AuthError); ok {
		return authErr.ErrorType == errorType
	}
	return false
}

func New(err error) *AuthError {
	return FormatError(http.StatusInternalServerError, "internal server error", err, "InternalServerError")
}

func NoFieldsToUpdateError() *AuthError {
	return FormatError(http.StatusBadRequest, "no fields to update", fmt.Errorf("no fields to update"), "NoFieldsToUpdate")
}

func CheckForeignKey(err error, fkKey string) bool {
	return strings.Contains(err.Error(), fkKey) && strings.Contains(err.Error(), "constraint") && strings.Contains(err.Error(), "violates")
}
