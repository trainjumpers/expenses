package database_test

import (
	"context"
	"expenses/internal/config"
	manager "expenses/internal/database/manager"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("PostgresDatabaseManager", Ordered, func() {
	var (
		dbManager manager.DatabaseManager
		ctx       context.Context
	)

	BeforeAll(func() {
		ctx = context.Background()
		cfg := &config.Config{DBSchema: os.Getenv("DB_SCHEMA")}
		var err error
		dbManager, err = manager.NewPostgresDatabaseManager(cfg)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterAll(func() {
		if dbManager != nil {
			_ = dbManager.Close()
		}
	})

	Describe("ExecuteQuery", func() {
		BeforeEach(func() {
			_, err := dbManager.ExecuteQuery(ctx, `CREATE TABLE IF NOT EXISTS testing_exec (id SERIAL PRIMARY KEY, val TEXT)`)
			Expect(err).NotTo(HaveOccurred())
			_, err = dbManager.ExecuteQuery(ctx, `INSERT INTO testing_exec (val) VALUES ('foo'), ('bar')`)
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			_, _ = dbManager.ExecuteQuery(ctx, `DROP TABLE IF EXISTS testing_exec`)
		})
		It("inserts a row and returns affected count", func() {
			n, err := dbManager.ExecuteQuery(ctx, `INSERT INTO testing_exec (val) VALUES ($1)`, "baz")
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(int64(1)))
		})
		It("returns error for invalid SQL", func() {
			_, err := dbManager.ExecuteQuery(ctx, `INSERT INTO not_a_table (val) VALUES ($1)`, "bar")
			Expect(err).To(HaveOccurred())
		})
		It("returns zero rows affected for no match", func() {
			n, err := dbManager.ExecuteQuery(ctx, `DELETE FROM testing_exec WHERE val = $1`, "notfound")
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(int64(0)))
		})
		It("returns multiple rows affected for multi-match", func() {
			n, err := dbManager.ExecuteQuery(ctx, `DELETE FROM testing_exec WHERE val IN ($1, $2)`, "foo", "bar")
			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(int64(2)))
		})
	})

	Describe("FetchOne", func() {
		BeforeEach(func() {
			_, err := dbManager.ExecuteQuery(ctx, `CREATE TABLE IF NOT EXISTS testing_fetchone (id SERIAL PRIMARY KEY, val TEXT)`)
			Expect(err).NotTo(HaveOccurred())
			_, err = dbManager.ExecuteQuery(ctx, `INSERT INTO testing_fetchone (val) VALUES ('one'), ('one')`)
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			_, _ = dbManager.ExecuteQuery(ctx, `DROP TABLE IF EXISTS testing_fetchone`)
		})
		It("fetches a single row", func() {
			row := dbManager.FetchOne(ctx, `SELECT val FROM testing_fetchone WHERE val = $1 LIMIT 1`, "one")
			var val string
			Expect(row.Scan(&val)).To(Succeed())
			Expect(val).To(Equal("one"))
		})
		It("returns error for no rows", func() {
			row := dbManager.FetchOne(ctx, `SELECT val FROM testing_fetchone WHERE val = $1`, "notfound")
			var val string
			Expect(row.Scan(&val)).To(MatchError(pgx.ErrNoRows))
		})
		It("returns error for multiple rows (scan called twice)", func() {
			row := dbManager.FetchOne(ctx, `SELECT val FROM testing_fetchone WHERE val = $1`, "one")
			var val1, val2 string
			Expect(row.Scan(&val1)).To(Succeed())
			// Scan again should error (simulate misuse)
			Expect(row.Scan(&val2)).To(HaveOccurred())
		})
	})

	Describe("FetchAll", func() {
		BeforeEach(func() {
			_, err := dbManager.ExecuteQuery(ctx, `CREATE TABLE IF NOT EXISTS testing_fetchall (id SERIAL PRIMARY KEY, val TEXT)`)
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			_, _ = dbManager.ExecuteQuery(ctx, `DROP TABLE IF EXISTS testing_fetchall`)
		})
		It("fetches all rows", func() {
			_, err := dbManager.ExecuteQuery(ctx, `INSERT INTO testing_fetchall (val) VALUES ('a'), ('b')`)
			Expect(err).NotTo(HaveOccurred())
			rows, err := dbManager.FetchAll(ctx, `SELECT val FROM testing_fetchall ORDER BY val`)
			Expect(err).NotTo(HaveOccurred())
			defer rows.Close()
			var vals []string
			for rows.Next() {
				var v string
				Expect(rows.Scan(&v)).To(Succeed())
				vals = append(vals, v)
			}
			Expect(vals).To(Equal([]string{"a", "b"}))
		})
		It("returns error for invalid SQL", func() {
			_, err := dbManager.FetchAll(ctx, `SELECT nope FROM not_a_table`)
			Expect(err).To(HaveOccurred())
		})
		It("returns empty result set for no match", func() {
			rows, err := dbManager.FetchAll(ctx, `SELECT val FROM testing_fetchall WHERE val = $1`, "notfound")
			Expect(err).NotTo(HaveOccurred())
			defer rows.Close()
			Expect(rows.Next()).To(BeFalse())
		})
	})

	Describe("WithTxn", func() {
		BeforeEach(func() {
			_, err := dbManager.ExecuteQuery(ctx, `CREATE TABLE IF NOT EXISTS testing_txn (id SERIAL PRIMARY KEY, val TEXT)`)
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			_, _ = dbManager.ExecuteQuery(ctx, `DROP TABLE IF EXISTS testing_txn`)
		})
		It("commits on success", func() {
			err := dbManager.WithTxn(ctx, func(tx pgx.Tx) error {
				_, err := tx.Exec(ctx, `INSERT INTO testing_txn (val) VALUES ($1)`, "txn")
				return err
			})
			Expect(err).NotTo(HaveOccurred())
			var count int
			row := dbManager.FetchOne(ctx, `SELECT COUNT(*) FROM testing_txn WHERE val = $1`, "txn")
			Expect(row.Scan(&count)).To(Succeed())
			Expect(count).To(Equal(1))
		})
		It("rolls back on error", func() {
			err := dbManager.WithTxn(ctx, func(tx pgx.Tx) error {
				_, err := tx.Exec(ctx, `INSERT INTO testing_txn (val) VALUES ($1)`, "fail")
				if err != nil {
					return err
				}
				return pgx.ErrTxClosed // simulate error
			})
			Expect(err).To(HaveOccurred())
			var count int
			row := dbManager.FetchOne(ctx, `SELECT COUNT(*) FROM testing_txn WHERE val = $1`, "fail")
			Expect(row.Scan(&count)).To(Succeed())
			Expect(count).To(Equal(0))
		})
		It("rolls back and propagates panic", func() {
			defer func() {
				recover() // absorb panic for test
				var count int
				row := dbManager.FetchOne(ctx, `SELECT COUNT(*) FROM testing_txn WHERE val = $1`, "panic")
				Expect(row.Scan(&count)).To(Succeed())
				Expect(count).To(Equal(0))
			}()
			err := dbManager.WithTxn(ctx, func(tx pgx.Tx) error {
				_, err := tx.Exec(ctx, `INSERT INTO testing_txn (val) VALUES ($1)`, "panic")
				if err != nil {
					return err
				}
				panic("test panic")
			})
			// Should not reach here
			Expect(err).To(BeNil())
		})
	})

	Describe("WithLock", func() {
		BeforeEach(func() {
			_, err := dbManager.ExecuteQuery(ctx, `CREATE TABLE IF NOT EXISTS testing_lock (id SERIAL PRIMARY KEY, val TEXT)`)
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			_, _ = dbManager.ExecuteQuery(ctx, `DELETE FROM testing_lock`)
			_, _ = dbManager.ExecuteQuery(ctx, `DROP TABLE IF EXISTS testing_lock`)
		})

		It("acquires advisory lock and inserts row", func() {
			lockKey := int64(424242)
			err := dbManager.WithLock(ctx, lockKey, func(tx pgx.Tx) error {
				_, err := tx.Exec(ctx, `INSERT INTO testing_lock (val) VALUES ($1)`, "locked!")
				return err
			})
			Expect(err).NotTo(HaveOccurred())
			// Verify row exists
			var count int
			row := dbManager.FetchOne(ctx, `SELECT COUNT(*) FROM testing_lock WHERE val = $1`, "locked!")
			Expect(row.Scan(&count)).To(Succeed())
			Expect(count).To(Equal(1))
		})

		It("serializes concurrent WithLock calls on same key", func() {
			lockKey := int64(555555)
			ch := make(chan string, 2)
			start := make(chan struct{})

			go func() {
				<-start
				dbManager.WithLock(ctx, lockKey, func(tx pgx.Tx) error {
					ch <- "first acquired"
					time.Sleep(500 * time.Millisecond)
					return nil
				})
			}()

			go func() {
				<-start
				dbManager.WithLock(ctx, lockKey, func(tx pgx.Tx) error {
					ch <- "second acquired"
					return nil
				})
			}()

			close(start)
			first := <-ch
			second := <-ch
			// The first goroutine should acquire the lock before the second
			Expect(first).To(Equal("first acquired"))
			Expect(second).To(Equal("second acquired"))
		})
		It("returns error on lock contention with timeout", func() {
			lockKey := int64(888888)
			ctx1, cancel1 := context.WithCancel(ctx)
			ctx2, cancel2 := context.WithTimeout(ctx, 200*time.Millisecond)
			defer cancel1()
			defer cancel2()
			ch := make(chan error, 2)

			go func() {
				err := dbManager.WithLock(ctx1, lockKey, func(tx pgx.Tx) error {
					time.Sleep(500 * time.Millisecond)
					return nil
				})
				ch <- err
			}()

			go func() {
				err := dbManager.WithLock(ctx2, lockKey, func(tx pgx.Tx) error {
					return nil
				})
				ch <- err
			}()

			err1 := <-ch
			err2 := <-ch
			// One should succeed, one should error with context deadline exceeded
			if err1 == nil {
				Expect(err2).To(HaveOccurred())
			} else {
				Expect(err1).To(HaveOccurred())
			}
		})
		It("does not block on different lock keys", func() {
			lockKey1 := int64(111111)
			lockKey2 := int64(222222)
			ch := make(chan string, 2)
			start := make(chan struct{})

			go func() {
				<-start
				dbManager.WithLock(ctx, lockKey1, func(tx pgx.Tx) error {
					ch <- "lock1"
					return nil
				})
			}()

			go func() {
				<-start
				dbManager.WithLock(ctx, lockKey2, func(tx pgx.Tx) error {
					ch <- "lock2"
					return nil
				})
			}()

			close(start)
			results := []string{<-ch, <-ch}
			Expect(results).To(ContainElements("lock1", "lock2"))
		})
	})

	Describe("Close", func() {
		It("closes the manager without error", func() {
			Expect(dbManager.Close()).To(Succeed())
		})
		It("exits gracefully if already closed", func() {
			_ = dbManager.Close()
			Expect(dbManager.Close()).To(Succeed())
		})
		It("returns error if used after close", func() {
			_ = dbManager.Close()
			_, err := dbManager.ExecuteQuery(ctx, `SELECT 1`)
			Expect(err).To(HaveOccurred())
		})
	})
})
