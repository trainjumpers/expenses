package service

import (
	mock "expenses/internal/mock/repository"
	"expenses/internal/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gin-gonic/gin"
)

var _ = Describe("AccountService", func() {
	var (
		accountService AccountServiceInterface
		mockRepo       *mock.MockAccountRepository
		ctx            *gin.Context
	)

	BeforeEach(func() {
		ctx = &gin.Context{}
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
		It("should return error for non-existent user id", func() {
			_, err := accountService.ListAccounts(ctx, 9999)
			Expect(err).To(HaveOccurred())
		})
		It("should return error while listing accounts of other user", func() {
			_, err := accountService.ListAccounts(ctx, 4)
			Expect(err).To(HaveOccurred())
		})
	})
})
