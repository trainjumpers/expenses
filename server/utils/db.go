package utils

import (
	"os"
	"strings"
)

func GetPGSchema() string {
	schema := os.Getenv("PGSCHEMA")
	if schema == "" {
		panic("PGSCHEMA environment variable is not set")
	}
	return schema
}

func CheckForeignKey(err error, table string, fkKey string) bool {
	return strings.Contains(err.Error(), "fk_"+table) && strings.Contains(err.Error(), fkKey)
}
