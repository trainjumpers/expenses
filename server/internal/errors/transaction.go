package errors

import "net/http"

func NewTransactionNotFoundError(err error) *AuthError {
	return formatError(http.StatusNotFound, "transaction not found", err, "TransactionNotFound")
}

func NewTransactionAlreadyExistsError(err error) *AuthError {
	return formatError(http.StatusConflict, "transaction already exists", err, "TransactionAlreadyExists")
}

func NewTransactionDateInFutureError(err error) *AuthError {
	return formatError(http.StatusBadRequest, "transaction date cannot be in the future", err, "TransactionDateInFuture")
}
