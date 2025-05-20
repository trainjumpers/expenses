package config

import (
	"os"
	"sync"
)

var (
	jwtSecret []byte
	once      sync.Once
)

// GetJWTSecret returns the JWT secret key, initializing it if not already done
func GetJWTSecret() []byte {
	once.Do(func() {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			panic("JWT_SECRET environment variable is not set")
		}
		jwtSecret = []byte(secret)
	})
	return jwtSecret
}
