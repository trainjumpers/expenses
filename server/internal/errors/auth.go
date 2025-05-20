package errors

import "net/http"

func NewInvalidCredentialsError(err error) *AuthError {
	return FormatError(http.StatusUnauthorized, "invalid credentials", err, "InvalidCredentials")
}

func NewInvalidTokenError(err error) *AuthError {
	return FormatError(http.StatusUnauthorized, "invalid token", err, "InvalidToken")
}

func NewTokenGenerationError(err error) *AuthError {
	return FormatError(http.StatusInternalServerError, "error generating token", err, "TokenGeneration")
}
