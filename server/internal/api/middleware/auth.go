package middleware

import (
	"bytes"
	"encoding/json"
	"expenses/internal/config"
	"expenses/pkg/logger"
	"fmt"
	"io"
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

// InjectCreatedBy is a middleware that injects the 'created_by' field into the request body
// for POST/PUT/PATCH requests. This middleware should be used after Protected middleware.
func InjectCreatedBy() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		method := ctx.Request.Method

		// Only process POST, PUT, and PATCH requests
		if method != "POST" && method != "PUT" && method != "PATCH" {
			ctx.Next()
			return
		}

		// Get the authenticated user ID from context
		authUserId, exists := ctx.Get("authUserId")
		if !exists {
			logger.Warnf("InjectCreatedBy middleware used without authentication")
			ctx.Next()
			return
		}

		userId, ok := authUserId.(int64)
		if !ok {
			logger.Errorf("Invalid authUserId type in context")
			ctx.Next()
			return
		}

		// Read the original request body
		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			logger.Errorf("Failed to read request body: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process request"})
			ctx.Abort()
			return
		}

		// Close the original body
		ctx.Request.Body.Close()

		// Parse the JSON body
		var bodyMap map[string]any
		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
				logger.Errorf("Failed to parse JSON body: %v", err)
				// If JSON parsing fails, restore original body and continue
				ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				ctx.Next()
				return
			}
		} else {
			bodyMap = make(map[string]any)
		}

		// Inject created_by field
		bodyMap["created_by"] = userId

		// Marshal back to JSON
		modifiedBodyBytes, err := json.Marshal(bodyMap)
		if err != nil {
			logger.Errorf("Failed to marshal modified body: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to process request"})
			ctx.Abort()
			return
		}

		// Replace the request body with the modified one
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(modifiedBodyBytes))
		ctx.Request.ContentLength = int64(len(modifiedBodyBytes))

		logger.Debugf("Injected created_by field with user ID %d for %s request", userId, method)
		ctx.Next()
	}
}
