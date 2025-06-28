package middleware

import (
	"expenses/internal/config"

	"github.com/gin-gonic/gin"
)

// ProtectedWithCreatedBy returns a slice of middleware that applies both
// authentication and created_by injection. Use with route groups to avoid repetition.
func ProtectedWithCreatedBy(cfg *config.Config) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		Protected(cfg),
		InjectCreatedBy(),
	}
}
