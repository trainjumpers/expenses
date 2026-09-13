package middleware

import (
	"expenses/internal/config"
	"expenses/pkg/logger"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Protected is a middleware that checks if the request has a valid JWT token from HTTP-only cookie
func Protected(cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookieToken, err := ctx.Cookie("access_token")
		if err != nil || cookieToken == "" {
			logger.Warnf("Request received without access_token cookie")
			response := gin.H{
				"message": "please log in to continue",
			}
			if cfg.IsDev() {
				response["error"] = "No access_token cookie provided"
			}
			ctx.JSON(http.StatusUnauthorized, response)
			ctx.Abort()
			return
		}

		claims, err := verifyAuthToken(cookieToken, cfg)
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
			logger.Errorf("Malformed user Id in token claims")
			response := gin.H{
				"message": "Something went wrong",
			}
			if cfg.IsDev() {
				response["error"] = "Malformed user Id in token claims"
			}
			ctx.JSON(http.StatusInternalServerError, response)
			ctx.Abort()
			return
		}
		ctx.Set("authUserId", int64(userId))
		logger.Debugf("Request authenticated for user Id %d", int64(userId))
		ctx.Next()
	}
}

func verifyAuthToken(tokenString string, cfg *config.Config) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
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
