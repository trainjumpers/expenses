package repository

import (
	"context"
	"fmt"
	"time"

	"expenses/internal/config"
	"expenses/internal/models"
	databasemanager "expenses/pkg/database/manager"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StatementRepository", func() {
	var (
		ctx       context.Context
		cfg       *config.Config
		db        databasemanager.DatabaseManager
		repo      StatementRepositoryInterface
		userID    int64
		accountID int64
	)

	BeforeEach(func() {
		var err error
		cfg, err = config.NewConfig()
		Expect(err).NotTo(HaveOccurred())

		db, err = databasemanager.NewDatabaseManager(cfg)
		Expect(err).NotTo(HaveOccurred())

		repo = NewStatementRepository(db, cfg)
		ctx = context.Background()
		userID = createRepoTestUser(ctx, db, cfg.DBSchema)
		accountID = createRepoTestAccount(ctx, db, cfg.DBSchema, userID)
	})

	AfterEach(func() {
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf(`DELETE FROM %s.statement_transaction_mapping WHERE statement_id IN (SELECT id FROM %s.statement WHERE created_by = $1)`, cfg.DBSchema, cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.statement WHERE created_by = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.transaction WHERE created_by = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.account WHERE created_by = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.user WHERE id = $1", cfg.DBSchema), userID)
		Expect(db.Close()).To(Succeed())
	})

	createStatement := func(filename string) models.StatementResponse {
		statement, err := repo.CreateStatement(ctx, models.CreateStatementInput{
			AccountId:        accountID,
			CreatedBy:        userID,
			OriginalFilename: filename,
			FileType:         "csv",
			Status:           models.StatementStatusPending,
		})
		Expect(err).NotTo(HaveOccurred())
		return statement
	}

	createTransaction := func(name string) int64 {
		var id int64
		Expect(db.FetchOne(ctx,
			fmt.Sprintf("INSERT INTO %s.transaction (name, amount, date, account_id, created_by) VALUES ($1, $2, $3, $4, $5) RETURNING id", cfg.DBSchema),
			name, 10.0, time.Now().UTC(), accountID, userID,
		).Scan(&id)).To(Succeed())
		return id
	}

	Describe("CreateStatement", func() {
		It("stores a statement", func() {
			statement := createStatement("hdfc-april.csv")

			Expect(statement.Id).NotTo(BeZero())
			Expect(statement.OriginalFilename).To(Equal("hdfc-april.csv"))
			Expect(statement.Status).To(Equal(models.StatementStatusPending))
			Expect(statement.AccountId).To(Equal(accountID))
		})

		It("returns an account error for an unknown account", func() {
			_, err := repo.CreateStatement(ctx, models.CreateStatementInput{
				AccountId:        999999,
				CreatedBy:        userID,
				OriginalFilename: "orphan.csv",
				FileType:         "csv",
				Status:           models.StatementStatusPending,
			})
			expectAuthErrorType(err, "AccountNotFound")
		})
	})

	Describe("CreateStatementTxn and CreateStatementTxns", func() {
		It("links a single transaction", func() {
			statement := createStatement("single.csv")
			transactionID := createTransaction("Single txn")

			Expect(repo.CreateStatementTxn(ctx, statement.Id, transactionID)).To(Succeed())

			var count int
			Expect(db.FetchOne(ctx,
				fmt.Sprintf("SELECT COUNT(*) FROM %s.statement_transaction_mapping WHERE statement_id = $1 AND transaction_id = $2", cfg.DBSchema),
				statement.Id, transactionID,
			).Scan(&count)).To(Succeed())
			Expect(count).To(Equal(1))
		})

		It("returns an error when linking a duplicate mapping", func() {
			statement := createStatement("duplicate-link.csv")
			transactionID := createTransaction("Duplicate link")
			Expect(repo.CreateStatementTxn(ctx, statement.Id, transactionID)).To(Succeed())

			err := repo.CreateStatementTxn(ctx, statement.Id, transactionID)
			expectAuthErrorType(err, "StatementCreateError")
		})

		It("links a batch of transactions", func() {
			statement := createStatement("batch.csv")
			first := createTransaction("Batch first")
			second := createTransaction("Batch second")

			Expect(repo.CreateStatementTxns(ctx, statement.Id, []int64{first, second})).To(Succeed())

			var count int
			Expect(db.FetchOne(ctx,
				fmt.Sprintf("SELECT COUNT(*) FROM %s.statement_transaction_mapping WHERE statement_id = $1", cfg.DBSchema),
				statement.Id,
			).Scan(&count)).To(Succeed())
			Expect(count).To(Equal(2))
		})

		It("does nothing for an empty batch", func() {
			statement := createStatement("empty-batch.csv")
			Expect(repo.CreateStatementTxns(ctx, statement.Id, nil)).To(Succeed())
		})

		It("returns an error for an unknown transaction", func() {
			statement := createStatement("bad-batch.csv")
			err := repo.CreateStatementTxns(ctx, statement.Id, []int64{999999})
			expectAuthErrorType(err, "StatementCreateError")
		})
	})

	Describe("UpdateStatementStatus", func() {
		It("updates the status and message", func() {
			statement := createStatement("status.csv")
			message := "processed 12 transactions"

			updated, err := repo.UpdateStatementStatus(ctx, statement.Id, models.UpdateStatementStatusInput{
				Status:  models.StatementStatusDone,
				Message: &message,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(updated.Status).To(Equal(models.StatementStatusDone))
			Expect(updated.Message).NotTo(BeNil())
			Expect(*updated.Message).To(Equal(message))
		})

		It("returns not found for an unknown statement", func() {
			_, err := repo.UpdateStatementStatus(ctx, 999999, models.UpdateStatementStatusInput{Status: models.StatementStatusDone})
			expectAuthErrorType(err, "StatementNotFound")
		})

		It("returns an update error for an empty update", func() {
			statement := createStatement("empty-status.csv")
			_, err := repo.UpdateStatementStatus(ctx, statement.Id, models.UpdateStatementStatusInput{})
			expectAuthErrorType(err, "StatementUpdateError")
		})
	})

	Describe("GetStatementByID", func() {
		It("finds a statement for its owner", func() {
			statement := createStatement("find-me.csv")

			fetched, err := repo.GetStatementByID(ctx, statement.Id, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetched.Id).To(Equal(statement.Id))
			Expect(fetched.OriginalFilename).To(Equal("find-me.csv"))
		})

		It("returns not found for an unknown statement", func() {
			_, err := repo.GetStatementByID(ctx, 999999, userID)
			expectAuthErrorType(err, "StatementNotFound")
		})

		It("hides statements from another user", func() {
			statement := createStatement("private.csv")
			otherUserID := createRepoTestUser(ctx, db, cfg.DBSchema)
			defer func() {
				_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.user WHERE id = $1", cfg.DBSchema), otherUserID)
			}()

			_, err := repo.GetStatementByID(ctx, statement.Id, otherUserID)
			expectAuthErrorType(err, "StatementNotFound")
		})
	})

	Describe("ListStatementByUserId and CountStatementsByUserId", func() {
		var (
			csvStatement   models.StatementResponse
			excelStatement models.StatementResponse
		)

		BeforeEach(func() {
			csvStatement = createStatement("account-one.csv")
			var err error
			excelStatement, err = repo.CreateStatement(ctx, models.CreateStatementInput{
				AccountId:        accountID,
				CreatedBy:        userID,
				OriginalFilename: "excel-statement.xlsx",
				FileType:         "excel",
				Status:           models.StatementStatusDone,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("lists statements for the user", func() {
			statements, err := repo.ListStatementByUserId(ctx, userID, 10, 0, models.StatementListQuery{})
			Expect(err).NotTo(HaveOccurred())
			Expect(statements).To(HaveLen(2))

			total, err := repo.CountStatementsByUserId(ctx, userID, models.StatementListQuery{})
			Expect(err).NotTo(HaveOccurred())
			Expect(total).To(Equal(2))
		})

		It("filters by account", func() {
			otherAccountID := createRepoTestAccount(ctx, db, cfg.DBSchema, userID)
			_, err := repo.CreateStatement(ctx, models.CreateStatementInput{
				AccountId:        otherAccountID,
				CreatedBy:        userID,
				OriginalFilename: "other-account.csv",
				FileType:         "csv",
				Status:           models.StatementStatusPending,
			})
			Expect(err).NotTo(HaveOccurred())

			query := models.StatementListQuery{AccountId: &accountID}
			statements, err := repo.ListStatementByUserId(ctx, userID, 10, 0, query)
			Expect(err).NotTo(HaveOccurred())
			Expect(statements).To(HaveLen(2))

			total, err := repo.CountStatementsByUserId(ctx, userID, query)
			Expect(err).NotTo(HaveOccurred())
			Expect(total).To(Equal(2))
		})

		It("filters by filename search", func() {
			search := "excel"
			query := models.StatementListQuery{Search: &search}

			statements, err := repo.ListStatementByUserId(ctx, userID, 10, 0, query)
			Expect(err).NotTo(HaveOccurred())
			Expect(statements).To(HaveLen(1))
			Expect(statements[0].Id).To(Equal(excelStatement.Id))

			total, err := repo.CountStatementsByUserId(ctx, userID, query)
			Expect(err).NotTo(HaveOccurred())
			Expect(total).To(Equal(1))
		})

		It("filters by creation date", func() {
			yesterday := time.Now().UTC().Add(-24 * time.Hour).Truncate(time.Second)
			_, err := db.ExecuteQuery(ctx,
				fmt.Sprintf("UPDATE %s.statement SET created_at = $1 WHERE id = $2", cfg.DBSchema),
				yesterday, csvStatement.Id,
			)
			Expect(err).NotTo(HaveOccurred())

			from := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
			query := models.StatementListQuery{DateFrom: &from}

			statements, err := repo.ListStatementByUserId(ctx, userID, 10, 0, query)
			Expect(err).NotTo(HaveOccurred())
			Expect(statements).To(HaveLen(1))
			Expect(statements[0].Id).To(Equal(excelStatement.Id))

			total, err := repo.CountStatementsByUserId(ctx, userID, query)
			Expect(err).NotTo(HaveOccurred())
			Expect(total).To(Equal(1))

			until := time.Now().UTC().Add(-time.Hour).Truncate(time.Second)
			query = models.StatementListQuery{DateTo: &until}
			statements, err = repo.ListStatementByUserId(ctx, userID, 10, 0, query)
			Expect(err).NotTo(HaveOccurred())
			Expect(statements).To(HaveLen(1))
			Expect(statements[0].Id).To(Equal(csvStatement.Id))

			total, err = repo.CountStatementsByUserId(ctx, userID, query)
			Expect(err).NotTo(HaveOccurred())
			Expect(total).To(Equal(1))
		})

		It("paginates and excludes soft deleted statements", func() {
			page, err := repo.ListStatementByUserId(ctx, userID, 1, 0, models.StatementListQuery{})
			Expect(err).NotTo(HaveOccurred())
			Expect(page).To(HaveLen(1))

			secondPage, err := repo.ListStatementByUserId(ctx, userID, 1, 1, models.StatementListQuery{})
			Expect(err).NotTo(HaveOccurred())
			Expect(secondPage).To(HaveLen(1))

			_, err = db.ExecuteQuery(ctx,
				fmt.Sprintf("UPDATE %s.statement SET deleted_at = NOW() WHERE id = $1", cfg.DBSchema),
				secondPage[0].Id,
			)
			Expect(err).NotTo(HaveOccurred())

			total, err := repo.CountStatementsByUserId(ctx, userID, models.StatementListQuery{})
			Expect(err).NotTo(HaveOccurred())
			Expect(total).To(Equal(1))

			remaining, err := repo.ListStatementByUserId(ctx, userID, 10, 0, models.StatementListQuery{})
			Expect(err).NotTo(HaveOccurred())
			Expect(remaining).To(HaveLen(1))
			Expect(remaining[0].Id).NotTo(Equal(secondPage[0].Id))
		})
	})
})
