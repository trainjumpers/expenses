package postgres

import (
	"expenses/internal/config"
	"expenses/pkg/database/manager/base"
)

// PostgreSQLFactory implements DatabaseManagerFactory for PostgreSQL
type PostgreSQLFactory struct{}

// NewPostgreSQLFactory creates a new PostgreSQL factory
func NewPostgreSQLFactory() *PostgreSQLFactory {
	return &PostgreSQLFactory{}
}

// CreateDatabaseManager creates a unified PostgreSQL database manager
func (f *PostgreSQLFactory) CreateDatabaseManager(cfg *config.Config, managerConfig *base.DatabaseManagerConfig) (base.DatabaseManager, error) {
	pool, err := createConnectionPool(cfg)
	if err != nil {
		return nil, err
	}
	
	return NewPostgresDatabaseManager(pool, managerConfig), nil
}
