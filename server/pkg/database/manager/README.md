# Database Manager

This package provides a **unified, feature-complete database management layer** with configurable capabilities for different use cases.

## 🎯 Unified Approach

Instead of choosing between multiple managers, you now get **one powerful manager** with configurable features:

```go
// Simple - uses smart defaults based on environment
dbManager, err := manager.NewDatabaseManager(cfg)

// Custom configuration
dbManager, err := manager.NewDatabaseManagerWithConfig(cfg, manager.DefaultConfig())

// Convenience functions
basicDB, err := manager.NewBasicDatabaseManager(cfg)      // Minimal features
devDB, err := manager.NewDevelopmentDatabaseManager(cfg)  // Dev-optimized
prodDB, err := manager.NewProductionDatabaseManager(cfg)  // All features
```

## 🏗️ Architecture

```
manager/
├── base/           # Common interfaces and configuration
│   ├── interface.go    # Unified DatabaseManager interface
│   ├── config.go       # Configuration options
│   ├── context.go      # Transaction context management
│   ├── options.go      # Transaction options
│   ├── monitoring.go   # Monitoring types
│   └── factory.go      # Database type validation
├── postgres/       # PostgreSQL implementation
│   ├── postgres.go     # Single unified implementation
│   ├── factory.go      # PostgreSQL factory
│   ├── pool.go         # Connection pool optimization
│   └── utils.go        # Utility functions
└── manager.go      # Main entry point
```

## ⚙️ Configuration Options

### DatabaseManagerConfig

```go
type DatabaseManagerConfig struct {
    // Core features (always enabled)
    EnableTransactions bool // Always true
    EnableLocks        bool // Always true
    
    // Enhanced features
    EnableRetry      bool          // Retry policies
    EnableSavepoints bool          // Nested transactions
    EnableBatch      bool          // Batch operations
    DefaultTimeout   time.Duration // Transaction timeout
    
    // Monitoring features
    EnableMonitoring bool // Performance monitoring
    EnableMetrics    bool // Detailed metrics
    
    // Connection pool optimization
    OptimizePool bool // Pool optimization
}
```

### Pre-configured Options

```go
// Production-ready (all features enabled)
manager.DefaultConfig()

// Minimal features only
manager.BasicConfig()

// Development-optimized
manager.DevelopmentConfig()

// Auto-detect based on GIN_MODE environment
manager.AutoConfig()
```

## 🚀 Usage Examples

### Basic Usage (Recommended)

```go
import "expenses/pkg/database/manager"

// Smart defaults - automatically configures based on environment
dbManager, err := manager.NewDatabaseManager(cfg)
if err != nil {
    log.Fatal(err)
}
defer dbManager.Close()

// All core operations work the same
rowsAffected, err := dbManager.ExecuteQuery(ctx, 
    "INSERT INTO users (name, email) VALUES ($1, $2)", 
    "John", "john@example.com")

// Transaction
err = dbManager.WithTxn(ctx, func(txCtx context.Context) error {
    _, err := dbManager.ExecuteQuery(txCtx, "INSERT INTO accounts (user_id, name) VALUES ($1, $2)", userID, "Savings")
    return err
})
```

### Advanced Features (When Enabled)

```go
// Enhanced features are available if enabled in config
if dbManager.IsFeatureEnabled(manager.FeatureRetry) {
    err = dbManager.WithRetryableTxn(ctx, func(txCtx context.Context) error {
        return criticalOperation(txCtx)
    })
}

// Nested transactions with savepoints
if dbManager.IsFeatureEnabled(manager.FeatureSavepoints) {
    err = dbManager.WithTxn(ctx, func(txCtx context.Context) error {
        // Main transaction work
        
        return dbManager.WithSavepoint(txCtx, "checkpoint1", func(spCtx context.Context) error {
            // This can fail without affecting the main transaction
            return riskyOperation(spCtx)
        })
    })
}

// Batch operations
if dbManager.IsFeatureEnabled(manager.FeatureBatch) {
    batch := &pgx.Batch{}
    batch.Queue("INSERT INTO logs (message) VALUES ($1)", "Log 1")
    batch.Queue("INSERT INTO logs (message) VALUES ($1)", "Log 2")
    err = dbManager.ExecuteBatch(ctx, batch)
}
```

### Monitoring (When Enabled)

```go
// Check if monitoring is available
if dbManager.IsFeatureEnabled(manager.FeatureMonitoring) {
    // Get performance metrics
    metrics := dbManager.GetMonitoringMetrics()
    fmt.Printf("Success rate: %.2f%%", 
        float64(metrics.CommittedTransactions)/float64(metrics.TotalTransactions)*100)
    fmt.Printf("Average duration: %v", metrics.AverageDuration)
    fmt.Printf("Active transactions: %d", metrics.ActiveTransactions)
}
```

### Custom Configuration

```go
// Create custom configuration
config := &manager.DatabaseManagerConfig{
    EnableTransactions: true,
    EnableLocks:        true,
    EnableRetry:        true,
    EnableSavepoints:   false, // Disable savepoints
    EnableBatch:        true,
    EnableMonitoring:   true,
    EnableMetrics:      false, // Disable detailed metrics
    DefaultTimeout:     15 * time.Second,
    OptimizePool:       true,
}

dbManager, err := manager.NewDatabaseManagerWithConfig(cfg, config)
```

## 🎛️ Feature Detection

```go
// Check what features are available
if dbManager.IsFeatureEnabled("retry") {
    // Use retry features
}

if dbManager.IsFeatureEnabled("monitoring") {
    // Access monitoring features
}

// Get current configuration
config := dbManager.GetConfig()
fmt.Printf("Retry enabled: %v", config.EnableRetry)
```

## 📊 All Features Available

### ✅ Core Operations (Always Available)
- `ExecuteQuery()` - Execute INSERT/UPDATE/DELETE queries
- `FetchOne()` - Fetch single row
- `FetchAll()` - Fetch multiple rows
- `WithTxn()` - Execute within transaction
- `WithLock()` - Execute with advisory lock
- `Close()` - Close database connections

### ⚡ Enhanced Operations (Configurable)
- `WithTxnOptions()` - Transactions with custom options (timeout, isolation, retry)
- `WithReadOnlyTxn()` - Read-only transactions
- `WithRetryableTxn()` - Transactions with aggressive retry policy
- `WithSavepoint()` - Nested transactions using savepoints
- `ExecuteBatch()` - Batch operations
- `WithConnection()` - Execute with dedicated connection

### 🔍 Health & Introspection (Always Available)
- `Ping()` - Health check
- `Stats()` - Connection pool statistics
- `GetTransactionInfo()` - Transaction introspection

### 📈 Monitoring (Configurable)
- `GetMonitoringMetrics()` - Transaction performance metrics
- `ResetMetrics()` - Clear monitoring data

### ⚙️ Configuration (Always Available)
- `GetConfig()` - Get current configuration
- `IsFeatureEnabled()` - Check if feature is enabled

## 🌟 Benefits of Unified Approach

### ✅ **Simplicity**
- **One manager to rule them all** - no more choosing between 3 options
- **Smart defaults** - works great out of the box
- **Progressive enhancement** - enable features as needed

### ✅ **Flexibility**
- **Configurable features** - enable only what you need
- **Environment-aware** - different configs for dev/prod
- **Runtime feature detection** - check what's available

### ✅ **Performance**
- **Zero overhead** - disabled features don't impact performance
- **Optimized connection pooling** - environment-specific tuning
- **Efficient monitoring** - only when enabled

### ✅ **Backward Compatibility**
- **Existing code works** - no breaking changes
- **Gradual migration** - upgrade at your own pace
- **Legacy support** - deprecated functions still work

## 🔧 Migration Guide

### From Old Approach
```go
// OLD - Multiple managers to choose from
basicDB := manager.NewDatabaseManager(cfg)           // ❌ Confusing
enhancedDB := manager.NewEnhancedDatabaseManager(cfg) // ❌ Too many choices  
monitoredDB := manager.NewMonitoredDatabaseManager(cfg) // ❌ Decision fatigue

// NEW - One manager with smart defaults
dbManager := manager.NewDatabaseManager(cfg)          // ✅ Simple & powerful
```

### Environment-Based Configuration
```go
// Automatically configures based on GIN_MODE
dbManager, err := manager.NewDatabaseManager(cfg)

// In development (GIN_MODE=debug):
// - Basic retry disabled (less noise)
// - Monitoring enabled (for debugging)
// - Savepoints enabled (for testing)

// In production (GIN_MODE=release):
// - All features enabled
// - Optimized connection pool
// - Full monitoring
```

## 📚 Configuration Examples

### Development Setup
```go
config := manager.DevelopmentConfig()
// - Core features: ✅
// - Retry: ❌ (less noise)
// - Savepoints: ✅ (useful for testing)
// - Batch: ✅ (useful for testing)
// - Monitoring: ✅ (debugging)
// - Pool optimization: ✅
```

### Production Setup
```go
config := manager.DefaultConfig()
// - All features: ✅
// - Full monitoring: ✅
// - Optimized performance: ✅
// - Maximum reliability: ✅
```

### Minimal Setup
```go
config := manager.BasicConfig()
// - Core features only: ✅
// - Enhanced features: ❌
// - Monitoring: ❌
// - Lightweight: ✅
```

## 🎯 Key Improvements

1. **🎯 Single Interface** - One manager with all features
2. **⚙️ Configurable** - Enable only what you need
3. **🚀 Smart Defaults** - Works great out of the box
4. **📊 Feature Detection** - Runtime capability checking
5. **🔄 Backward Compatible** - Existing code continues to work
6. **🌍 Environment Aware** - Different configs for different environments
7. **⚡ Zero Overhead** - Disabled features don't impact performance

---

**The unified database manager provides all the power you need with the simplicity you want. One manager, infinite possibilities!** 🚀
