package base

import (
	"expenses/internal/config"
	"fmt"
	"os"
)

// DatabaseType represents the type of database
type DatabaseType string

const (
	PostgreSQL DatabaseType = "postgres"
	// Future: MySQL, SQLite, etc.
)

// DatabaseManagerFactory creates unified database managers
type DatabaseManagerFactory interface {
	// CreateDatabaseManager creates a unified database manager with the given configuration
	CreateDatabaseManager(cfg *config.Config, managerConfig *DatabaseManagerConfig) (DatabaseManager, error)
}

// GetDatabaseType returns the database type from environment
func GetDatabaseType() DatabaseType {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		return PostgreSQL // Default to postgres
	}
	return DatabaseType(dbType)
}

// ValidateDatabaseType checks if the database type is supported
func ValidateDatabaseType(dbType DatabaseType) error {
	switch dbType {
	case PostgreSQL:
		return nil
	default:
		return fmt.Errorf("unsupported database type: %s. Only 'postgres' is supported", dbType)
	}
}
