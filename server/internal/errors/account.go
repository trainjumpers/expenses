package errors

import "net/http"

func NewAccountNotFoundError(err error) *AuthError {
	return formatError(http.StatusNotFound, "account not found", err, "AccountNotFound")
}

func NewAccountHasTransactionsError(err error) *AuthError {
	return formatError(http.StatusConflict, "cannot delete account with existing transactions", err, "AccountHasTransactions")
}
