package middleware

import (
	"expenses/internal/service"
	"expenses/pkg/logger"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Protected is a middleware that checks if the request has a valid JWT token
func Protected(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No token provided"})
			c.Abort()
			return
		}

		tokenString = strings.Split(tokenString, " ")[1]
		claims, err := authService.VerifyAuthToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		userId, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Malformed User Id"})
			c.Abort()
			return
		}
		c.Set("authUserId", int64(userId))
		logger.Info("Received request from user with Id: ", int64(userId))
		c.Next()
	}
}
