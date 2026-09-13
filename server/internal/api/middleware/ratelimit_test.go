package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func TestKeyedLimiterAllowsUpToBurst(t *testing.T) {
	limiter := NewKeyedLimiter(60, 2)

	if !limiter.allow("a") {
		t.Fatal("expected the first request to be allowed")
	}
	if !limiter.allow("a") {
		t.Fatal("expected the second request to be allowed")
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

func TestKeyedLimiterEvictsStaleEntries(t *testing.T) {
	limiter := NewKeyedLimiter(60, 1)
	if !limiter.allow("fresh") {
		t.Fatal("expected fresh key to be allowed")
	}

	limiter.mu.Lock()
	limiter.limiters["stale"] = &limiterEntry{
		limiter:  rate.NewLimiter(limiter.limit, limiter.burst),
		lastSeen: time.Now().Add(-time.Hour),
	}
	limiter.mu.Unlock()

	limiter.evictStaleOnce()

	if _, ok := limiter.limiters["stale"]; ok {
		t.Fatal("expected stale key to be evicted")
	}
	if _, ok := limiter.limiters["fresh"]; !ok {
		t.Fatal("expected fresh key to remain")
	}
}

func TestClientIPKeyUsesRemoteAddress(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
	ctx.Request = httptest.NewRequest(http.MethodPost, "/login", nil)
	ctx.Request.RemoteAddr = "203.0.113.7:1234"

	if got := ClientIPKey(ctx); got != "ip:203.0.113.7" {
		t.Fatalf("expected remote address key, got %q", got)
	}
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("read failure")
}

func TestLoginEmailKeyFallsBackToClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name string
		body io.Reader
	}{
		{"invalid json", strings.NewReader("not-json")},
		{"missing email", strings.NewReader(`{"password":"x"}`)},
		{"whitespace email", strings.NewReader(`{"email":"   "}`)},
		{"read failure", io.NopCloser(failingReader{})},
		{"nil body", nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest(http.MethodPost, "/login", tc.body)
			if tc.body == nil {
				ctx.Request.Body = nil
			}
			ctx.Request.RemoteAddr = "198.51.100.9:5555"

			if got := LoginEmailKey(ctx); got != "ip:198.51.100.9" {
				t.Fatalf("expected IP fallback, got %q", got)
			}
		})
	}
}
