package manager

import (
	"fmt"

	"expenses/internal/config"
	"expenses/pkg/database/manager/base"
	"expenses/pkg/database/manager/postgres"
)

// NewDatabaseManager creates a unified database manager with default configuration
func NewDatabaseManager(cfg *config.Config) (base.DatabaseManager, error) {
	return NewDatabaseManagerWithConfig(cfg, base.AutoConfig())
}

// NewDatabaseManagerWithConfig creates a unified database manager with custom configuration
func NewDatabaseManagerWithConfig(cfg *config.Config, managerConfig *base.DatabaseManagerConfig) (base.DatabaseManager, error) {
	dbType := base.GetDatabaseType()

	if err := base.ValidateDatabaseType(dbType); err != nil {
		return nil, err
	}

	switch dbType {
	case base.PostgreSQL:
		factory := postgres.NewPostgreSQLFactory()
		return factory.CreateDatabaseManager(cfg, managerConfig)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}

// NewBasicDatabaseManager creates a database manager with minimal features
func NewBasicDatabaseManager(cfg *config.Config) (base.DatabaseManager, error) {
	return NewDatabaseManagerWithConfig(cfg, base.BasicConfig())
}

// NewDevelopmentDatabaseManager creates a database manager optimized for development
func NewDevelopmentDatabaseManager(cfg *config.Config) (base.DatabaseManager, error) {
	return NewDatabaseManagerWithConfig(cfg, base.DevelopmentConfig())
}

// NewProductionDatabaseManager creates a database manager with all features enabled
func NewProductionDatabaseManager(cfg *config.Config) (base.DatabaseManager, error) {
	return NewDatabaseManagerWithConfig(cfg, base.DefaultConfig())
}

// DatabaseManager is an alias for the unified interface
type DatabaseManager = base.DatabaseManager

// TransactionFunc is an alias for the transaction function type
type TransactionFunc = base.TransactionFunc

// LockFunc is an alias for the lock function type
type LockFunc = base.LockFunc

// Configuration types
type DatabaseManagerConfig = base.DatabaseManagerConfig

// Configuration functions
var (
	DefaultConfig     = base.DefaultConfig
	BasicConfig       = base.BasicConfig
	DevelopmentConfig = base.DevelopmentConfig
	AutoConfig        = base.AutoConfig
)

// Feature constants
const (
	FeatureRetry      = base.FeatureRetry
	FeatureSavepoints = base.FeatureSavepoints
	FeatureBatch      = base.FeatureBatch
	FeatureMonitoring = base.FeatureMonitoring
	FeatureMetrics    = base.FeatureMetrics
)
