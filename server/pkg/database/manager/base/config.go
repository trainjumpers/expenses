package base

import (
	"os"
	"time"
)

// DatabaseManagerConfig configures the unified database manager
type DatabaseManagerConfig struct {
	// Core features (always enabled)
	EnableTransactions bool // Always true
	EnableLocks        bool // Always true
	
	// Enhanced features
	EnableRetry      bool          // Enable retry policies
	EnableSavepoints bool          // Enable nested transactions
	EnableBatch      bool          // Enable batch operations
	DefaultTimeout   time.Duration // Default transaction timeout
	
	// Monitoring features
	EnableMonitoring bool // Enable performance monitoring
	EnableMetrics    bool // Enable detailed metrics collection
	
	// Connection pool optimization
	OptimizePool bool // Enable connection pool optimization
}

// DefaultConfig returns a production-ready configuration with all features enabled
func DefaultConfig() *DatabaseManagerConfig {
	return &DatabaseManagerConfig{
		// Core features (always enabled)
		EnableTransactions: true,
		EnableLocks:        true,
		
		// Enhanced features (enabled by default)
		EnableRetry:      true,
		EnableSavepoints: true,
		EnableBatch:      true,
		DefaultTimeout:   30 * time.Second,
		
		// Monitoring (enabled by default)
		EnableMonitoring: true,
		EnableMetrics:    true,
		
		// Pool optimization (enabled by default)
		OptimizePool: true,
	}
}

// BasicConfig returns a minimal configuration for simple use cases
func BasicConfig() *DatabaseManagerConfig {
	return &DatabaseManagerConfig{
		// Core features only
		EnableTransactions: true,
		EnableLocks:        true,
		
		// Enhanced features (disabled)
		EnableRetry:      false,
		EnableSavepoints: false,
		EnableBatch:      false,
		DefaultTimeout:   10 * time.Second,
		
		// Monitoring (minimal)
		EnableMonitoring: false,
		EnableMetrics:    false,
		
		// Pool optimization (basic)
		OptimizePool: false,
	}
}

// DevelopmentConfig returns a configuration optimized for development
func DevelopmentConfig() *DatabaseManagerConfig {
	return &DatabaseManagerConfig{
		// Core features
		EnableTransactions: true,
		EnableLocks:        true,
		
		// Enhanced features (selective)
		EnableRetry:      false, // Less noise during development
		EnableSavepoints: true,  // Useful for testing
		EnableBatch:      true,  // Useful for testing
		DefaultTimeout:   5 * time.Second,
		
		// Monitoring (enabled for debugging)
		EnableMonitoring: true,
		EnableMetrics:    true,
		
		// Pool optimization (development settings)
		OptimizePool: true,
	}
}

// AutoConfig returns configuration based on environment
func AutoConfig() *DatabaseManagerConfig {
	if os.Getenv("GIN_MODE") == "debug" {
		return DevelopmentConfig()
	}
	return DefaultConfig()
}
