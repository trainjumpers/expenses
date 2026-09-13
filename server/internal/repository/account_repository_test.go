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

var _ = Describe("AccountRepository", func() {
	var (
		ctx    context.Context
		cfg    *config.Config
		db     databasemanager.DatabaseManager
		repo   AccountRepositoryInterface
		userID int64
	)

	BeforeEach(func() {
		var err error
		cfg, err = config.NewConfig()
		Expect(err).NotTo(HaveOccurred())

		db, err = databasemanager.NewDatabaseManager(cfg)
		Expect(err).NotTo(HaveOccurred())

		repo = NewAccountRepository(db, cfg)
		ctx = context.Background()
		userID = createRepoTestUser(ctx, db, cfg.DBSchema)
	})

	AfterEach(func() {
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.transaction WHERE created_by = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.account WHERE created_by = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.user WHERE id = $1", cfg.DBSchema), userID)
		Expect(db.Close()).To(Succeed())
	})

	Describe("CreateAccount", func() {
		It("creates a plain account with a balance", func() {
			balance := 1500.75
			currentValue := 999.0

			account, err := repo.CreateAccount(ctx, models.CreateAccountInput{
				Name:         "Savings",
				BankType:     models.BankTypeSBI,
				Currency:     models.CurrencyINR,
				Balance:      &balance,
				CurrentValue: &currentValue,
				CreatedBy:    userID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(account.Id).NotTo(BeZero())
			Expect(account.Balance).To(Equal(balance))
			Expect(account.CurrentValue).To(BeNil())
		})

		It("creates an investment account without a current value", func() {
			account, err := repo.CreateAccount(ctx, models.CreateAccountInput{
				Name:      "Stocks",
				BankType:  models.BankTypeInvestment,
				Currency:  models.CurrencyINR,
				CreatedBy: userID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(account.CurrentValue).To(BeNil())
		})

		It("creates an investment account with a current value", func() {
			currentValue := 25000.50
			account, err := repo.CreateAccount(ctx, models.CreateAccountInput{
				Name:         "Mutual funds",
				BankType:     models.BankTypeInvestment,
				Currency:     models.CurrencyINR,
				CurrentValue: &currentValue,
				CreatedBy:    userID,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(account.CurrentValue).NotTo(BeNil())
			Expect(*account.CurrentValue).To(Equal(currentValue))

			fetched, err := repo.GetAccountById(ctx, account.Id, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetched.CurrentValue).NotTo(BeNil())
			Expect(*fetched.CurrentValue).To(Equal(currentValue))
		})

		It("returns an error for an unknown user", func() {
			_, err := repo.CreateAccount(ctx, models.CreateAccountInput{
				Name:      "Orphan",
				BankType:  models.BankTypeOthers,
				Currency:  models.CurrencyINR,
				CreatedBy: 999999999,
			})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetAccountById", func() {
		It("returns not found for an unknown account", func() {
			_, err := repo.GetAccountById(ctx, 999999, userID)
			expectAuthErrorType(err, "AccountNotFound")
		})

		It("hides accounts from another user", func() {
			created, err := repo.CreateAccount(ctx, models.CreateAccountInput{
				Name:      "Private",
				BankType:  models.BankTypeOthers,
				Currency:  models.CurrencyINR,
				CreatedBy: userID,
			})
			Expect(err).NotTo(HaveOccurred())

			otherUserID := createRepoTestUser(ctx, db, cfg.DBSchema)
			defer func() {
				_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.user WHERE id = $1", cfg.DBSchema), otherUserID)
			}()

			_, err = repo.GetAccountById(ctx, created.Id, otherUserID)
			expectAuthErrorType(err, "AccountNotFound")
		})
	})

	Describe("UpdateAccount", func() {
		It("updates name and balance", func() {
			created, err := repo.CreateAccount(ctx, models.CreateAccountInput{
				Name:      "Before",
				BankType:  models.BankTypeAxis,
				Currency:  models.CurrencyINR,
				CreatedBy: userID,
			})
			Expect(err).NotTo(HaveOccurred())

			newBalance := 200.25
			updated, err := repo.UpdateAccount(ctx, created.Id, userID, models.UpdateAccountInput{
				Name:    "After",
				Balance: &newBalance,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(updated.Name).To(Equal("After"))
			Expect(updated.Balance).To(Equal(newBalance))
		})

		It("upserts the current value of an investment account", func() {
			initialValue := 100.0
			created, err := repo.CreateAccount(ctx, models.CreateAccountInput{
				Name:         "Investment",
				BankType:     models.BankTypeInvestment,
				Currency:     models.CurrencyINR,
				CurrentValue: &initialValue,
				CreatedBy:    userID,
			})
			Expect(err).NotTo(HaveOccurred())

			nextValue := 175.0
			updated, err := repo.UpdateAccount(ctx, created.Id, userID, models.UpdateAccountInput{CurrentValue: &nextValue})
			Expect(err).NotTo(HaveOccurred())
			Expect(updated.CurrentValue).NotTo(BeNil())
			Expect(*updated.CurrentValue).To(Equal(nextValue))

			var stored float64
			Expect(db.FetchOne(ctx,
				fmt.Sprintf("SELECT current_value FROM %s.investment_account_value WHERE account_id = $1", cfg.DBSchema),
				created.Id,
			).Scan(&stored)).To(Succeed())
			Expect(stored).To(Equal(nextValue))
		})

		It("clears the current value when leaving investment type", func() {
			currentValue := 500.0
			created, err := repo.CreateAccount(ctx, models.CreateAccountInput{
				Name:         "Was investment",
				BankType:     models.BankTypeInvestment,
				Currency:     models.CurrencyINR,
				CurrentValue: &currentValue,
				CreatedBy:    userID,
			})
			Expect(err).NotTo(HaveOccurred())

			updated, err := repo.UpdateAccount(ctx, created.Id, userID, models.UpdateAccountInput{BankType: models.BankTypeAxis})
			Expect(err).NotTo(HaveOccurred())
			Expect(updated.BankType).To(Equal(models.BankTypeAxis))
			Expect(updated.CurrentValue).To(BeNil())

			var count int
			Expect(db.FetchOne(ctx,
				fmt.Sprintf("SELECT COUNT(*) FROM %s.investment_account_value WHERE account_id = $1", cfg.DBSchema),
				created.Id,
			).Scan(&count)).To(Succeed())
			Expect(count).To(Equal(0))
		})

		It("returns the account unchanged for an empty update", func() {
			created, err := repo.CreateAccount(ctx, models.CreateAccountInput{
				Name:      "Empty update",
				BankType:  models.BankTypeOthers,
				Currency:  models.CurrencyINR,
				CreatedBy: userID,
			})
			Expect(err).NotTo(HaveOccurred())

			updated, err := repo.UpdateAccount(ctx, created.Id, userID, models.UpdateAccountInput{})
			Expect(err).NotTo(HaveOccurred())
			Expect(updated.Id).To(Equal(created.Id))
			Expect(updated.Name).To(Equal("Empty update"))
		})

		It("returns not found for an unknown account", func() {
			balance := 10.0
			_, err := repo.UpdateAccount(ctx, 999999, userID, models.UpdateAccountInput{Balance: &balance})
			expectAuthErrorType(err, "AccountNotFound")
		})
	})

	Describe("DeleteAccount", func() {
		It("deletes an account once", func() {
			created, err := repo.CreateAccount(ctx, models.CreateAccountInput{
				Name:      "Delete me",
				BankType:  models.BankTypeOthers,
				Currency:  models.CurrencyINR,
				CreatedBy: userID,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.DeleteAccount(ctx, created.Id, userID)).To(Succeed())

			err = repo.DeleteAccount(ctx, created.Id, userID)
			expectAuthErrorType(err, "AccountNotFound")
		})

		It("refuses to delete an account that has transactions", func() {
			created, err := repo.CreateAccount(ctx, models.CreateAccountInput{
				Name:      "Busy account",
				BankType:  models.BankTypeOthers,
				Currency:  models.CurrencyINR,
				CreatedBy: userID,
			})
			Expect(err).NotTo(HaveOccurred())

			Expect(db.FetchOne(ctx,
				fmt.Sprintf("INSERT INTO %s.transaction (name, amount, date, account_id, created_by) VALUES ($1, $2, $3, $4, $5) RETURNING id", cfg.DBSchema),
				"Blocking transaction", 1.0, time.Now().UTC(), created.Id, userID,
			).Scan(new(int64))).To(Succeed())

			err = repo.DeleteAccount(ctx, created.Id, userID)
			expectAuthErrorType(err, "AccountHasTransactions")
		})

		It("returns not found for an unknown account", func() {
			err := repo.DeleteAccount(ctx, 999999, userID)
			expectAuthErrorType(err, "AccountNotFound")
		})
	})

	Describe("ListAccounts", func() {
		It("returns an empty list when the user has no accounts", func() {
			accounts, err := repo.ListAccounts(ctx, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(accounts).To(BeEmpty())
		})

		It("lists accounts newest first and includes investment values", func() {
			first, err := repo.CreateAccount(ctx, models.CreateAccountInput{
				Name:      "Older",
				BankType:  models.BankTypeSBI,
				Currency:  models.CurrencyINR,
				CreatedBy: userID,
			})
			Expect(err).NotTo(HaveOccurred())

			currentValue := 4200.0
			second, err := repo.CreateAccount(ctx, models.CreateAccountInput{
				Name:         "Newer investment",
				BankType:     models.BankTypeInvestment,
				Currency:     models.CurrencyINR,
				CurrentValue: &currentValue,
				CreatedBy:    userID,
			})
			Expect(err).NotTo(HaveOccurred())

			_, err = db.ExecuteQuery(ctx,
				fmt.Sprintf("UPDATE %s.account SET created_at = $1 WHERE id = $2", cfg.DBSchema),
				time.Now().UTC().Add(-time.Hour), first.Id,
			)
			Expect(err).NotTo(HaveOccurred())

			accounts, err := repo.ListAccounts(ctx, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(accounts).To(HaveLen(2))
			Expect(accounts[0].Id).To(Equal(second.Id))
			Expect(accounts[0].CurrentValue).NotTo(BeNil())
			Expect(*accounts[0].CurrentValue).To(Equal(currentValue))
			Expect(accounts[1].Id).To(Equal(first.Id))
			Expect(accounts[1].CurrentValue).To(BeNil())
		})
	})
})
