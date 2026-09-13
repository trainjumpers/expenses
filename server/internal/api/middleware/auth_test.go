package middleware

import (
	"expenses/internal/config"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("verifyAuthToken", func() {
	var cfg *config.Config

	BeforeEach(func() {
		cfg = &config.Config{
			Environment: config.EnvironmentTest,
			JWTSecret:   []byte("01234567890123456789012345678901"),
		}
	})

	sign := func(method jwt.SigningMethod, claims jwt.MapClaims, secret []byte) string {
		token := jwt.NewWithClaims(method, claims)
		signed, err := token.SignedString(secret)
		Expect(err).NotTo(HaveOccurred())
		return signed
	}

	validClaims := func() jwt.MapClaims {
		return jwt.MapClaims{
			"user_id": float64(42),
			"exp":     time.Now().Add(time.Hour).Unix(),
		}
	}

	It("accepts a valid HS256 token", func() {
		claims, err := verifyAuthToken(sign(jwt.SigningMethodHS256, validClaims(), cfg.JWTSecret), cfg)
		Expect(err).NotTo(HaveOccurred())
		Expect(claims["user_id"]).To(Equal(float64(42)))
	})

	It("rejects a token signed with a different HMAC method", func() {
		_, err := verifyAuthToken(sign(jwt.SigningMethodHS512, validClaims(), cfg.JWTSecret), cfg)
		Expect(err).To(HaveOccurred())
	})

	It("rejects a token signed with the wrong secret", func() {
		_, err := verifyAuthToken(sign(jwt.SigningMethodHS256, validClaims(), []byte("wrong-secret")), cfg)
		Expect(err).To(HaveOccurred())
	})

	It("rejects a token without an expiry", func() {
		claims := jwt.MapClaims{"user_id": float64(42)}
		_, err := verifyAuthToken(sign(jwt.SigningMethodHS256, claims, cfg.JWTSecret), cfg)
		Expect(err).To(HaveOccurred())
	})

	It("rejects an expired token", func() {
		claims := jwt.MapClaims{
			"user_id": float64(42),
			"exp":     time.Now().Add(-time.Hour).Unix(),
		}
		_, err := verifyAuthToken(sign(jwt.SigningMethodHS256, claims, cfg.JWTSecret), cfg)
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("Protected middleware", func() {
	var (
		cfg    *config.Config
		router *gin.Engine
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		cfg = &config.Config{
			Environment: config.EnvironmentDev,
			JWTSecret:   []byte("01234567890123456789012345678901"),
		}
		router = gin.New()
		router.GET("/protected", Protected(cfg), func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"authUserId": ctx.GetInt64("authUserId")})
		})
	})

	sign := func(claims jwt.MapClaims) string {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(cfg.JWTSecret)
		Expect(err).NotTo(HaveOccurred())
		return signed
	}

	request := func(token string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		if token != "" {
			req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
		}
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)
		return recorder
	}

	It("rejects a request without the access token cookie", func() {
		recorder := request("")

		Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		Expect(recorder.Body.String()).To(ContainSubstring("No access_token cookie provided"))
	})

	It("rejects an invalid token and includes the error in dev", func() {
		recorder := request("invalid-token")

		Expect(recorder.Code).To(Equal(http.StatusUnauthorized))
		Expect(recorder.Body.String()).To(ContainSubstring("invalid token. please log in again"))
		Expect(recorder.Body.String()).To(ContainSubstring(`"error"`))
	})

	It("accepts a valid token and sets the authenticated user id", func() {
		recorder := request(sign(jwt.MapClaims{"user_id": float64(42), "exp": time.Now().Add(time.Hour).Unix()}))

		Expect(recorder.Code).To(Equal(http.StatusOK))
		Expect(recorder.Body.String()).To(ContainSubstring("42"))
	})

	It("rejects a token without a numeric user id claim", func() {
		recorder := request(sign(jwt.MapClaims{"exp": time.Now().Add(time.Hour).Unix()}))

		Expect(recorder.Code).To(Equal(http.StatusInternalServerError))
		Expect(recorder.Body.String()).To(ContainSubstring("Malformed user Id in token claims"))
	})
})
