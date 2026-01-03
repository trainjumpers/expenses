package postgres_test

import (
	"context"
	"os"

	"expenses/internal/config"
	"expenses/pkg/database/manager/base"
	"expenses/pkg/database/manager/postgres"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostgreSQLFactory", func() {
	var (
		factory       *postgres.PostgreSQLFactory
		cfg           *config.Config
		managerConfig *base.DatabaseManagerConfig
	)

	BeforeEach(func() {
		factory = postgres.NewPostgreSQLFactory()
		cfg = &config.Config{DBSchema: "public"}
		managerConfig = base.DefaultConfig()
	})

	Describe("CreateDatabaseManager", func() {
		// Note: This test requires a running PostgreSQL database instance to pass.
		// It is skipped by default unless running in a CI environment.
		It("should create a unified PostgreSQL database manager successfully", func() {
			if os.Getenv("CI") == "" {
				Skip("Skipping database integration test. Set CI=true to run.")
			}

			dbManager, err := factory.CreateDatabaseManager(cfg, managerConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(dbManager).NotTo(BeNil())

			// Ensure the created manager is usable
			Expect(dbManager.Ping(context.TODO())).To(Succeed())

			dbManager.Close()
		})

		It("should return an error if connection pool creation fails due to invalid configuration", func() {
			// Preserve original DB_PORT and restore after test to avoid leaking env changes
			origPort := os.Getenv("DB_PORT")
			defer func() {
				if origPort == "" {
					_ = os.Unsetenv("DB_PORT")
				} else {
					_ = os.Setenv("DB_PORT", origPort)
				}
			}()

			// Set an invalid port to force strconv.Atoi to fail within createConnectionPool
			os.Setenv("DB_PORT", "this-is-not-a-port-number")

			dbManager, err := factory.CreateDatabaseManager(cfg, managerConfig)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid database port number"))
			Expect(dbManager).To(BeNil())
		})
	})
})
