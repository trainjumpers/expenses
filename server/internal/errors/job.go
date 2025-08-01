package errors

import "net/http"

func NewJobNotFoundError(err error) *AuthError {
	return formatError(http.StatusNotFound, "job not found", err, "JobNotFound")
}

func NewJobRepositoryError(message string, err error) *AuthError {
	return formatError(http.StatusInternalServerError, message, err, "JobRepositoryError")
}
