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

var _ = Describe("TransactionRepository", func() {
	var (
		ctx        context.Context
		cfg        *config.Config
		db         databasemanager.DatabaseManager
		repo       TransactionRepositoryInterface
		userID     int64
		accountID  int64
		otherAccID int64
	)

	BeforeEach(func() {
		var err error
		cfg, err = config.NewConfig()
		Expect(err).NotTo(HaveOccurred())

		db, err = databasemanager.NewDatabaseManager(cfg)
		Expect(err).NotTo(HaveOccurred())

		repo = NewTransactionRepository(db, cfg)
		ctx = context.Background()
		userID = createRepoTestUser(ctx, db, cfg.DBSchema)
		accountID = createRepoTestAccount(ctx, db, cfg.DBSchema, userID)
		otherAccID = createRepoTestAccount(ctx, db, cfg.DBSchema, userID)
	})

	AfterEach(func() {
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf(`DELETE FROM %s.transaction_category_mapping WHERE transaction_id IN (SELECT id FROM %s.transaction WHERE created_by = $1)`, cfg.DBSchema, cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.transaction WHERE created_by = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.categories WHERE created_by = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.account WHERE created_by = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.user WHERE id = $1", cfg.DBSchema), userID)
		Expect(db.Close()).To(Succeed())
	})

	createCategory := func(name string) int64 {
		var id int64
		Expect(db.FetchOne(ctx,
			fmt.Sprintf("INSERT INTO %s.categories (name, created_by) VALUES ($1, $2) RETURNING id", cfg.DBSchema),
			name, userID,
		).Scan(&id)).To(Succeed())
		return id
	}

	txInput := func(name string, amount float64, account int64) models.CreateBaseTransactionInput {
		return models.CreateBaseTransactionInput{
			Name:      name,
			Amount:    &amount,
			Date:      time.Now().UTC().Truncate(time.Second),
			CreatedBy: userID,
			AccountId: account,
		}
	}

	Describe("CreateTransaction", func() {
		It("stores the transaction and its category mappings", func() {
			categoryID := createCategory("Create txn category")
			input := txInput("Coffee", 120.5, accountID)

			created, err := repo.CreateTransaction(ctx, input, []int64{categoryID})
			Expect(err).NotTo(HaveOccurred())
			Expect(created.Id).NotTo(BeZero())
			Expect(created.CategoryIds).To(Equal([]int64{categoryID}))

			fetched, err := repo.GetTransactionById(ctx, created.Id, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetched.Name).To(Equal("Coffee"))
			Expect(fetched.Amount).To(Equal(120.5))
			Expect(fetched.CategoryIds).To(Equal([]int64{categoryID}))
		})

		It("rejects a duplicate transaction", func() {
			input := txInput("Duplicate", 50, accountID)
			_, err := repo.CreateTransaction(ctx, input, nil)
			Expect(err).NotTo(HaveOccurred())

			_, err = repo.CreateTransaction(ctx, input, nil)
			expectAuthErrorType(err, "TransactionAlreadyExists")
		})

		It("returns a category error for an unknown category", func() {
			_, err := repo.CreateTransaction(ctx, txInput("Bad category", 10, accountID), []int64{999999})
			expectAuthErrorType(err, "CategoryNotFound")
		})

		It("returns an error for an unknown account", func() {
			_, err := repo.CreateTransaction(ctx, txInput("Bad account", 10, 999999), nil)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("CreateTransactions", func() {
		It("returns an empty result for no transactions", func() {
			created, err := repo.CreateTransactions(ctx, nil, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(created).To(BeEmpty())
		})

		It("rejects mismatched category id slices", func() {
			_, err := repo.CreateTransactions(ctx, []models.CreateBaseTransactionInput{txInput("Mismatch", 1, accountID)}, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("same length"))
		})

		It("stores a batch with category mappings", func() {
			categoryID := createCategory("Batch category")
			inputs := []models.CreateBaseTransactionInput{
				txInput("Batch one", 10, accountID),
				txInput("Batch two", 20, accountID),
				txInput("Batch three", 30, otherAccID),
			}

			created, err := repo.CreateTransactions(ctx, inputs, [][]int64{{categoryID}, nil, {}})
			Expect(err).NotTo(HaveOccurred())
			Expect(created).To(HaveLen(3))

			byName := map[string]models.TransactionResponse{}
			for _, tx := range created {
				byName[tx.Name] = tx
			}
			Expect(byName["Batch one"].CategoryIds).To(Equal([]int64{categoryID}))
			Expect(byName["Batch two"].CategoryIds).To(BeEmpty())
		})

		It("returns a category error for an unknown category in the batch", func() {
			inputs := []models.CreateBaseTransactionInput{txInput("Batch bad category", 5, accountID)}
			_, err := repo.CreateTransactions(ctx, inputs, [][]int64{{999999}})
			expectAuthErrorType(err, "CategoryNotFound")
		})
	})

	Describe("GetTransactionById and GetTransactionsByIds", func() {
		It("finds a stored transaction and hides deleted ones", func() {
			created, err := repo.CreateTransaction(ctx, txInput("Find me", 75, accountID), nil)
			Expect(err).NotTo(HaveOccurred())

			fetched, err := repo.GetTransactionById(ctx, created.Id, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetched.Id).To(Equal(created.Id))

			Expect(repo.DeleteTransaction(ctx, created.Id, userID)).To(Succeed())

			_, err = repo.GetTransactionById(ctx, created.Id, userID)
			expectAuthErrorType(err, "TransactionNotFound")
		})

		It("returns not found for an unknown id", func() {
			_, err := repo.GetTransactionById(ctx, 999999, userID)
			expectAuthErrorType(err, "TransactionNotFound")
		})

		It("fetches multiple transactions by id", func() {
			first, err := repo.CreateTransaction(ctx, txInput("Multi one", 1, accountID), nil)
			Expect(err).NotTo(HaveOccurred())
			second, err := repo.CreateTransaction(ctx, txInput("Multi two", 2, accountID), nil)
			Expect(err).NotTo(HaveOccurred())

			found, err := repo.GetTransactionsByIds(ctx, []int64{first.Id, second.Id, 999999}, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(HaveLen(2))
		})

		It("returns an empty list for no ids", func() {
			found, err := repo.GetTransactionsByIds(ctx, nil, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(found).To(BeEmpty())
		})
	})

	Describe("UpdateTransaction", func() {
		It("updates fields and rejects duplicates", func() {
			original, err := repo.CreateTransaction(ctx, txInput("Original", 100, accountID), nil)
			Expect(err).NotTo(HaveOccurred())
			other, err := repo.CreateTransaction(ctx, txInput("Other", 200, accountID), nil)
			Expect(err).NotTo(HaveOccurred())

			newName := "Renamed"
			newAmount := 250.0
			err = repo.UpdateTransaction(ctx, other.Id, userID, models.UpdateBaseTransactionInput{
				Name:   newName,
				Amount: &newAmount,
			})
			Expect(err).NotTo(HaveOccurred())

			fetched, err := repo.GetTransactionById(ctx, other.Id, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetched.Name).To(Equal(newName))
			Expect(fetched.Amount).To(Equal(newAmount))

			duplicateAmount := original.Amount
			duplicate := models.UpdateBaseTransactionInput{Name: original.Name, Amount: &duplicateAmount}
			err = repo.UpdateTransaction(ctx, other.Id, userID, duplicate)
			expectAuthErrorType(err, "TransactionAlreadyExists")
		})

		It("returns not found for an unknown transaction", func() {
			name := "Missing"
			err := repo.UpdateTransaction(ctx, 999999, userID, models.UpdateBaseTransactionInput{Name: name})
			expectAuthErrorType(err, "TransactionNotFound")
		})

		It("rejects an empty update", func() {
			created, err := repo.CreateTransaction(ctx, txInput("Empty update", 100, accountID), nil)
			Expect(err).NotTo(HaveOccurred())

			err = repo.UpdateTransaction(ctx, created.Id, userID, models.UpdateBaseTransactionInput{})
			expectAuthErrorType(err, "NoFieldsToUpdate")
		})
	})

	Describe("DeleteTransaction", func() {
		It("soft deletes a transaction once", func() {
			created, err := repo.CreateTransaction(ctx, txInput("Delete me", 42, accountID), nil)
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.DeleteTransaction(ctx, created.Id, userID)).To(Succeed())

			err = repo.DeleteTransaction(ctx, created.Id, userID)
			expectAuthErrorType(err, "TransactionNotFound")
		})

		It("returns not found for an unknown transaction", func() {
			err := repo.DeleteTransaction(ctx, 999999, userID)
			expectAuthErrorType(err, "TransactionNotFound")
		})
	})

	Describe("UpdateCategoryMapping", func() {
		It("replaces category mappings", func() {
			categoryOne := createCategory("Mapping one")
			categoryTwo := createCategory("Mapping two")
			created, err := repo.CreateTransaction(ctx, txInput("Mappings", 15, accountID), []int64{categoryOne})
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.UpdateCategoryMapping(ctx, created.Id, userID, []int64{categoryOne, categoryTwo})).To(Succeed())

			fetched, err := repo.GetTransactionById(ctx, created.Id, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetched.CategoryIds).To(ConsistOf(categoryOne, categoryTwo))

			Expect(repo.UpdateCategoryMapping(ctx, created.Id, userID, nil)).To(Succeed())

			fetched, err = repo.GetTransactionById(ctx, created.Id, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetched.CategoryIds).To(BeEmpty())
		})

		It("returns an error for an unknown category", func() {
			created, err := repo.CreateTransaction(ctx, txInput("Bad mapping", 15, accountID), nil)
			Expect(err).NotTo(HaveOccurred())

			err = repo.UpdateCategoryMapping(ctx, created.Id, userID, []int64{999999})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ListTransactions", func() {
		var (
			categoryID    int64
			uncategorized models.TransactionResponse
			foodTx        models.TransactionResponse
			salaryTx      models.TransactionResponse
		)

		BeforeEach(func() {
			categoryID = createCategory("List category")

			description := "team lunch"
			lunch := txInput("Lunch", 500, accountID)
			lunch.Description = description

			var err error
			foodTx, err = repo.CreateTransaction(ctx, lunch, []int64{categoryID})
			Expect(err).NotTo(HaveOccurred())
			salaryTx, err = repo.CreateTransaction(ctx, txInput("Salary", 50000, otherAccID), nil)
			Expect(err).NotTo(HaveOccurred())
			uncategorized, err = repo.CreateTransaction(ctx, txInput("Misc", 300, accountID), nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("filters by account", func() {
			result, err := repo.ListTransactions(ctx, userID, models.TransactionListQuery{Page: 1, PageSize: 10, AccountId: &otherAccID})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(1))
			Expect(result.Transactions[0].Id).To(Equal(salaryTx.Id))
		})

		It("filters by amount range", func() {
			minAmount := 400.0
			maxAmount := 600.0
			result, err := repo.ListTransactions(ctx, userID, models.TransactionListQuery{Page: 1, PageSize: 10, MinAmount: &minAmount, MaxAmount: &maxAmount})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(1))
			Expect(result.Transactions[0].Id).To(Equal(foodTx.Id))
		})

		It("filters by date range", func() {
			yesterday := time.Now().UTC().Add(-24 * time.Hour)
			tomorrow := time.Now().UTC().Add(24 * time.Hour)

			result, err := repo.ListTransactions(ctx, userID, models.TransactionListQuery{Page: 1, PageSize: 10, DateFrom: &yesterday, DateTo: &tomorrow})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(3))

			lastWeek := time.Now().UTC().Add(-7 * 24 * time.Hour)
			pastOnly, err := repo.ListTransactions(ctx, userID, models.TransactionListQuery{Page: 1, PageSize: 10, DateTo: &lastWeek})
			Expect(err).NotTo(HaveOccurred())
			Expect(pastOnly.Total).To(Equal(0))
		})

		It("searches by name and description", func() {
			search := "lunch"
			result, err := repo.ListTransactions(ctx, userID, models.TransactionListQuery{Page: 1, PageSize: 10, Search: &search})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(1))
			Expect(result.Transactions[0].Id).To(Equal(foodTx.Id))

			empty := ""
			all, err := repo.ListTransactions(ctx, userID, models.TransactionListQuery{Page: 1, PageSize: 10, Search: &empty})
			Expect(err).NotTo(HaveOccurred())
			Expect(all.Total).To(Equal(3))
		})

		It("filters by category and uncategorized", func() {
			byCategory, err := repo.ListTransactions(ctx, userID, models.TransactionListQuery{Page: 1, PageSize: 10, CategoryId: &categoryID})
			Expect(err).NotTo(HaveOccurred())
			Expect(byCategory.Total).To(Equal(1))
			Expect(byCategory.Transactions[0].Id).To(Equal(foodTx.Id))

			uncategorizedOnly := true
			uncategorizedResult, err := repo.ListTransactions(ctx, userID, models.TransactionListQuery{Page: 1, PageSize: 10, Uncategorized: &uncategorizedOnly})
			Expect(err).NotTo(HaveOccurred())
			Expect(uncategorizedResult.Total).To(Equal(2))
			ids := []int64{uncategorizedResult.Transactions[0].Id, uncategorizedResult.Transactions[1].Id}
			Expect(ids).To(ConsistOf(salaryTx.Id, uncategorized.Id))
		})

		It("filters by statement", func() {
			statementID := createRepoTestStatement(ctx, db, cfg.DBSchema, userID, accountID)
			Expect(db.FetchOne(ctx,
				fmt.Sprintf("INSERT INTO %s.statement_transaction_mapping (statement_id, transaction_id) VALUES ($1, $2) RETURNING id", cfg.DBSchema),
				statementID, foodTx.Id,
			).Scan(new(int64))).To(Succeed())
			defer func() {
				_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.statement_transaction_mapping WHERE statement_id = $1", cfg.DBSchema), statementID)
				_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.statement WHERE id = $1", cfg.DBSchema), statementID)
			}()

			result, err := repo.ListTransactions(ctx, userID, models.TransactionListQuery{Page: 1, PageSize: 10, StatementId: &statementID})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Total).To(Equal(1))
			Expect(result.Transactions[0].Id).To(Equal(foodTx.Id))
		})

		It("sorts by amount and ignores unknown sort columns", func() {
			ascending := "asc"
			result, err := repo.ListTransactions(ctx, userID, models.TransactionListQuery{Page: 1, PageSize: 10, SortBy: "amount", SortOrder: ascending})
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Transactions).To(HaveLen(3))
			Expect(result.Transactions[0].Amount).To(Equal(300.0))
			Expect(result.Transactions[2].Amount).To(Equal(50000.0))

			unknownSort, err := repo.ListTransactions(ctx, userID, models.TransactionListQuery{Page: 1, PageSize: 10, SortBy: "not-a-column"})
			Expect(err).NotTo(HaveOccurred())
			Expect(unknownSort.Transactions).To(HaveLen(3))
		})

		It("paginates results", func() {
			firstPage, err := repo.ListTransactions(ctx, userID, models.TransactionListQuery{Page: 1, PageSize: 2, SortBy: "amount", SortOrder: "asc"})
			Expect(err).NotTo(HaveOccurred())
			Expect(firstPage.Total).To(Equal(3))
			Expect(firstPage.Transactions).To(HaveLen(2))
			Expect(firstPage.Page).To(Equal(1))

			secondPage, err := repo.ListTransactions(ctx, userID, models.TransactionListQuery{Page: 2, PageSize: 2, SortBy: "amount", SortOrder: "asc"})
			Expect(err).NotTo(HaveOccurred())
			Expect(secondPage.Transactions).To(HaveLen(1))
		})
	})
})

func createRepoTestStatement(ctx context.Context, db databasemanager.DatabaseManager, schema string, userID, accountID int64) int64 {
	var id int64
	Expect(db.FetchOne(ctx,
		fmt.Sprintf("INSERT INTO %s.statement (account_id, created_by, original_filename, file_type, status) VALUES ($1, $2, $3, $4, $5) RETURNING id", schema),
		accountID, userID, "list-filter.csv", "csv", "done",
	).Scan(&id)).To(Succeed())
	return id
}
