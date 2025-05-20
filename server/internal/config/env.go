package config

import (
	"os"
	"strings"
	"sync"
)

var (
	env     string
	envOnce sync.Once
)

// GetEnv returns the current environment, initializing it if not already done
// Returns "dev" if ENV is not set
func GetEnv() string {
	envOnce.Do(func() {
		env = strings.ToLower(os.Getenv("ENV"))
		if env == "" {
			env = "dev"
		}
	})
	return env
}

func IsDev() bool {
	return !IsProd()
}

func IsProd() bool {
	return GetEnv() == "prod"
}
