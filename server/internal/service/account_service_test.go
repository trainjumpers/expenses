package service

import (
	"context"
	mock "expenses/internal/mock/repository"
	"expenses/internal/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AccountService", func() {
	var (
		accountService AccountServiceInterface
		mockRepo       *mock.MockAccountRepository
		ctx            context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockRepo = mock.NewMockAccountRepository()
		accountService = NewAccountService(mockRepo)
	})

	Describe("CreateAccount", func() {
		It("should create a new account with default balance if not provided", func() {
			input := models.CreateAccountInput{
				Name:      "Test Account",
				BankType:  models.BankTypeAxis,
				Currency:  models.CurrencyINR,
				CreatedBy: 1,
			}
			acc, err := accountService.CreateAccount(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.Name).To(Equal(input.Name))
			Expect(acc.Balance).To(Equal(0.0))
			Expect(acc.CurrentValue).To(BeNil())
		})
		It("should create a new account with provided balance", func() {
			bal := 100.5
			input := models.CreateAccountInput{
				Name:      "Test Account 2",
				BankType:  models.BankTypeSBI,
				Currency:  models.CurrencyUSD,
				Balance:   &bal,
				CreatedBy: 2,
			}
			acc, err := accountService.CreateAccount(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.Balance).To(Equal(bal))
			Expect(acc.CurrentValue).To(BeNil())
		})
		It("should create a new account with 'others' bank type", func() {
			input := models.CreateAccountInput{
				Name:      "Others Bank Account",
				BankType:  models.BankTypeOthers,
				Currency:  models.CurrencyINR,
				CreatedBy: 1,
			}
			acc, err := accountService.CreateAccount(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.Name).To(Equal(input.Name))
			Expect(acc.BankType).To(Equal(models.BankTypeOthers))
			Expect(acc.Balance).To(Equal(0.0))
			Expect(acc.CurrentValue).To(BeNil())
		})
		It("should create an investment account without current value", func() {
			input := models.CreateAccountInput{
				Name:      "Investment Account",
				BankType:  models.BankTypeInvestment,
				Currency:  models.CurrencyINR,
				CreatedBy: 1,
			}
			acc, err := accountService.CreateAccount(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.BankType).To(Equal(models.BankTypeInvestment))
			Expect(acc.CurrentValue).To(BeNil())
		})
		It("should create an investment account with current value", func() {
			currentValue := 15000.0
			input := models.CreateAccountInput{
				Name:         "Investment Account with Value",
				BankType:     models.BankTypeInvestment,
				Currency:     models.CurrencyINR,
				CurrentValue: &currentValue,
				CreatedBy:    2,
			}
			acc, err := accountService.CreateAccount(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.BankType).To(Equal(models.BankTypeInvestment))
			Expect(acc.CurrentValue).NotTo(BeNil())
			Expect(*acc.CurrentValue).To(Equal(currentValue))
		})
		It("should ignore current value for non-investment account", func() {
			currentValue := 5000.0
			input := models.CreateAccountInput{
				Name:         "Regular Account",
				BankType:     models.BankTypeAxis,
				Currency:     models.CurrencyINR,
				CurrentValue: &currentValue,
				CreatedBy:    3,
			}
			acc, err := accountService.CreateAccount(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.BankType).To(Equal(models.BankTypeAxis))
			Expect(acc.CurrentValue).To(BeNil())
		})
	})

	Describe("GetAccountById", func() {
		var created models.AccountResponse
		BeforeEach(func() {
			input := models.CreateAccountInput{
				Name:      "Account Get",
				BankType:  models.BankTypeHDFC,
				Currency:  models.CurrencyINR,
				CreatedBy: 3,
			}
			var err error
			created, err = accountService.CreateAccount(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should get account by id", func() {
			acc, err := accountService.GetAccountById(ctx, created.Id, 3)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.Name).To(Equal("Account Get"))
		})
		It("should return error for non-existent id", func() {
			_, err := accountService.GetAccountById(ctx, 9999, 3)
			Expect(err).To(HaveOccurred())
		})
		It("should return error for non-existent user id", func() {
			_, err := accountService.GetAccountById(ctx, created.Id, 9999)
			Expect(err).To(HaveOccurred())
		})
		It("should return error while accessing account of other user", func() {
			_, err := accountService.GetAccountById(ctx, created.Id, 4)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("UpdateAccount", func() {
		var created models.AccountResponse
		BeforeEach(func() {
			input := models.CreateAccountInput{
				Name:      "Account Update",
				BankType:  models.BankTypeICICI,
				Currency:  models.CurrencyUSD,
				CreatedBy: 4,
			}
			var err error
			created, err = accountService.CreateAccount(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should update account name", func() {
			update := models.UpdateAccountInput{Name: "Updated Name"}
			acc, err := accountService.UpdateAccount(ctx, created.Id, 4, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.Name).To(Equal("Updated Name"))
		})
		It("should update account balance", func() {
			balance := 100.5
			update := models.UpdateAccountInput{Balance: &balance}
			acc, err := accountService.UpdateAccount(ctx, created.Id, 4, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.Balance).To(Equal(100.5))
		})
		It("should update account bank type", func() {
			update := models.UpdateAccountInput{BankType: models.BankTypeHDFC}
			acc, err := accountService.UpdateAccount(ctx, created.Id, 4, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.BankType).To(Equal(models.BankTypeHDFC))
		})
		It("should update account bank type to 'others'", func() {
			update := models.UpdateAccountInput{BankType: models.BankTypeOthers}
			acc, err := accountService.UpdateAccount(ctx, created.Id, 4, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.BankType).To(Equal(models.BankTypeOthers))
		})
		It("should update account currency", func() {
			update := models.UpdateAccountInput{Currency: models.CurrencyUSD}
			acc, err := accountService.UpdateAccount(ctx, created.Id, 4, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.Currency).To(Equal(models.CurrencyUSD))
		})
		It("should set account balance to 0 if provided balance is 0", func() {
			balance := 0.0
			update := models.UpdateAccountInput{Balance: &balance}
			acc, err := accountService.UpdateAccount(ctx, created.Id, 4, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.Balance).To(Equal(0.0))
		})
		It("should return error for non-existent id", func() {
			update := models.UpdateAccountInput{Name: "Updated Name"}
			_, err := accountService.UpdateAccount(ctx, 9999, 4, update)
			Expect(err).To(HaveOccurred())
		})
		It("should return error for non-existent user id", func() {
			update := models.UpdateAccountInput{Name: "Updated Name"}
			_, err := accountService.UpdateAccount(ctx, created.Id, 9999, update)
			Expect(err).To(HaveOccurred())
		})
		It("should return error while updating account of other user", func() {
			update := models.UpdateAccountInput{Name: "Updated Name"}
			_, err := accountService.UpdateAccount(ctx, created.Id, 5, update)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("UpdateAccount with Investment", func() {
		var createdInvestment models.AccountResponse
		var createdRegular models.AccountResponse
		BeforeEach(func() {
			currentValue := 10000.0
			investmentInput := models.CreateAccountInput{
				Name:         "Investment Account",
				BankType:     models.BankTypeInvestment,
				Currency:     models.CurrencyINR,
				CurrentValue: &currentValue,
				CreatedBy:    6,
			}
			var err error
			createdInvestment, err = accountService.CreateAccount(ctx, investmentInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(createdInvestment.CurrentValue).NotTo(BeNil())

			regularInput := models.CreateAccountInput{
				Name:      "Regular Account",
				BankType:  models.BankTypeAxis,
				Currency:  models.CurrencyINR,
				CreatedBy: 6,
			}
			createdRegular, err = accountService.CreateAccount(ctx, regularInput)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should set current value for investment account", func() {
			newValue := 15000.0
			update := models.UpdateAccountInput{CurrentValue: &newValue}
			acc, err := accountService.UpdateAccount(ctx, createdInvestment.Id, 6, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.CurrentValue).NotTo(BeNil())
			Expect(*acc.CurrentValue).To(Equal(newValue))
		})
		It("should update current value from initial to new value", func() {
			initialValue := *createdInvestment.CurrentValue
			Expect(initialValue).NotTo(Equal(0.0))

			newValue := initialValue + 5000.0
			update := models.UpdateAccountInput{CurrentValue: &newValue}
			acc, err := accountService.UpdateAccount(ctx, createdInvestment.Id, 6, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(*acc.CurrentValue).To(Equal(newValue))
			Expect(*acc.CurrentValue).NotTo(Equal(initialValue))
		})
		It("should not set current value for non-investment account", func() {
			value := 20000.0
			update := models.UpdateAccountInput{CurrentValue: &value}
			acc, err := accountService.UpdateAccount(ctx, createdRegular.Id, 6, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.CurrentValue).To(BeNil())
		})
		It("should clear current value when changing bank type from investment to regular", func() {
			update := models.UpdateAccountInput{
				BankType:     models.BankTypeAxis,
				CurrentValue: nil,
			}
			acc, err := accountService.UpdateAccount(ctx, createdInvestment.Id, 6, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.BankType).To(Equal(models.BankTypeAxis))
			Expect(acc.CurrentValue).To(BeNil())
		})
		It("should update regular account to investment and set current value", func() {
			value := 12000.0
			update := models.UpdateAccountInput{
				BankType:     models.BankTypeInvestment,
				CurrentValue: &value,
			}
			acc, err := accountService.UpdateAccount(ctx, createdRegular.Id, 6, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(acc.BankType).To(Equal(models.BankTypeInvestment))
			Expect(acc.CurrentValue).NotTo(BeNil())
			Expect(*acc.CurrentValue).To(Equal(value))
		})
	})

	Describe("DeleteAccount", func() {
		var created models.AccountResponse
		BeforeEach(func() {
			input := models.CreateAccountInput{
				Name:      "Account Delete",
				BankType:  models.BankTypeInvestment,
				Currency:  models.CurrencyINR,
				CreatedBy: 5,
			}
			var err error
			created, err = accountService.CreateAccount(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should delete account by id", func() {
			err := accountService.DeleteAccount(ctx, created.Id, 5)
			Expect(err).NotTo(HaveOccurred())
			_, err = accountService.GetAccountById(ctx, created.Id, 5)
			Expect(err).To(HaveOccurred())
		})
		It("should return error for non-existent id", func() {
			err := accountService.DeleteAccount(ctx, 9999, 5)
			Expect(err).To(HaveOccurred())
		})
		It("should return error for non-existent user id", func() {
			err := accountService.DeleteAccount(ctx, created.Id, 9999)
			Expect(err).To(HaveOccurred())
		})
		It("should return error while deleting account of other user", func() {
			err := accountService.DeleteAccount(ctx, created.Id, 4)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("ListAccounts", func() {
		BeforeEach(func() {
			for i := 0; i < 3; i++ {
				input := models.CreateAccountInput{
					Name:      "Account List " + string(rune('A'+i)),
					BankType:  models.BankTypeAxis,
					Currency:  models.CurrencyINR,
					CreatedBy: 6,
				}
				_, err := accountService.CreateAccount(ctx, input)
				Expect(err).NotTo(HaveOccurred())
			}
		})
		It("should list all accounts", func() {
			accounts, err := accountService.ListAccounts(ctx, 6)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(accounts)).To(BeNumerically(">=", 3))
		})
		It("should not return error for non-existent user id", func() {
			_, err := accountService.ListAccounts(ctx, 9999)
			Expect(err).NotTo(HaveOccurred())
		})
		It("should not return error while listing accounts of other user", func() {
			_, err := accountService.ListAccounts(ctx, 4)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
