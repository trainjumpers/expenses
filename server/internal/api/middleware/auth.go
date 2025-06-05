package middleware

import (
	"expenses/internal/config"
	"expenses/pkg/logger"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Protected is a middleware that checks if the request has a valid JWT token
func Protected(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" {
			logger.Warnf("Request received without authorization token")
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
			logger.Warnf("Invalid token format received")
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
		claims, err := verifyAuthToken(tokenString, cfg)
		if err != nil {
			logger.Warnf("Invalid token received: %v", err)
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
			logger.Errorf("Malformed user ID in token claims")
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
		logger.Infof("Request authenticated for user ID %d", int64(userId))
		ctx.Next()
	}
}

func verifyAuthToken(tokenString string, cfg *config.Config) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return cfg.JWTSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
