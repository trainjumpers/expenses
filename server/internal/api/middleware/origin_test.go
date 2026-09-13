package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func originTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(VerifyOrigin([]string{"http://localhost:3000", "https://neurospend.vercel.app"}))
	router.POST("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
	return router
}

func TestVerifyOriginRejectsUnknownOrigin(t *testing.T) {
	router := originTestRouter()
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestVerifyOriginAllowsKnownOrigin(t *testing.T) {
	router := originTestRouter()
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	req.Header.Set("Origin", "https://neurospend.vercel.app")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestVerifyOriginAllowsMissingOrigin(t *testing.T) {
	router := originTestRouter()
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/x", nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestVerifyOriginFallsBackToReferer(t *testing.T) {
	router := originTestRouter()
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	req.Header.Set("Referer", "https://evil.example.com/page")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", recorder.Code)
	}
}

func TestVerifyOriginIgnoresSafeMethods(t *testing.T) {
	router := originTestRouter()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}

func TestVerifyOriginAllowsRefererWithoutHost(t *testing.T) {
	router := originTestRouter()
	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	req.Header.Set("Referer", "/relative/path")
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", recorder.Code)
	}
}
