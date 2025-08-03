package postgres

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"expenses/internal/config"
	"expenses/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig provides optimized connection pool configuration
type PoolConfig struct {
	// Connection limits
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration

	// Health check
	HealthCheckPeriod time.Duration

	// Timeouts
	ConnectTimeout time.Duration
	AcquireTimeout time.Duration

	// Application specific
	ApplicationName string

	// Performance tuning
	PreferSimpleProtocol   bool
	StatementCacheCapacity int
}

// OptimizedPoolConfig returns a production-optimized pool configuration
func OptimizedPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxConns:               25, // Reasonable default for most apps
		MinConns:               5,  // Keep some connections warm
		MaxConnLifetime:        time.Hour,
		MaxConnIdleTime:        30 * time.Minute,
		HealthCheckPeriod:      5 * time.Minute,
		ConnectTimeout:         10 * time.Second,
		AcquireTimeout:         5 * time.Second,
		ApplicationName:        "neurospend",
		PreferSimpleProtocol:   false,
		StatementCacheCapacity: 512,
	}
}

// DevelopmentPoolConfig returns a development-optimized configuration
func DevelopmentPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxConns:               10,
		MinConns:               2,
		MaxConnLifetime:        30 * time.Minute,
		MaxConnIdleTime:        10 * time.Minute,
		HealthCheckPeriod:      2 * time.Minute,
		ConnectTimeout:         5 * time.Second,
		AcquireTimeout:         3 * time.Second,
		ApplicationName:        "neurospend-dev",
		PreferSimpleProtocol:   true, // Simpler for debugging
		StatementCacheCapacity: 128,
	}
}

// createConnectionPool creates an optimized PostgreSQL connection pool
func createConnectionPool(cfg *config.Config) (*pgxpool.Pool, error) {
	host := os.Getenv("DB_HOST")
	port, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		return nil, fmt.Errorf("invalid database port number: %w", err)
	}
	user := os.Getenv("DB_USER")
	dbname := os.Getenv("DB_NAME")
	pass := os.Getenv("DB_PASSWORD")
	sslmode := os.Getenv("DB_SSL_MODE")
	if sslmode == "" {
		sslmode = "verify-full"
	}

	// Determine environment and use appropriate pool config
	var poolConfig *PoolConfig
	if os.Getenv("GIN_MODE") == "debug" {
		poolConfig = DevelopmentPoolConfig()
	} else {
		poolConfig = OptimizedPoolConfig()
	}

	// Build connection string with pool parameters
	connStr := buildConnectionString(host, user, pass, dbname, port, cfg.DBSchema, sslmode, poolConfig)

	logger.Debugf("Connecting to database with optimized pool configuration")
	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Debugf("Database connected successfully with pool stats: max=%d, min=%d",
		poolConfig.MaxConns, poolConfig.MinConns)

	return pool, nil
}

// buildConnectionString creates an optimized connection string
func buildConnectionString(host, user, password, dbname string, port int, schema, sslmode string, config *PoolConfig) string {
	connStr := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=%s&search_path=%s&application_name=%s",
		user, password, host, port, dbname, sslmode, schema, config.ApplicationName,
	)

	// Add pool parameters
	connStr += fmt.Sprintf("&pool_max_conns=%d", config.MaxConns)
	connStr += fmt.Sprintf("&pool_min_conns=%d", config.MinConns)
	connStr += fmt.Sprintf("&pool_max_conn_lifetime=%s", config.MaxConnLifetime)
	connStr += fmt.Sprintf("&pool_max_conn_idle_time=%s", config.MaxConnIdleTime)
	connStr += fmt.Sprintf("&pool_health_check_period=%s", config.HealthCheckPeriod)

	// Add performance parameters
	if config.PreferSimpleProtocol {
		connStr += "&prefer_simple_protocol=true"
	}

	return connStr
}
