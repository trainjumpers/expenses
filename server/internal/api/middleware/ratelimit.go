package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// KeyedLimiter rate limits independently per key (for example per email or per IP).
type KeyedLimiter struct {
	mu       sync.Mutex
	limiters map[string]*limiterEntry
	limit    rate.Limit
	burst    int
	ttl      time.Duration
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

func NewKeyedLimiter(perMinute, burst int) *KeyedLimiter {
	limiter := &KeyedLimiter{
		limiters: make(map[string]*limiterEntry),
		limit:    rate.Limit(float64(perMinute) / 60.0),
		burst:    burst,
		ttl:      10 * time.Minute,
	}
	go limiter.evictStale()
	return limiter
}

func (l *KeyedLimiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	entry, ok := l.limiters[key]
	if !ok {
		entry = &limiterEntry{limiter: rate.NewLimiter(l.limit, l.burst)}
		l.limiters[key] = entry
	}
	entry.lastSeen = time.Now()
	return entry.limiter.Allow()
}

func (l *KeyedLimiter) evictStale() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		l.mu.Lock()
		for key, entry := range l.limiters {
			if time.Since(entry.lastSeen) > l.ttl {
				delete(l.limiters, key)
			}
		}
		l.mu.Unlock()
	}
}

// RateLimit rejects requests whose key exceeds the limiter budget with 429.
func RateLimit(limiter *KeyedLimiter, keyFn func(*gin.Context) string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !limiter.allow(keyFn(ctx)) {
			ctx.Header("Retry-After", "60")
			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "too many attempts, please try again later",
			})
			return
		}
		ctx.Next()
	}
}

// ClientIPKey keys a limiter by the requesting client address.
func ClientIPKey(ctx *gin.Context) string {
	return "ip:" + ctx.ClientIP()
}

// LoginEmailKey keys a limiter by the submitted email, falling back to the client address.
// The request body is read and restored so downstream binding still works.
func LoginEmailKey(ctx *gin.Context) string {
	if ctx.Request.Body == nil {
		return ClientIPKey(ctx)
	}
	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		return ClientIPKey(ctx)
	}
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var body struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(bodyBytes, &body); err != nil || strings.TrimSpace(body.Email) == "" {
		return ClientIPKey(ctx)
	}
	return "email:" + strings.ToLower(strings.TrimSpace(body.Email))
}
