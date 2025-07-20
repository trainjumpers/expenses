package errors

import (
	"fmt"
	"net/http"
)

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

// Custom CSV Import Errors
func NewInvalidCSVFormatError(err error) *AuthError {
	return formatError(http.StatusBadRequest, "invalid CSV file format", err, "InvalidCSVFormat")
}

func NewCSVFileTooLargeError() *AuthError {
	return formatError(http.StatusBadRequest, "CSV file size exceeds 256KB limit", nil, "CSVFileTooLarge")
}

func NewInsufficientColumnsError() *AuthError {
	return formatError(http.StatusBadRequest, "CSV file must have at least one column", nil, "InsufficientColumns")
}

func NewMissingRequiredFieldError(field string) *AuthError {
	return formatError(http.StatusBadRequest, fmt.Sprintf("required field '%s' not mapped", field), nil, "MissingRequiredField")
}

func NewInvalidDateFormatError(err error) *AuthError {
	return formatError(http.StatusBadRequest, "invalid date format in CSV", err, "InvalidDateFormat")
}

func NewInvalidAmountFormatError(err error) *AuthError {
	return formatError(http.StatusBadRequest, "invalid amount format in CSV", err, "InvalidAmountFormat")
}

func NewDuplicateMappingError(field string) *AuthError {
	return formatError(http.StatusBadRequest, fmt.Sprintf("multiple columns mapped to field '%s'", field), nil, "DuplicateMapping")
}
