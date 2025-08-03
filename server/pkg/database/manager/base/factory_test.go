package base_test

import (
	"os"

	"expenses/pkg/database/manager/base"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Database Factory", func() {

	Describe("GetDatabaseType", func() {
		var originalDBType string

		BeforeEach(func() {
			originalDBType = os.Getenv("DB_TYPE")
		})

		AfterEach(func() {
			os.Setenv("DB_TYPE", originalDBType)
		})

		Context("when DB_TYPE environment variable is set", func() {
			It("should return the specified database type", func() {
				os.Setenv("DB_TYPE", "postgres")
				dbType := base.GetDatabaseType()
				Expect(dbType).To(Equal(base.PostgreSQL))
			})
		})

		Context("when DB_TYPE environment variable is not set", func() {
			It("should return PostgreSQL as the default type", func() {
				os.Unsetenv("DB_TYPE")
				dbType := base.GetDatabaseType()
				Expect(dbType).To(Equal(base.PostgreSQL))
			})
		})
	})

	Describe("ValidateDatabaseType", func() {
		Context("with a supported database type", func() {
			It("should not return an error for PostgreSQL", func() {
				err := base.ValidateDatabaseType(base.PostgreSQL)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with an unsupported database type", func() {
			It("should return an error", func() {
				err := base.ValidateDatabaseType("unsupported_db")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unsupported database type: unsupported_db"))
			})
		})
	})
})
