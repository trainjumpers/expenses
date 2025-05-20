package config

import (
	"expenses/pkg/logger"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
	authOnce             sync.Once
)

// GetAccessTokenDuration returns the duration for which access tokens are valid
// Defaults to 12 hours if not set
func GetAccessTokenDuration() time.Duration {
	initAuthConfig()
	return accessTokenDuration
}

// GetRefreshTokenDuration returns the duration for which refresh tokens are valid
// Defaults to 7 days if not set
func GetRefreshTokenDuration() time.Duration {
	initAuthConfig()
	return refreshTokenDuration
}

func initAuthConfig() {
	authOnce.Do(func() {
		// Initialize access token duration
		accessTokenHours := getEnvInt("ACCESS_TOKEN_HOURS", 12)
		accessTokenDuration = time.Duration(accessTokenHours) * time.Hour
		logger.Info("Access token duration set to ", accessTokenHours, " hours")

		// Initialize refresh token duration
		refreshTokenDays := getEnvInt("REFRESH_TOKEN_DAYS", 7)
		refreshTokenDuration = time.Duration(refreshTokenDays) * 24 * time.Hour
		logger.Info("Refresh token duration set to ", refreshTokenDays, " days")
	})
}

// getEnvInt is a helper function to get an integer from environment variables
// Returns the default value if the environment variable is not set or invalid
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil && intValue > 0 {
			return intValue
		}
		logger.Warn("Invalid value for ", key, ", using default: ", defaultValue)
	}
	return defaultValue
}
