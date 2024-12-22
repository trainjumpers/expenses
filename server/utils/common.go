package utils

import (
	"errors"
	"expenses/logger"
	"os"

	"github.com/gin-gonic/gin"
)

func GetPGSchema() string {
	schema := os.Getenv("PGSCHEMA")
	if schema == "" {
		panic("PGSCHEMA environment variable is not set")
	}
	return schema
}

func GetUserIdFromContext(c *gin.Context) (int64, error) {
	userID, exists := c.Get("userID")
	if !exists {
		logger.Error("Failed to get userID from context")
		return 0, errors.New("Invalid user ID")
	}
	id, ok := userID.(int64)
	if !ok {
		logger.Error("userID is not of type int64")
		return 0, errors.New("Invalid user ID type")
	}
	return id, nil
}
