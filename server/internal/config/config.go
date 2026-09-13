package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	EnvironmentDev  = "dev"
	EnvironmentProd = "prod"
	EnvironmentTest = "test"

	minProdJWTSecretLength = 32
)

type Config struct {
	Environment          string
	JWTSecret            []byte
	DBSchema             string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	CookieDomain         string
	LoggingLevel         string
	TrustedProxies       []string
	CORSAllowedOrigins   []string
}

func GetEnvironment() string {
	env := strings.ToLower(os.Getenv("ENV"))
	if env == "" {
		env = EnvironmentDev
	}
	return env
}

func NewConfig() (*Config, error) {
	config := &Config{}
	config.Environment = GetEnvironment()
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET environment variable is not set")
	}
	if config.IsProd() && len(jwtSecret) < minProdJWTSecretLength {
		return nil, fmt.Errorf("JWT_SECRET must be at least %d bytes in production", minProdJWTSecretLength)
	}
	config.JWTSecret = []byte(jwtSecret)
	config.DBSchema = strings.ToLower(os.Getenv("DB_SCHEMA"))
	if config.DBSchema == "" {
		return nil, errors.New("DB_SCHEMA environment variable is not set")
	}
	accessTokenHours, err := config.getEnvInt("ACCESS_TOKEN_HOURS", 12)
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_HOURS: %w", err)
	}
	config.AccessTokenDuration = time.Duration(accessTokenHours) * time.Hour
	refreshTokenDays, err := config.getEnvInt("REFRESH_TOKEN_DAYS", 7)
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_DAYS: %w", err)
	}
	config.RefreshTokenDuration = time.Duration(refreshTokenDays) * 24 * time.Hour
	config.CookieDomain = os.Getenv("COOKIE_DOMAIN")
	config.LoggingLevel = os.Getenv("LOGGING_LEVEL")
	config.TrustedProxies = getEnvList("TRUSTED_PROXIES")
	config.CORSAllowedOrigins = getEnvListWithDefault("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:3100", "https://neurospend.vercel.app"})
	return config, nil
}

// getEnvListWithDefault reads a comma-separated environment variable into a
// slice, falling back to defaultValue when unset or empty.
func getEnvListWithDefault(key string, defaultValue []string) []string {
	if items := getEnvList(key); len(items) > 0 {
		return items
	}
	return defaultValue
}

// String returns a human-readable string representation of the Config
func (cfg *Config) String() string {
	return fmt.Sprintf(
		"Config{Environment: %q, JWTSecret: %q, DBSchema: %q, AccessTokenDuration: %v, RefreshTokenDuration: %v}",
		cfg.Environment,
		"***",
		cfg.DBSchema,
		cfg.AccessTokenDuration,
		cfg.RefreshTokenDuration,
	)
}

// IsDev returns true if the environment is development
func (cfg *Config) IsDev() bool {
	return cfg.Environment == EnvironmentDev
}

// IsProd returns true if the environment is production
func (cfg *Config) IsProd() bool {
	return cfg.Environment == EnvironmentProd
}

func (cfg *Config) IsTest() bool {
	return cfg.Environment == EnvironmentTest
}

// getEnvList reads a comma-separated environment variable into a slice, dropping blanks.
func getEnvList(key string) []string {
	value := os.Getenv(key)
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

// getEnvInt is a helper function to get an integer from environment variables
func (cfg *Config) getEnvInt(key string, defaultValue int) (int, error) {
	if value := os.Getenv(key); value != "" {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return 0, fmt.Errorf("invalid integer value for %s: %w", key, err)
		}
		if intValue <= 0 {
			return 0, fmt.Errorf("%s must be greater than 0", key)
		}
		return intValue, nil
	}
	return defaultValue, nil
}
