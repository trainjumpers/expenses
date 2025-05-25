package errors

import "net/http"

func NewUserNotFoundError(err error) *AuthError {
	return formatError(http.StatusNotFound, "user not found", err, "UserNotFound")
}

func NewUserAlreadyExistsError(err error) *AuthError {
	return formatError(http.StatusConflict, "user already exists", err, "UserAlreadyExists")
}
