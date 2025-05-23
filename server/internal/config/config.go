package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Environment          string
	JWTSecret            []byte
	DBSchema             string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

func NewConfig() (*Config, error) {
	config := &Config{}
	config.Environment = strings.ToLower(os.Getenv("ENV"))
	if config.Environment == "" {
		config.Environment = "dev"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET environment variable is not set")
	}
	config.JWTSecret = []byte(jwtSecret)
	config.DBSchema = strings.ToLower(os.Getenv("DB_SCHEMA"))
	if config.DBSchema == "" {
		return nil, errors.New("DB_SCHEMA environment variable is not set")
	}
	accessTokenHours, err := getEnvInt("ACCESS_TOKEN_HOURS", 12)
	if err != nil {
		return nil, fmt.Errorf("invalid ACCESS_TOKEN_HOURS: %w", err)
	}
	config.AccessTokenDuration = time.Duration(accessTokenHours) * time.Hour
	refreshTokenDays, err := getEnvInt("REFRESH_TOKEN_DAYS", 7)
	if err != nil {
		return nil, fmt.Errorf("invalid REFRESH_TOKEN_DAYS: %w", err)
	}
	config.RefreshTokenDuration = time.Duration(refreshTokenDays) * 24 * time.Hour
	return config, nil
}

// IsDev returns true if the environment is development
func (cfg *Config) IsDev() bool {
	return cfg.Environment == "dev"
}

// IsProd returns true if the environment is production
func (cfg *Config) IsProd() bool {
	return cfg.Environment == "prod"
}

// getEnvInt is a helper function to get an integer from environment variables
func getEnvInt(key string, defaultValue int) (int, error) {
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
