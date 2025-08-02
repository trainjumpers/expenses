package postgres_test

import (
	"context"
	"errors"
	"time"

	"expenses/internal/config"
	"expenses/pkg/database/manager/base"
	"expenses/pkg/database/manager/postgres"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostgreSQL Database Manager", Ordered, func() {
	var (
		dbManager base.DatabaseManager
		ctx       context.Context
		cfg       *config.Config
	)

	BeforeAll(func() {
		cfg = &config.Config{
			DBSchema: "public",
		}
		
		var err error
		factory := postgres.NewPostgreSQLFactory()
		dbManager, err = factory.CreateDatabaseManager(cfg, base.DefaultConfig())
		Expect(err).NotTo(HaveOccurred())
		Expect(dbManager).NotTo(BeNil())
		
		ctx = context.Background()
	})

	AfterAll(func() {
		if dbManager != nil {
			dbManager.Close()
		}
	})

	Describe("Basic Operations", func() {
		BeforeEach(func() {
			_, err := dbManager.ExecuteQuery(ctx, `CREATE TABLE IF NOT EXISTS test_basic (id SERIAL PRIMARY KEY, name TEXT)`)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_, _ = dbManager.ExecuteQuery(ctx, `DROP TABLE IF EXISTS test_basic`)
		})

		It("should execute queries successfully", func() {
			rowsAffected, err := dbManager.ExecuteQuery(ctx, `INSERT INTO test_basic (name) VALUES ($1)`, "test")
			Expect(err).NotTo(HaveOccurred())
			Expect(rowsAffected).To(Equal(int64(1)))
		})

		It("should fetch single rows", func() {
			_, err := dbManager.ExecuteQuery(ctx, `INSERT INTO test_basic (name) VALUES ($1)`, "test")
			Expect(err).NotTo(HaveOccurred())

			row := dbManager.FetchOne(ctx, `SELECT name FROM test_basic WHERE name = $1`, "test")
			var name string
			err = row.Scan(&name)
			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal("test"))
		})

		It("should fetch multiple rows", func() {
			_, err := dbManager.ExecuteQuery(ctx, `INSERT INTO test_basic (name) VALUES ($1), ($2)`, "test1", "test2")
			Expect(err).NotTo(HaveOccurred())

			rows, err := dbManager.FetchAll(ctx, `SELECT name FROM test_basic ORDER BY name`)
			Expect(err).NotTo(HaveOccurred())
			defer rows.Close()

			var names []string
			for rows.Next() {
				var name string
				err := rows.Scan(&name)
				Expect(err).NotTo(HaveOccurred())
				names = append(names, name)
			}
			Expect(names).To(Equal([]string{"test1", "test2"}))
		})
	})

	Describe("Transaction Operations", func() {
		BeforeEach(func() {
			_, err := dbManager.ExecuteQuery(ctx, `CREATE TABLE IF NOT EXISTS test_txn (id SERIAL PRIMARY KEY, name TEXT)`)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_, _ = dbManager.ExecuteQuery(ctx, `DROP TABLE IF EXISTS test_txn`)
		})

		It("should commit transactions successfully", func() {
			err := dbManager.WithTxn(ctx, func(txCtx context.Context) error {
				_, err := dbManager.ExecuteQuery(txCtx, `INSERT INTO test_txn (name) VALUES ($1)`, "txn_test")
				return err
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify the data was committed
			row := dbManager.FetchOne(ctx, `SELECT COUNT(*) FROM test_txn WHERE name = $1`, "txn_test")
			var count int
			err = row.Scan(&count)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(1))
		})

		It("should rollback transactions on error", func() {
			err := dbManager.WithTxn(ctx, func(txCtx context.Context) error {
				_, err := dbManager.ExecuteQuery(txCtx, `INSERT INTO test_txn (name) VALUES ($1)`, "rollback_test")
				if err != nil {
					return err
				}
				return errors.New("forced error")
			})
			Expect(err).To(HaveOccurred())

			// Verify the data was rolled back
			row := dbManager.FetchOne(ctx, `SELECT COUNT(*) FROM test_txn WHERE name = $1`, "rollback_test")
			var count int
			err = row.Scan(&count)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(0))
		})

		It("should use transaction context for operations within transaction", func() {
			err := dbManager.WithTxn(ctx, func(txCtx context.Context) error {
				// Insert data
				_, err := dbManager.ExecuteQuery(txCtx, `INSERT INTO test_txn (name) VALUES ($1)`, "context_test")
				if err != nil {
					return err
				}

				// This should see the data within the same transaction
				row := dbManager.FetchOne(txCtx, `SELECT name FROM test_txn WHERE name = $1`, "context_test")
				var name string
				err = row.Scan(&name)
				if err != nil {
					return err
				}
				Expect(name).To(Equal("context_test"))

				return nil
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Lock Operations", func() {
		BeforeEach(func() {
			_, err := dbManager.ExecuteQuery(ctx, `CREATE TABLE IF NOT EXISTS test_lock (id SERIAL PRIMARY KEY, name TEXT)`)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_, _ = dbManager.ExecuteQuery(ctx, `DROP TABLE IF EXISTS test_lock`)
		})

		It("should acquire advisory locks", func() {
			lockKey := int64(12345)
			err := dbManager.WithLock(ctx, lockKey, func(txCtx context.Context) error {
				_, err := dbManager.ExecuteQuery(txCtx, `INSERT INTO test_lock (name) VALUES ($1)`, "locked")
				return err
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify the data was inserted
			row := dbManager.FetchOne(ctx, `SELECT COUNT(*) FROM test_lock WHERE name = $1`, "locked")
			var count int
			err = row.Scan(&count)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(1))
		})
	})

	Describe("Enhanced Features", func() {
		BeforeEach(func() {
			_, err := dbManager.ExecuteQuery(ctx, `CREATE TABLE IF NOT EXISTS test_enhanced (id SERIAL PRIMARY KEY, name TEXT)`)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_, _ = dbManager.ExecuteQuery(ctx, `DROP TABLE IF EXISTS test_enhanced`)
		})

		It("should support read-only transactions", func() {
			// First insert some data
			_, err := dbManager.ExecuteQuery(ctx, `INSERT INTO test_enhanced (name) VALUES ($1)`, "readonly_test")
			Expect(err).NotTo(HaveOccurred())

			// Now test read-only transaction
			err = dbManager.WithReadOnlyTxn(ctx, func(txCtx context.Context) error {
				row := dbManager.FetchOne(txCtx, `SELECT name FROM test_enhanced WHERE name = $1`, "readonly_test")
				var name string
				err := row.Scan(&name)
				Expect(err).NotTo(HaveOccurred())
				Expect(name).To(Equal("readonly_test"))
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should support transaction options", func() {
			opts := &base.TransactionOptions{
				Timeout:  5 * time.Second,
				ReadOnly: false,
			}

			err := dbManager.WithTxnOptions(ctx, opts, func(txCtx context.Context) error {
				_, err := dbManager.ExecuteQuery(txCtx, `INSERT INTO test_enhanced (name) VALUES ($1)`, "options_test")
				return err
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should provide database statistics", func() {
			stats := dbManager.Stats()
			Expect(stats.MaxConnections).To(BeNumerically(">", 0))
		})

		It("should support ping", func() {
			err := dbManager.Ping(ctx)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Monitoring Features", func() {
		BeforeEach(func() {
			_, err := dbManager.ExecuteQuery(ctx, `CREATE TABLE IF NOT EXISTS test_monitored (id SERIAL PRIMARY KEY, name TEXT)`)
			Expect(err).NotTo(HaveOccurred())
			
			// Reset metrics before each test
			dbManager.ResetMetrics()
		})

		AfterEach(func() {
			_, _ = dbManager.ExecuteQuery(ctx, `DROP TABLE IF EXISTS test_monitored`)
		})

		It("should track transaction metrics", func() {
			// Execute a successful transaction
			err := dbManager.WithTxn(ctx, func(txCtx context.Context) error {
				_, err := dbManager.ExecuteQuery(txCtx, `INSERT INTO test_monitored (name) VALUES ($1)`, "monitored_test")
				return err
			})
			Expect(err).NotTo(HaveOccurred())

			// Check metrics
			metrics := dbManager.GetMonitoringMetrics()
			Expect(metrics.TotalTransactions).To(Equal(int64(1)))
			Expect(metrics.CommittedTransactions).To(Equal(int64(1)))
			Expect(metrics.FailedTransactions).To(Equal(int64(0)))
			Expect(metrics.ActiveTransactions).To(Equal(int64(0)))
		})

		It("should track failed transactions", func() {
			// Execute a failed transaction
			err := dbManager.WithTxn(ctx, func(txCtx context.Context) error {
				return errors.New("test error")
			})
			Expect(err).To(HaveOccurred())

			// Check metrics
			metrics := dbManager.GetMonitoringMetrics()
			Expect(metrics.TotalTransactions).To(Equal(int64(1)))
			Expect(metrics.CommittedTransactions).To(Equal(int64(0)))
			Expect(metrics.FailedTransactions).To(Equal(int64(1)))
			Expect(len(metrics.ErrorsByType)).To(BeNumerically(">", 0))
		})
	})

	Describe("Feature Detection", func() {
		It("should detect enabled features", func() {
			// With default config, all features should be enabled
			Expect(dbManager.IsFeatureEnabled(base.FeatureRetry)).To(BeTrue())
			Expect(dbManager.IsFeatureEnabled(base.FeatureSavepoints)).To(BeTrue())
			Expect(dbManager.IsFeatureEnabled(base.FeatureBatch)).To(BeTrue())
			Expect(dbManager.IsFeatureEnabled(base.FeatureMonitoring)).To(BeTrue())
		})

		It("should return current configuration", func() {
			config := dbManager.GetConfig()
			Expect(config).NotTo(BeNil())
			Expect(config.EnableTransactions).To(BeTrue())
			Expect(config.EnableLocks).To(BeTrue())
		})
	})
})
