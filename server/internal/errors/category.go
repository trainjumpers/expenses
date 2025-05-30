package errors

import (
	"net/http"
)

// CategoryNotFoundError returns an error when a category is not found
func NewCategoryNotFoundError(err error) *AuthError {
	return formatError(http.StatusNotFound, "category not found", err, "CategoryNotFound")
}

// CategoryAlreadyExistsError returns an error when trying to create a category with a name that already exists for the user
func NewCategoryAlreadyExistsError(err error) *AuthError {
	return formatError(http.StatusConflict, "category with this name already exists for this user", err, "CategoryAlreadyExists")
}

// CategoryInvalidInputError returns an error for invalid category input
func NewCategoryInvalidInputError(message string, err error) *AuthError {
	return formatError(http.StatusBadRequest, message, err, "CategoryInvalidInput")
}
