package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaxBodySize caps the request body so oversized uploads are rejected while
// reading, before the handler buffers them in memory. Handlers detect the
// resulting *http.MaxBytesError and respond with 413.
func MaxBodySize(maxBytes int64) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Body != nil {
			ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxBytes)
		}
		ctx.Next()
	}
}
