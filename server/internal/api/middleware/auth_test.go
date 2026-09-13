package middleware

import (
	"expenses/internal/config"
	"time"

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
