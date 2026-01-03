package postgres_test

import (
	"context"
	"errors"
	"os"
	"time"

	"expenses/internal/config"
	"expenses/pkg/database/manager/base"
	"expenses/pkg/database/manager/postgres"

	"github.com/jackc/pgx/v5"
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
		if (os.Getenv("DB_PORT") == "") {
			os.Setenv("DB_PORT", "5432")
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

		It("should return transaction info when inside a transaction", func() {
			err := dbManager.WithTxn(ctx, func(txCtx context.Context) error {
				info, err := dbManager.GetTransactionInfo(txCtx)
				Expect(err).NotTo(HaveOccurred())
				Expect(info).NotTo(BeNil())
				Expect(info.ID).NotTo(BeEmpty())
				Expect(info.IsReadOnly).To(BeFalse())
				Expect(info.IsNested).To(BeFalse())
				return nil
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Batch Operations", func() {
		BeforeEach(func() {
			_, err := dbManager.ExecuteQuery(ctx, `CREATE TABLE IF NOT EXISTS test_batch (id SERIAL PRIMARY KEY, name TEXT)`)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			_, _ = dbManager.ExecuteQuery(ctx, `DROP TABLE IF EXISTS test_batch`)
		})

		It("should execute a batch of operations successfully using the pool", func() {
			batch := &pgx.Batch{}
			batch.Queue("INSERT INTO test_batch (name) VALUES ($1)", "batch1")
			batch.Queue("INSERT INTO test_batch (name) VALUES ($1)", "batch2")

			err := dbManager.ExecuteBatch(ctx, batch)
			Expect(err).NotTo(HaveOccurred())

			var count int
			row := dbManager.FetchOne(ctx, `SELECT COUNT(*) FROM test_batch`)
			Expect(row.Scan(&count)).To(Succeed())
			Expect(count).To(Equal(2))
		})

		It("should execute a batch of operations successfully within a transaction", func() {
			err := dbManager.WithTxn(ctx, func(txCtx context.Context) error {
				batch := &pgx.Batch{}
				batch.Queue("INSERT INTO test_batch (name) VALUES ($1)", "txn_batch1")
				batch.Queue("INSERT INTO test_batch (name) VALUES ($1)", "txn_batch2")
				return dbManager.ExecuteBatch(txCtx, batch)
			})
			Expect(err).NotTo(HaveOccurred())

			var count int
			row := dbManager.FetchOne(ctx, `SELECT COUNT(*) FROM test_batch`)
			Expect(row.Scan(&count)).To(Succeed())
			Expect(count).To(Equal(2))
		})

		It("should return an error if a batch operation fails", func() {
			// Note: pgx.Batch will not fail on invalid queries during Queue, only on execution.
			// The entire transaction should be rolled back.
			err := dbManager.WithTxn(ctx, func(txCtx context.Context) error {
				batch := &pgx.Batch{}
				batch.Queue("INSERT INTO test_batch (name) VALUES ($1)", "good_batch")
				// This one will fail because of a wrong column name
				batch.Queue("INSERT INTO test_batch (invalid_column) VALUES ($1)", "bad_batch")
				return dbManager.ExecuteBatch(txCtx, batch)
			})

			Expect(err).To(HaveOccurred())
			// Check for a common postgres error message for undefined columns
			Expect(err.Error()).To(ContainSubstring("column \"invalid_column\" of relation \"test_batch\" does not exist"))

			// Verify that the transaction was rolled back and no data was inserted
			var count int
			row := dbManager.FetchOne(ctx, `SELECT COUNT(*) FROM test_batch`)
			Expect(row.Scan(&count)).To(Succeed())
			Expect(count).To(Equal(0))
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

		It("should successfully execute a retryable transaction", func() {
			err := dbManager.WithRetryableTxn(ctx, func(txCtx context.Context) error {
				_, err := dbManager.ExecuteQuery(txCtx, `INSERT INTO test_enhanced (name) VALUES ($1)`, "retryable_test")
				return err
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify data was committed
			var count int
			row := dbManager.FetchOne(ctx, `SELECT COUNT(*) FROM test_enhanced WHERE name = $1`, "retryable_test")
			err = row.Scan(&count)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(1))
		})

		It("should successfully commit a savepoint within a transaction", func() {
			err := dbManager.WithTxn(ctx, func(txCtx context.Context) error {
				// Insert initial data in the main transaction
				_, err := dbManager.ExecuteQuery(txCtx, `INSERT INTO test_enhanced (name) VALUES ($1)`, "before_savepoint")
				Expect(err).NotTo(HaveOccurred())

				// Create a savepoint
				err = dbManager.WithSavepoint(txCtx, "sp1", func(spCtx context.Context) error {
					_, err := dbManager.ExecuteQuery(spCtx, `INSERT INTO test_enhanced (name) VALUES ($1)`, "in_savepoint")
					return err
				})
				Expect(err).NotTo(HaveOccurred())
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify both records are there
			var count int
			row := dbManager.FetchOne(ctx, `SELECT COUNT(*) FROM test_enhanced`)
			err = row.Scan(&count)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(2))
		})

		It("should roll back a savepoint without affecting the parent transaction", func() {
			err := dbManager.WithTxn(ctx, func(txCtx context.Context) error {
				// Insert initial data
				_, err := dbManager.ExecuteQuery(txCtx, `INSERT INTO test_enhanced (name) VALUES ($1)`, "before_savepoint")
				Expect(err).NotTo(HaveOccurred())

				// Create a savepoint that will be rolled back
				err = dbManager.WithSavepoint(txCtx, "sp1", func(spCtx context.Context) error {
					_, err := dbManager.ExecuteQuery(spCtx, `INSERT INTO test_enhanced (name) VALUES ($1)`, "in_savepoint")
					Expect(err).NotTo(HaveOccurred())
					return errors.New("force rollback") // Force the savepoint to roll back
				})
				// This error is expected
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("force rollback"))

				// The parent transaction should continue
				return nil
			})
			Expect(err).NotTo(HaveOccurred())

			// Verify only the initial data is present
			var count int
			row := dbManager.FetchOne(ctx, `SELECT COUNT(*) FROM test_enhanced`)
			err = row.Scan(&count)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(1))

			var name string
			row = dbManager.FetchOne(ctx, `SELECT name FROM test_enhanced`)
			err = row.Scan(&name)
			Expect(err).NotTo(HaveOccurred())
			Expect(name).To(Equal("before_savepoint"))
		})

		It("should execute a function with a dedicated connection", func() {
			err := dbManager.WithConnection(ctx, func(conn *pgx.Conn) error {
				var result int
				// Use the connection to perform a simple query
				err := conn.QueryRow(ctx, "SELECT 1").Scan(&result)
				Expect(result).To(Equal(1))
				return err
			})
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

	Describe("Error Paths and Disabled Features", func() {
		var (
			basicDbManager base.DatabaseManager
		)

		BeforeAll(func() {
			// Create a manager with most features disabled
			factory := postgres.NewPostgreSQLFactory()
			var err error
			basicDbManager, err = factory.CreateDatabaseManager(cfg, base.BasicConfig())
			Expect(err).NotTo(HaveOccurred())
			Expect(basicDbManager).NotTo(BeNil())
		})

		AfterAll(func() {
			if basicDbManager != nil {
				basicDbManager.Close()
			}
		})

		It("should return an error when using WithTxnOptions if retry is disabled", func() {
			err := basicDbManager.WithTxnOptions(ctx, nil, func(txCtx context.Context) error {
				return nil
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("retry feature is disabled"))
		})

		It("should return an error when using WithReadOnlyTxn if retry is disabled", func() {
			err := basicDbManager.WithReadOnlyTxn(ctx, func(txCtx context.Context) error {
				return nil
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("retry feature is disabled"))
		})

		It("should return an error when using WithRetryableTxn if retry is disabled", func() {
			err := basicDbManager.WithRetryableTxn(ctx, func(txCtx context.Context) error {
				return nil
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("retry feature is disabled"))
		})

		It("should return an error when using WithSavepoint if savepoints are disabled", func() {
			// Need to be in a transaction to even attempt a savepoint
			err := basicDbManager.WithTxn(ctx, func(txCtx context.Context) error {
				return basicDbManager.WithSavepoint(txCtx, "sp1", func(spCtx context.Context) error {
					return nil
				})
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("savepoints feature is disabled"))
		})

		It("should return an error when using ExecuteBatch if batch is disabled", func() {
			batch := &pgx.Batch{}
			err := basicDbManager.ExecuteBatch(ctx, batch)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("batch feature is disabled"))
		})

		It("should return an error when using WithSavepoint outside of a transaction", func() {
			err := dbManager.WithSavepoint(ctx, "sp1", func(spCtx context.Context) error {
				return nil
			})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("savepoints can only be used within an existing transaction"))
		})

		It("should return an error when getting transaction info outside of a transaction", func() {
			_, err := dbManager.GetTransactionInfo(ctx)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("not in a transaction"))
		})
	})
})
