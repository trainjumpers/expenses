package errors

import "net/http"

func NewStatementNotFoundError(err error) *AuthError {
	return formatError(http.StatusNotFound, "statement not found", err, "StatementNotFound")
}

func NewStatementCreateError(err error) *AuthError {
	return formatError(http.StatusInternalServerError, "failed to create statement", err, "StatementCreateError")
}

func NewStatementGetError(err error) *AuthError {
	return formatError(http.StatusInternalServerError, "failed to get statement", err, "StatementGetError")
}

func NewStatementUpdateError(err error) *AuthError {
	return formatError(http.StatusInternalServerError, "failed to update statement", err, "StatementUpdateError")
}

func NewStatementBadRequestError(err error) *AuthError {
	return formatError(http.StatusBadRequest, "invalid request", err, "StatementBadRequestError")
}
