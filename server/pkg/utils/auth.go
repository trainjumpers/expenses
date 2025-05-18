package utils

import (
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type RefreshTokenData struct {
	UserId int64
	Email  string
	Expiry time.Time
}

var refreshTokenStore = struct {
	sync.RWMutex
	Tokens map[string]RefreshTokenData
}{Tokens: make(map[string]RefreshTokenData)}

// hashPassword hashes the password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// checkPasswordHash checks if the password matches the hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// SaveRefreshToken stores a refresh token and its data
func SaveRefreshToken(token string, data RefreshTokenData) {
	refreshTokenStore.Lock()
	defer refreshTokenStore.Unlock()
	refreshTokenStore.Tokens[token] = data
}

// GetRefreshTokenData retrieves refresh token data if valid
func GetRefreshTokenData(token string) (RefreshTokenData, bool) {
	refreshTokenStore.RLock()
	defer refreshTokenStore.RUnlock()
	data, ok := refreshTokenStore.Tokens[token]
	if !ok || data.Expiry.Before(time.Now()) {
		return RefreshTokenData{}, false
	}
	return data, true
}

// DeleteRefreshToken removes a refresh token
func DeleteRefreshToken(token string) {
	refreshTokenStore.Lock()
	defer refreshTokenStore.Unlock()
	delete(refreshTokenStore.Tokens, token)
}
