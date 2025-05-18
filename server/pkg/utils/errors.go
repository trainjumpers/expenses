package utils

import (
	"errors"
	"strings"
)

var ErrNoFieldsToUpdate = errors.New("no fields to update")

// CheckForeignKey checks if the error is related to a foreign key constraint violation
func CheckForeignKey(err error, table string, fkKey string) bool {
	return strings.Contains(err.Error(), table) && strings.Contains(err.Error(), fkKey)
}
