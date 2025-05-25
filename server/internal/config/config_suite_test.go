package config

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config", func() {
	var envVars map[string]string

	BeforeEach(func() {
		// Clear all relevant environment variables before each test
		os.Unsetenv("ENV")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("DB_SCHEMA")
		os.Unsetenv("ACCESS_TOKEN_HOURS")
		os.Unsetenv("REFRESH_TOKEN_DAYS")
	})

	Context("when creating a new config", func() {
		BeforeEach(func() {
			envVars = map[string]string{
				"JWT_SECRET": "test-secret",
				"DB_SCHEMA":  "test_schema",
			}
			for key, value := range envVars {
				os.Setenv(key, value)
			}
		})

		It("should create a valid config with default values", func() {
			cfg, err := NewConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Environment).To(Equal("dev"))
			Expect(string(cfg.JWTSecret)).To(Equal("test-secret"))
			Expect(cfg.DBSchema).To(Equal("test_schema"))
			Expect(cfg.AccessTokenDuration).To(Equal(12 * time.Hour))
			Expect(cfg.RefreshTokenDuration).To(Equal(7 * 24 * time.Hour))
		})

		It("should create a config with custom environment and token durations", func() {
			os.Setenv("ENV", "prod")
			os.Setenv("ACCESS_TOKEN_HOURS", "24")
			os.Setenv("REFRESH_TOKEN_DAYS", "30")

			cfg, err := NewConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.Environment).To(Equal("prod"))
			Expect(cfg.AccessTokenDuration).To(Equal(24 * time.Hour))
			Expect(cfg.RefreshTokenDuration).To(Equal(30 * 24 * time.Hour))
		})
	})

	Context("when required environment variables are missing", func() {
		It("should return error when JWT_SECRET is missing", func() {
			os.Setenv("DB_SCHEMA", "test_schema")
			_, err := NewConfig()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("JWT_SECRET environment variable is not set"))
		})

		It("should return error when DB_SCHEMA is missing", func() {
			os.Setenv("JWT_SECRET", "test-secret")
			_, err := NewConfig()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("DB_SCHEMA environment variable is not set"))
		})
	})

	Context("when token duration values are invalid", func() {
		BeforeEach(func() {
			os.Setenv("JWT_SECRET", "test-secret")
			os.Setenv("DB_SCHEMA", "test_schema")
		})

		It("should return error for invalid ACCESS_TOKEN_HOURS", func() {
			os.Setenv("ACCESS_TOKEN_HOURS", "invalid")
			_, err := NewConfig()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid ACCESS_TOKEN_HOURS"))
		})

		It("should return error for invalid REFRESH_TOKEN_DAYS", func() {
			os.Setenv("REFRESH_TOKEN_DAYS", "invalid")
			_, err := NewConfig()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid REFRESH_TOKEN_DAYS"))
		})

		It("should return error for zero ACCESS_TOKEN_HOURS", func() {
			os.Setenv("ACCESS_TOKEN_HOURS", "0")
			_, err := NewConfig()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("ACCESS_TOKEN_HOURS must be greater than 0"))
		})

		It("should return error for zero REFRESH_TOKEN_DAYS", func() {
			os.Setenv("REFRESH_TOKEN_DAYS", "0")
			_, err := NewConfig()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("REFRESH_TOKEN_DAYS must be greater than 0"))
		})
	})

	Context("environment checks", func() {
		BeforeEach(func() {
			os.Setenv("JWT_SECRET", "test-secret")
			os.Setenv("DB_SCHEMA", "test_schema")
		})

		It("should correctly identify dev environment", func() {
			os.Setenv("ENV", "dev")
			cfg, err := NewConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.IsDev()).To(BeTrue())
			Expect(cfg.IsProd()).To(BeFalse())
		})

		It("should correctly identify prod environment", func() {
			os.Setenv("ENV", "prod")
			cfg, err := NewConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.IsDev()).To(BeFalse())
			Expect(cfg.IsProd()).To(BeTrue())
		})

		It("should correctly identify other environments", func() {
			os.Setenv("ENV", "staging")
			cfg, err := NewConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(cfg.IsDev()).To(BeFalse())
			Expect(cfg.IsProd()).To(BeFalse())
		})
	})
})
