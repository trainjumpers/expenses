package middleware

import (
	"expenses/internal/config"
	"expenses/internal/service"
	"expenses/pkg/logger"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Protected is a middleware that checks if the request has a valid JWT token
func Protected(authService service.AuthServiceInterface, cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" {
			logger.Warn("Request received without authorization token")
			response := gin.H{
				"message": "please log in to continue",
			}
			if cfg.IsDev() {
				response["error"] = "No authorization token provided"
			}
			ctx.JSON(http.StatusUnauthorized, response)
			ctx.Abort()
			return
		}

		tokenParts := strings.Fields(strings.TrimSpace(tokenString))
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			logger.Warn("Invalid token format received")
			response := gin.H{
				"message": "Invalid authorization format",
			}
			if cfg.IsDev() {
				response["error"] = "Token must be in format: Bearer <token>"
			}
			ctx.JSON(http.StatusBadRequest, response)
			ctx.Abort()
			return
		}
		tokenString = tokenParts[1]
		claims, err := authService.VerifyAuthToken(tokenString)
		if err != nil {
			logger.Warn("Invalid token received: ", err)
			response := gin.H{
				"message": "invalid token. please log in again",
			}
			if cfg.IsDev() {
				response["error"] = err.Error()
			}
			ctx.JSON(http.StatusUnauthorized, response)
			ctx.Abort()
			return
		}

		userId, ok := claims["user_id"].(float64)
		if !ok {
			logger.Error("Malformed user ID in token claims")
			response := gin.H{
				"message": "Something went wrong",
			}
			if cfg.IsDev() {
				response["error"] = "Malformed user ID in token claims"
			}
			ctx.JSON(http.StatusInternalServerError, response)
			ctx.Abort()
			return
		}
		ctx.Set("authUserId", int64(userId))
		logger.Info("Request authenticated for user ID: ", int64(userId))
		ctx.Next()
	}
}
