package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestKeyedLimiterAllowsUpToBurst(t *testing.T) {
	limiter := NewKeyedLimiter(60, 2)

	if !limiter.allow("a") || !limiter.allow("a") {
		t.Fatal("expected the first two requests to be allowed")
	}
	if limiter.allow("a") {
		t.Fatal("expected the third request to be rejected")
	}
}

func TestKeyedLimiterTracksKeysIndependently(t *testing.T) {
	limiter := NewKeyedLimiter(60, 1)

	if !limiter.allow("a") {
		t.Fatal("expected first key to be allowed")
	}
	if !limiter.allow("b") {
		t.Fatal("expected second key to be allowed")
	}
	if limiter.allow("a") {
		t.Fatal("expected exhausted key to be rejected")
	}
}

func TestRateLimitMiddlewareRejectsAfterBurst(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/x", RateLimit(NewKeyedLimiter(60, 1), func(*gin.Context) string { return "k" }), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	first := httptest.NewRecorder()
	router.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/x", nil))
	if first.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d", first.Code)
	}

	second := httptest.NewRecorder()
	router.ServeHTTP(second, httptest.NewRequest(http.MethodGet, "/x", nil))
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request 429, got %d", second.Code)
	}
	if second.Header().Get("Retry-After") != "60" {
		t.Fatalf("expected Retry-After header, got %q", second.Header().Get("Retry-After"))
	}
}

func TestLoginEmailKeyNormalizesAndRestoresBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	limiter := NewKeyedLimiter(60, 1)
	var boundEmail string
	router.POST("/login", RateLimit(limiter, LoginEmailKey), func(c *gin.Context) {
		var body struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			t.Fatalf("body was not readable downstream: %v", err)
		}
		boundEmail = body.Email
		c.Status(http.StatusOK)
	})

	firstReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"User@Example.com"}`))
	firstReq.Header.Set("Content-Type", "application/json")
	first := httptest.NewRecorder()
	router.ServeHTTP(first, firstReq)
	if first.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d", first.Code)
	}
	if boundEmail != "User@Example.com" {
		t.Fatalf("expected body preserved, got %q", boundEmail)
	}

	// Same email with different casing must map to the same limiter key.
	secondReq := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(`{"email":"user@example.com"}`))
	secondReq.Header.Set("Content-Type", "application/json")
	second := httptest.NewRecorder()
	router.ServeHTTP(second, secondReq)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request 429, got %d", second.Code)
	}
}
