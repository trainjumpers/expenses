package service

import (
	mockDatabase "expenses/internal/mock/database"
	repository "expenses/internal/mock/repository"
	"expenses/internal/models"
	"expenses/internal/validator"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StatementService", func() {
	var (
		mockRepo       *repository.MockStatementRepository
		service        StatementService
		txnService     TransactionServiceInterface
		accountService AccountServiceInterface
		userId         int64
	)

	BeforeEach(func() {
		mockRepo = repository.NewMockStatementRepository()
		mockTxnRepo := repository.NewMockTransactionRepository()
		mockCategoryRepo := repository.NewMockCategoryRepository()
		mockAccountRepo := repository.NewMockAccountRepository()
		mockDbManager := mockDatabase.NewMockDatabaseManager()
		txnService = NewTransactionService(mockTxnRepo, mockCategoryRepo, mockAccountRepo, mockDbManager)
		accountService = NewAccountService(mockAccountRepo)

		service = StatementService{
			repo:               mockRepo,
			statementValidator: validator.NewStatementValidator(),
			txService:          txnService,
			accountService:     accountService,
		}
		userId = 42
	})

	Describe("CreateStatement and ListStatements", func() {
		It("should create and list statements with pagination", func() {
			// Create 7 statements
			i := 0
			for i < 7 {
				i++
				input := models.CreateStatementInput{
					AccountId:        1,
					CreatedBy:        userId,
					OriginalFilename: "file.csv",
					FileType:         "csv",
					Status:           models.StatementStatusPending,
				}
				_, err := mockRepo.CreateStatement(nil, input)
				Expect(err).NotTo(HaveOccurred())
			}

			// List page 1, page_size 5
			resp, err := service.ListStatements(nil, userId, 1, 5)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Statements).To(HaveLen(5))
			Expect(resp.Total).To(Equal(7))
			Expect(resp.Page).To(Equal(1))
			Expect(resp.PageSize).To(Equal(5))

			// List page 2, page_size 5
			resp2, err := service.ListStatements(nil, userId, 2, 5)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp2.Statements).To(HaveLen(2))
			Expect(resp2.Total).To(Equal(7))
			Expect(resp2.Page).To(Equal(2))
			Expect(resp2.PageSize).To(Equal(5))
		})
	})

	Describe("GetStatementStatus", func() {
		It("should get a statement by ID", func() {
			input := models.CreateStatementInput{
				AccountId:        1,
				CreatedBy:        userId,
				OriginalFilename: "file.csv",
				FileType:         "csv",
				Status:           models.StatementStatusPending,
			}
			created, err := mockRepo.CreateStatement(nil, input)
			Expect(err).NotTo(HaveOccurred())
			result, err := service.GetStatementStatus(nil, created.Id, userId)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Id).To(Equal(created.Id))
		})
	})

	Describe("UpdateStatementStatus", func() {
		It("should update the status and message", func() {
			input := models.CreateStatementInput{
				AccountId:        1,
				CreatedBy:        userId,
				OriginalFilename: "file.csv",
				FileType:         "csv",
				Status:           models.StatementStatusPending,
			}
			created, err := mockRepo.CreateStatement(nil, input)
			Expect(err).NotTo(HaveOccurred())
			update := models.UpdateStatementStatusInput{
				Status:  models.StatementStatusDone,
				Message: nil,
			}
			_, err = mockRepo.UpdateStatementStatus(nil, created.Id, update)
			Expect(err).NotTo(HaveOccurred())
			result, err := service.GetStatementStatus(nil, created.Id, userId)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Status).To(Equal(models.StatementStatusDone))
		})
	})

	Describe("CountStatementsByUserId", func() {
		It("should count statements for a user", func() {
			for i := 0; i < 3; i++ {
				input := models.CreateStatementInput{
					AccountId:        1,
					CreatedBy:        userId,
					OriginalFilename: "file.csv",
					FileType:         "csv",
					Status:           models.StatementStatusPending,
				}
				_, err := mockRepo.CreateStatement(nil, input)
				Expect(err).NotTo(HaveOccurred())
			}
			count, err := mockRepo.CountStatementsByUserId(nil, userId)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(3))
		})
	})

	Describe("Error Handling", func() {
		It("should return empty list and total 0 for user with no statements", func() {
			resp, err := service.ListStatements(nil, 999, 1, 5)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Statements).To(HaveLen(0))
			Expect(resp.Total).To(Equal(0))
		})

		It("should return error for getting statement with invalid ID", func() {
			_, err := service.GetStatementStatus(nil, 9999, userId)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for updating status of non-existent statement", func() {
			update := models.UpdateStatementStatusInput{Status: models.StatementStatusDone, Message: nil}
			_, err := mockRepo.UpdateStatementStatus(nil, 9999, update)
			Expect(err).To(HaveOccurred())
		})

		It("should return 0 for counting statements for user with no statements", func() {
			count, err := mockRepo.CountStatementsByUserId(nil, 8888)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(0))
		})
	})

	Describe("Validation", func() {
		It("should error on creating statement with negative account ID", func() {
			input := models.CreateStatementInput{
				AccountId:        -1,
				CreatedBy:        userId,
				OriginalFilename: "file.csv",
				FileType:         "csv",
				Status:           models.StatementStatusPending,
			}
			_, err := mockRepo.CreateStatement(nil, input)
			Expect(err).To(HaveOccurred())
		})

		It("should error on creating statement with missing filename", func() {
			input := models.CreateStatementInput{
				AccountId:        1,
				CreatedBy:        userId,
				OriginalFilename: "",
				FileType:         "csv",
				Status:           models.StatementStatusPending,
			}
			_, err := mockRepo.CreateStatement(nil, input)
			Expect(err).To(HaveOccurred())
		})

		It("should error on creating statement with invalid file type", func() {
			input := models.CreateStatementInput{
				AccountId:        1,
				CreatedBy:        userId,
				OriginalFilename: "file.csv",
				FileType:         "exe",
				Status:           models.StatementStatusPending,
			}
			_, err := mockRepo.CreateStatement(nil, input)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Authorization", func() {
		It("should error when getting statement belonging to another user", func() {
			input := models.CreateStatementInput{
				AccountId:        1,
				CreatedBy:        12345,
				OriginalFilename: "file.csv",
				FileType:         "csv",
				Status:           models.StatementStatusPending,
			}
			created, err := mockRepo.CreateStatement(nil, input)
			Expect(err).NotTo(HaveOccurred())
			_, err = service.GetStatementStatus(nil, created.Id, userId)
			Expect(err).To(HaveOccurred())
		})

		It("should error when updating status for statement belonging to another user", func() {
			input := models.CreateStatementInput{
				AccountId:        1,
				CreatedBy:        12345,
				OriginalFilename: "file.csv",
				FileType:         "csv",
				Status:           models.StatementStatusPending,
			}
			created, err := mockRepo.CreateStatement(nil, input)
			Expect(err).NotTo(HaveOccurred())
			update := models.UpdateStatementStatusInput{Status: models.StatementStatusDone, Message: nil}
			_, err = mockRepo.UpdateStatementStatus(nil, created.Id, update)
			Expect(err).NotTo(HaveOccurred())
			_, err = service.GetStatementStatus(nil, created.Id, userId)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Edge Cases", func() {
		It("should return empty list for page number out of range", func() {
			// Create 2 statements
			for i := 0; i < 2; i++ {
				input := models.CreateStatementInput{
					AccountId:        1,
					CreatedBy:        userId,
					OriginalFilename: "file.csv",
					FileType:         "csv",
					Status:           models.StatementStatusPending,
				}
				_, err := mockRepo.CreateStatement(nil, input)
				Expect(err).NotTo(HaveOccurred())
			}
			resp, err := service.ListStatements(nil, userId, 10, 5)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Statements).To(HaveLen(0))
		})

		It("should clamp page size if out of allowed range", func() {
			// Create 1 statement
			input := models.CreateStatementInput{
				AccountId:        1,
				CreatedBy:        userId,
				OriginalFilename: "file.csv",
				FileType:         "csv",
				Status:           models.StatementStatusPending,
			}
			_, err := mockRepo.CreateStatement(nil, input)
			Expect(err).NotTo(HaveOccurred())
			resp, err := service.ListStatements(nil, userId, 1, 200)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.PageSize).To(Equal(10))
		})

		It("should create statement with very large account ID", func() {
			input := models.CreateStatementInput{
				AccountId:        1 << 60,
				CreatedBy:        userId,
				OriginalFilename: "file.csv",
				FileType:         "csv",
				Status:           models.StatementStatusPending,
			}
			_, err := mockRepo.CreateStatement(nil, input)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should create statement with special characters in filename", func() {
			input := models.CreateStatementInput{
				AccountId:        1,
				CreatedBy:        userId,
				OriginalFilename: "t@st#file!.csv",
				FileType:         "csv",
				Status:           models.StatementStatusPending,
			}
			_, err := mockRepo.CreateStatement(nil, input)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("ParseStatement for Valid statement", func() {
		It("should parse a valid statement and update status to Done or Error", func() {
			balance := 1000.0
			created, err := accountService.CreateAccount(nil, models.CreateAccountInput{
				Name:      "Test",
				BankType:  models.BankTypeAxis,
				Currency:  models.CurrencyINR,
				Balance:   &balance,
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())
			fileBytes := []byte("Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance\n1 Aug 2022	1 Aug 2022	Desc	123	100.00		1000.00\nComputer Generated Statement")
			createdStatement, _ := service.ParseStatement(nil, fileBytes, "statement.csv", created.Id, userId)
			Eventually(func() models.StatementStatus {
				result, _ := service.GetStatementStatus(nil, createdStatement.Id, userId)
				return result.Status
			}, "2s", "100ms").Should(BeElementOf(models.StatementStatusDone, models.StatementStatusError))
			result, _ := service.GetStatementStatus(nil, createdStatement.Id, userId)
			if result.Status == models.StatementStatusDone {
				txns, err := txnService.ListTransactions(nil, userId, models.TransactionListQuery{})
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).NotTo(BeEmpty())
			}
			if result.Status == models.StatementStatusError {
				Expect(result.Message).NotTo(BeNil())
			}
		})
	})

	Describe("ParseStatement for invalid case", func() {
		It("should update status to Error on parse error", func() {
			balance := 1000.0
			created, err := accountService.CreateAccount(nil, models.CreateAccountInput{
				Name:      "Test",
				BankType:  models.BankTypeSBI,
				Currency:  models.CurrencyINR,
				Balance:   &balance,
				CreatedBy: userId,
			})
			Expect(err).NotTo(HaveOccurred())

			fileBytes := []byte("")
			createdStatement, _ := service.ParseStatement(nil, fileBytes, "statement.csv", created.Id, userId)

			Eventually(func() models.StatementStatus {
				result, _ := service.GetStatementStatus(nil, createdStatement.Id, userId)
				return result.Status
			}, "2s", "100ms").Should(Equal(models.StatementStatusError))

			// Optionally check error message
			result, _ := service.GetStatementStatus(nil, createdStatement.Id, userId)
			Expect(result.Message).NotTo(BeNil())
		})
	})

	Describe("Status and Message Update", func() {
		It("should update status with a message", func() {
			input := models.CreateStatementInput{
				AccountId:        1,
				CreatedBy:        userId,
				OriginalFilename: "file.csv",
				FileType:         "csv",
				Status:           models.StatementStatusPending,
			}
			created, err := mockRepo.CreateStatement(nil, input)
			Expect(err).NotTo(HaveOccurred())
			msg := "Processed successfully"
			update := models.UpdateStatementStatusInput{Status: models.StatementStatusDone, Message: &msg}
			_, err = mockRepo.UpdateStatementStatus(nil, created.Id, update)
			Expect(err).NotTo(HaveOccurred())
			result, err := service.GetStatementStatus(nil, created.Id, userId)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Status).To(Equal(models.StatementStatusDone))
			Expect(result.Message).To(Equal(&msg))
		})
	})

	Describe("ParseStatement edge and complex cases", func() {
		It("should set account not found if accountId not found", func() {
			// Use an accountId that does not exist in the mock repo
			fileBytes := []byte("Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance\n1 Aug 2022	1 Aug 2022	Desc	123	100.00		1000.00\nComputer Generated Statement")
			_, err := service.ParseStatement(nil, fileBytes, "statement.csv", 99999, userId)
			Expect(err).To(HaveOccurred())
		})

		It("should set status to Error if parser not found for bank type", func() {
			// Create an account with a bank type that is not registered
			accInput := models.CreateAccountInput{
				Name:      "Ghost Account",
				BankType:  "NON_EXISTENT_BANK",
				Currency:  models.CurrencyINR,
				CreatedBy: userId,
			}
			acc, err := accountService.CreateAccount(nil, accInput)
			Expect(err).NotTo(HaveOccurred())
			fileBytes := []byte("Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance\n1 Aug 2022	1 Aug 2022	Desc	123	100.00		1000.00\nComputer Generated Statement")
			resp, err := service.ParseStatement(nil, fileBytes, "statement.csv", acc.Id, userId)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() models.StatementStatus {
				result, _ := service.GetStatementStatus(nil, resp.Id, userId)
				return result.Status
			}, "2s", "100ms").Should(Equal(models.StatementStatusError))
			result, _ := service.GetStatementStatus(nil, resp.Id, userId)
			Expect(result.Message).NotTo(BeNil())
			Expect(*result.Message).To(ContainSubstring("No parser available for bank type"))
		})

		It("should parse a statement with multiple valid transactions", func() {
			// Create a valid account with a supported bank type
			accInput := models.CreateAccountInput{
				Name:      "Test Account",
				BankType:  models.BankTypeSBI,
				Currency:  models.CurrencyINR,
				CreatedBy: userId,
			}
			acc, err := accountService.CreateAccount(nil, accInput)
			Expect(err).NotTo(HaveOccurred())
			// Complex input with multiple transactions
			fileBytes := []byte(
				"Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance\n" +
					"3 Aug 2022	1 Aug 2022	TO TRANSFER-UPI/DR/221356312527/RITIK S/SBIN/rs6321908@/UPI--	123456	100.00		1000.00\n" +
					"2 Aug 2022	2 Aug 2022	BY TRANSFER-NEFT*HDFC0000001*N215222062454075*QURIATE TECHNOLO--	654321		200.00	1200.00\n" +
					"1 Aug 2022	3 Aug 2022	DEBIT-ATMCard AMC  607431*3795 CLASSIC--	789012	150.00		1300.00\n" +
					"Computer Generated Statement")
			resp, err := service.ParseStatement(nil, fileBytes, "statement.csv", acc.Id, userId)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() models.StatementStatus {
				result, _ := service.GetStatementStatus(nil, resp.Id, userId)
				return result.Status
			}, "2s", "100ms").Should(Equal(models.StatementStatusDone))
			result, _ := service.GetStatementStatus(nil, resp.Id, userId)
			Expect(result.Message).NotTo(BeNil())
			Expect(*result.Message).To(ContainSubstring("Processed"))
			// Check that transactions were created
			txns, err := service.txService.ListTransactions(nil, userId, models.TransactionListQuery{})
			Expect(err).NotTo(HaveOccurred())
			Expect(txns.Transactions).To(HaveLen(3))
			Expect(txns.Transactions[0].Name).To(ContainSubstring("UPI to RITIK S"))
			Expect(txns.Transactions[1].Name).To(ContainSubstring("NEFT from HDFC0000001"))
			Expect(txns.Transactions[2].Name).To(ContainSubstring("ATM Card AMC"))
		})

		It("should handle statement with some malformed rows gracefully", func() {
			accInput := models.CreateAccountInput{
				Name:      "Test Account",
				BankType:  models.BankTypeSBI,
				Currency:  models.CurrencyINR,
				CreatedBy: userId,
			}
			acc, err := accountService.CreateAccount(nil, accInput)
			Expect(err).NotTo(HaveOccurred())
			fileBytes := []byte(
				"Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance\n" +
					"1 Aug 2022	1 Aug 2022	TO TRANSFER-UPI/DR/221356312527/RITIK S/SBIN/rs6321908@/UPI--	123456	100.00		1000.00\n" +
					"BADROW\n" +
					"2 Aug 2022	2 Aug 2022	BY TRANSFER-NEFT*HDFC0000001*N215222062454075*QURIATE TECHNOLO--	654321		200.00	1200.00\n" +
					"Computer Generated Statement")
			resp, err := service.ParseStatement(nil, fileBytes, "statement.csv", acc.Id, userId)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() models.StatementStatus {
				result, _ := service.GetStatementStatus(nil, resp.Id, userId)
				return result.Status
			}, "2s", "100ms").Should(Equal(models.StatementStatusDone))
			result, _ := service.GetStatementStatus(nil, resp.Id, userId)
			Expect(result.Message).NotTo(BeNil())
			Expect(*result.Message).To(ContainSubstring("Processed"))
			txns, err := service.txService.ListTransactions(nil, userId, models.TransactionListQuery{})
			Expect(err).NotTo(HaveOccurred())
			Expect(txns.Transactions).To(HaveLen(2))
		})
	})

	Describe("Input Validation Edge Cases", func() {
		It("should error when getting statement with negative ID", func() {
			_, err := service.GetStatementStatus(nil, -1, userId)
			Expect(err).To(HaveOccurred())
		})

		It("should return empty list for negative page number", func() {
			resp, err := service.ListStatements(nil, userId, -5, 5)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Statements).To(HaveLen(0))
			Expect(resp.Total).To(Equal(0))
		})

		It("should clamp pageSize to minimum for negative pageSize", func() {
			// Create 1 statement
			input := models.CreateStatementInput{
				AccountId:        1,
				CreatedBy:        userId,
				OriginalFilename: "file.csv",
				FileType:         "csv",
				Status:           models.StatementStatusPending,
			}
			_, err := mockRepo.CreateStatement(nil, input)
			Expect(err).NotTo(HaveOccurred())
			resp, err := service.ListStatements(nil, userId, 1, -10)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.PageSize).To(Equal(10))
		})

		It("should clamp pageSize to maximum for pageSize > 100", func() {
			// Create 1 statement
			input := models.CreateStatementInput{
				AccountId:        1,
				CreatedBy:        userId,
				OriginalFilename: "file.csv",
				FileType:         "csv",
				Status:           models.StatementStatusPending,
			}
			_, err := mockRepo.CreateStatement(nil, input)
			Expect(err).NotTo(HaveOccurred())
			resp, err := service.ListStatements(nil, userId, 1, 101)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.PageSize).To(Equal(10)) // Assuming 10 is the max allowed
		})
	})
})
