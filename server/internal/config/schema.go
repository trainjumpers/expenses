package config

import (
	"os"
	"strings"
	"sync"
)

var (
	schema     string
	schemaOnce sync.Once
)

// GetSchema returns the current environment, initializing it if not already done
// Returns "dev" if ENV is not set
func GetSchema() string {
	schemaOnce.Do(func() {
		schema = strings.ToLower(os.Getenv("DB_SCHEMA"))
		if schema == "" {
			panic("DB_SCHEMA environment variable is not set")
		}
	})
	return schema
}
