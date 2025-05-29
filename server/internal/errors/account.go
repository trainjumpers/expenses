package errors

import "net/http"

func NewAccountNotFoundError(err error) *AuthError {
	return formatError(http.StatusNotFound, "account not found", err, "AccountNotFound")
}
