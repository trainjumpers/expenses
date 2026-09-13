package api

import (
	"net/http"
	"testing"

	"expenses/internal/config"

	"github.com/gin-gonic/gin"
)

func TestInitRegistersAuthRoutesOutsideTest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := Init(&config.Config{
		Environment:        config.EnvironmentDev,
		CORSAllowedOrigins: []string{"http://localhost:3000"},
	}, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if !hasRoute(router, http.MethodPost, "/api/v1/login") {
		t.Fatal("expected login route to be registered")
	}

	// An invalid proxy list must not abort startup; proxy headers are ignored instead.
	router = Init(&config.Config{
		Environment:        config.EnvironmentDev,
		TrustedProxies:     []string{"not-an-ip"},
		CORSAllowedOrigins: []string{"http://localhost:3000"},
	}, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if !hasRoute(router, http.MethodPost, "/api/v1/login") {
		t.Fatal("expected login route to be registered with invalid trusted proxies")
	}
}

func hasRoute(router *gin.Engine, method, path string) bool {
	for _, route := range router.Routes() {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}
