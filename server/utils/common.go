package utils

import (
	"os"
)

func GetPGSchema() string {
	schema := os.Getenv("PGSCHEMA")
	if schema == "" {
		panic("PGSCHEMA environment variable is not set")
	}
	return schema
}
