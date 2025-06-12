package database_test

import (
	"context"
	"errors"
	"expenses/internal/config"
	database "expenses/internal/database/manager"
	mock_database "expenses/internal/mock/database"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDatabaseManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DatabaseManager Suite")
}

var _ = Describe("DatabaseManager interface contract", func() {
	var (
		db  database.DatabaseManager
		ctx context.Context
	)

	BeforeEach(func() {
		db = mock_database.NewMockDatabaseManager()
		ctx = context.Background()
	})

	It("ExecuteQuery returns 1 row affected and no error by default", func() {
		n, err := db.ExecuteQuery(ctx, "UPDATE table SET x=1")
		Expect(err).NotTo(HaveOccurred())
		Expect(n).To(Equal(int64(1)))
	})

	It("FetchOne returns a mock row that can be scanned", func() {
		row := db.FetchOne(ctx, "SELECT 1")
		var x int
		err := row.Scan(&x)
		Expect(err).NotTo(HaveOccurred())
	})

	It("FetchAll returns mock rows and no error by default", func() {
		rows, err := db.FetchAll(ctx, "SELECT * FROM table")
		Expect(err).NotTo(HaveOccurred())
		Expect(rows).NotTo(BeNil())
		rows.Close()
	})

	It("WithTxn executes the function and returns its error", func() {
		called := false
		err := db.WithTxn(ctx, func(tx pgx.Tx) error {
			called = true
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(called).To(BeTrue())
	})

	It("WithLock executes the function and returns its error", func() {
		called := false
		err := db.WithLock(ctx, "LOCK TABLE foo", func(tx pgx.Tx) error {
			called = true
			return nil
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(called).To(BeTrue())
	})

	It("Close does not error", func() {
		Expect(db.Close()).To(Succeed())
	})

	It("ExecuteQuery returns error if set in mock", func() {
		mock := db.(*mock_database.MockDatabaseManager)
		mock.ExecuteQueryError = errors.New("fail")
		_, err := db.ExecuteQuery(ctx, "bad query")
		Expect(err).To(MatchError("fail"))
	})

	It("FetchAll returns error if set in mock", func() {
		mock := db.(*mock_database.MockDatabaseManager)
		mock.FetchAllError = errors.New("fail fetch")
		_, err := db.FetchAll(ctx, "bad query")
		Expect(err).To(MatchError("fail fetch"))
	})

	It("WithTxn returns error if ShouldFailWithTxn is set", func() {
		mock := db.(*mock_database.MockDatabaseManager)
		mock.ShouldFailWithTxn = true
		err := db.WithTxn(ctx, func(tx pgx.Tx) error { return nil })
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("NewDatabaseManager", func() {
	var origDBType string
	BeforeEach(func() {
		origDBType = os.Getenv("DB_TYPE")
	})
	AfterEach(func() {
		_ = os.Setenv("DB_TYPE", origDBType)
	})

	It("returns error for unsupported DB_TYPE", func() {
		_ = os.Setenv("DB_TYPE", "notarealdb")
		cfg := &config.Config{DBSchema: "public"}
		db, err := database.NewDatabaseManager(cfg)
		Expect(err).To(HaveOccurred())
		Expect(db).To(BeNil())
	})
})
