package errors

import "net/http"

func NewInvalidCredentialsError(err error) *AuthError {
	return formatError(http.StatusUnauthorized, "invalid credentials", err, "InvalidCredentials")
}

func NewInvalidTokenError(err error) *AuthError {
	return formatError(http.StatusUnauthorized, "invalid token", err, "InvalidToken")
}

func NewTokenGenerationError(err error) *AuthError {
	return formatError(http.StatusInternalServerError, "error generating token", err, "TokenGeneration")
}
