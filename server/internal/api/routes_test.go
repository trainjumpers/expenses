package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

func TestRootRouteResponds(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := Init(&config.Config{
		Environment:        config.EnvironmentDev,
		CORSAllowedOrigins: []string{"http://localhost:3000"},
	}, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "Welcome to the expense tracker server") {
		t.Fatalf("unexpected response body: %s", recorder.Body.String())
	}
}
