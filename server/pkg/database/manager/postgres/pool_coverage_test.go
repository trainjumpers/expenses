package postgres

import (
	"os"

	"expenses/internal/config"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("createConnectionPool error paths", func() {
	setEnv := func(key, value string) {
		original, existed := os.LookupEnv(key)
		DeferCleanup(func() {
			if existed {
				_ = os.Setenv(key, original)
			} else {
				_ = os.Unsetenv(key)
			}
		})
		_ = os.Setenv(key, value)
	}

	It("defaults sslmode, uses the development pool config and reports a failed ping", func() {
		setEnv("DB_HOST", "127.0.0.1")
		setEnv("DB_PORT", "1")
		setEnv("DB_USER", "user")
		setEnv("DB_PASSWORD", "password")
		setEnv("DB_NAME", "db")
		setEnv("DB_SSL_MODE", "")
		setEnv("GIN_MODE", "debug")

		pool, err := createConnectionPool(&config.Config{DBSchema: "public"})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to ping database"))
		Expect(pool).To(BeNil())
	})

	It("returns an error when the connection configuration cannot be parsed", func() {
		setEnv("DB_HOST", "127.0.0.1")
		setEnv("DB_PORT", "5432")
		setEnv("DB_USER", "user")
		setEnv("DB_PASSWORD", "password")
		setEnv("DB_NAME", "db")
		setEnv("DB_SSL_MODE", "not-a-real-sslmode")
		setEnv("GIN_MODE", "release")

		pool, err := createConnectionPool(&config.Config{DBSchema: "public"})

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("failed to connect to database"))
		Expect(pool).To(BeNil())
	})
})
