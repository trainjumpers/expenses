package postgres

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// generateTransactionID creates a unique transaction ID
func generateTransactionID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// getCurrentTime returns the current time (useful for testing)
func getCurrentTime() time.Time {
	return time.Now()
}
