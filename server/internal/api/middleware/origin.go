package middleware

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// VerifyOrigin rejects state-changing requests whose Origin (or Referer) is not
// in the allow-list. Requests without either header (non-browser clients) are
// allowed, so API consumers keep working. Safe methods are always allowed.
func VerifyOrigin(allowed []string) gin.HandlerFunc {
	allowedSet := make(map[string]struct{}, len(allowed))
	for _, origin := range allowed {
		allowedSet[origin] = struct{}{}
	}

	return func(ctx *gin.Context) {
		switch ctx.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			ctx.Next()
			return
		}

		origin := ctx.GetHeader("Origin")
		if origin == "" {
			origin = refererOrigin(ctx.GetHeader("Referer"))
		}
		if origin == "" {
			ctx.Next()
			return
		}

		if _, ok := allowedSet[origin]; !ok {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "cross-origin request blocked"})
			return
		}
		ctx.Next()
	}
}

func refererOrigin(referer string) string {
	if referer == "" {
		return ""
	}
	parsed, err := url.Parse(referer)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}
