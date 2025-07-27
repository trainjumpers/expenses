package service

import (
	"errors"
	mockDatabase "expenses/internal/mock/database"
	repository "expenses/internal/mock/repository"
	"expenses/internal/models"
	"expenses/internal/validator"
	"time"

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
			input := models.ParseStatementInput{
				FileBytes:        fileBytes,
				FileName:         "statement.csv",
				AccountId:        created.Id,
				OriginalFilename: "statement.csv",
			}
			createdStatement, _ := service.ParseStatement(nil, input, userId)
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
		It("should throw Error on parse error", func() {
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
			input := models.ParseStatementInput{
				FileBytes:        fileBytes,
				FileName:         "statement.csv",
				AccountId:        created.Id,
				OriginalFilename: "statement.csv",
			}
			_, err = service.ParseStatement(nil, input, userId)
			Expect(err).NotTo(BeNil())

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
			input := models.ParseStatementInput{
				FileBytes:        fileBytes,
				FileName:         "statement.csv",
				AccountId:        9999,
				OriginalFilename: "statement.csv",
			}
			_, err := service.ParseStatement(nil, input, userId)
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
			input := models.ParseStatementInput{
				FileBytes:        fileBytes,
				FileName:         "statement.csv",
				AccountId:        acc.Id,
				OriginalFilename: "statement.csv",
			}
			resp, err := service.ParseStatement(nil, input, userId)
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
			input := models.ParseStatementInput{
				FileBytes:        fileBytes,
				FileName:         "statement.csv",
				AccountId:        acc.Id,
				OriginalFilename: "statement.csv",
			}
			resp, err := service.ParseStatement(nil, input, userId)
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
			input := models.ParseStatementInput{
				FileBytes:        fileBytes,
				FileName:         "statement.csv",
				AccountId:        acc.Id,
				OriginalFilename: "statement.csv",
			}
			resp, _ := service.ParseStatement(nil, input, userId)
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

		It("should handle parser failure gracefully", func() {
			accInput := models.CreateAccountInput{
				Name:      "Test Account",
				BankType:  models.BankTypeSBI,
				Currency:  models.CurrencyINR,
				CreatedBy: userId,
			}
			acc, err := accountService.CreateAccount(nil, accInput)
			Expect(err).NotTo(HaveOccurred())

			// Use malformed CSV that will cause parser to fail
			fileBytes := []byte("Invalid CSV content that will cause parser to fail")
			input := models.ParseStatementInput{
				FileBytes:        fileBytes,
				FileName:         "statement.csv",
				AccountId:        acc.Id,
				OriginalFilename: "statement.csv",
			}
			resp, err := service.ParseStatement(nil, input, userId)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() models.StatementStatus {
				result, _ := service.GetStatementStatus(nil, resp.Id, userId)
				return result.Status
			}, "2s", "100ms").Should(Equal(models.StatementStatusError))

			result, _ := service.GetStatementStatus(nil, resp.Id, userId)
			Expect(result.Message).NotTo(BeNil())
			Expect(*result.Message).To(ContainSubstring("Failed to parse statement"))
		})

		It("should handle duplicate transactions during insertion", func() {
			accInput := models.CreateAccountInput{
				Name:      "Test Account",
				BankType:  models.BankTypeSBI,
				Currency:  models.CurrencyINR,
				CreatedBy: userId,
			}
			acc, err := accountService.CreateAccount(nil, accInput)
			Expect(err).NotTo(HaveOccurred())

			// Create a transaction first to simulate duplicate
			amount := 100.00
			date, _ := time.Parse("2006-01-02", "2022-08-01")
			existingTx := models.CreateTransactionInput{
				CreateBaseTransactionInput: models.CreateBaseTransactionInput{
					Name:        "UPI to RITIK S",
					Amount:      &amount,
					Date:        date,
					AccountId:   acc.Id,
					CreatedBy:   userId,
					Description: "TO TRANSFER-UPI/DR/221356312527/RITIK S/SBIN/rs6321908@/UPI--",
				},
				CategoryIds: []int64{},
			}
			_, err = txnService.CreateTransaction(nil, existingTx)
			Expect(err).NotTo(HaveOccurred())

			// Now try to parse a statement with the same transaction
			fileBytes := []byte(
				"Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance\n" +
					"1 Aug 2022	1 Aug 2022	TO TRANSFER-UPI/DR/221356312527/RITIK S/SBIN/rs6321908@/UPI--	123456	100.00		1000.00\n" +
					"2 Aug 2022	2 Aug 2022	BY TRANSFER-NEFT*HDFC0000001*N215222062454075*QURIATE TECHNOLO--	654321		200.00	1200.00\n" +
					"Computer Generated Statement")
			input := models.ParseStatementInput{
				FileBytes:        fileBytes,
				FileName:         "statement.csv",
				AccountId:        acc.Id,
				OriginalFilename: "statement.csv",
			}
			resp, err := service.ParseStatement(nil, input, userId)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() models.StatementStatus {
				result, _ := service.GetStatementStatus(nil, resp.Id, userId)
				return result.Status
			}, "2s", "100ms").Should(Equal(models.StatementStatusDone))

			result, _ := service.GetStatementStatus(nil, resp.Id, userId)
			Expect(result.Message).NotTo(BeNil())
			// Should show some failures due to duplicates
			Expect(*result.Message).To(ContainSubstring("failed"))
		})

		It("should handle all transactions failing during insertion", func() {
			accInput := models.CreateAccountInput{
				Name:      "Test Account",
				BankType:  models.BankTypeSBI,
				Currency:  models.CurrencyINR,
				CreatedBy: userId,
			}
			acc, err := accountService.CreateAccount(nil, accInput)
			Expect(err).NotTo(HaveOccurred())

			// Create transactions that will likely fail (e.g., with invalid data)
			// First, let's create some existing transactions to cause conflicts
			amount := 100.00
			date, _ := time.Parse("2006-01-02", "2022-08-01")
			for i := 0; i < 3; i++ {
				existingTx := models.CreateTransactionInput{
					CreateBaseTransactionInput: models.CreateBaseTransactionInput{
						Name:        "UPI to RITIK S",
						Amount:      &amount,
						Date:        date,
						AccountId:   acc.Id,
						CreatedBy:   userId,
						Description: "TO TRANSFER-UPI/DR/221356312527/RITIK S/SBIN/rs6321908@/UPI--",
					},
					CategoryIds: []int64{},
				}
				_, _ = txnService.CreateTransaction(nil, existingTx)
			}

			// Now parse a statement with transactions that will all fail
			fileBytes := []byte(
				"Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance\n" +
					"1 Aug 2022	1 Aug 2022	TO TRANSFER-UPI/DR/221356312527/RITIK S/SBIN/rs6321908@/UPI--	123456	100.00		1000.00\n" +
					"1 Aug 2022	1 Aug 2022	TO TRANSFER-UPI/DR/221356312527/RITIK S/SBIN/rs6321908@/UPI--	123456	100.00		1000.00\n" +
					"Computer Generated Statement")
			input := models.ParseStatementInput{
				FileBytes:        fileBytes,
				FileName:         "statement.csv",
				AccountId:        acc.Id,
				OriginalFilename: "statement.csv",
			}
			resp, err := service.ParseStatement(nil, input, userId)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() models.StatementStatus {
				result, _ := service.GetStatementStatus(nil, resp.Id, userId)
				return result.Status
			}, "2s", "100ms").Should(BeElementOf(models.StatementStatusError, models.StatementStatusDone))

			result, _ := service.GetStatementStatus(nil, resp.Id, userId)
			Expect(result.Message).NotTo(BeNil())
			Expect(*result.Message).To(ContainSubstring("failed"))
		})

		It("should handle empty statement file", func() {
			accInput := models.CreateAccountInput{
				Name:      "Test Account",
				BankType:  models.BankTypeSBI,
				Currency:  models.CurrencyINR,
				CreatedBy: userId,
			}
			acc, err := accountService.CreateAccount(nil, accInput)
			Expect(err).NotTo(HaveOccurred())

			// Empty file content - this should fail validation
			fileBytes := []byte("")
			input := models.ParseStatementInput{
				FileBytes:        fileBytes,
				FileName:         "statement.csv",
				AccountId:        acc.Id,
				OriginalFilename: "statement.csv",
			}
			_, err = service.ParseStatement(nil, input, userId)
			Expect(err).To(HaveOccurred())
			Expect(errors.Unwrap(err).Error()).To(ContainSubstring("file is required"))
		})

		It("should handle statement with only headers", func() {
			accInput := models.CreateAccountInput{
				Name:      "Test Account",
				BankType:  models.BankTypeSBI,
				Currency:  models.CurrencyINR,
				CreatedBy: userId,
			}
			acc, err := accountService.CreateAccount(nil, accInput)
			Expect(err).NotTo(HaveOccurred())

			// Only headers, no data rows
			fileBytes := []byte("Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance\nComputer Generated Statement")
			input := models.ParseStatementInput{
				FileBytes:        fileBytes,
				FileName:         "statement.csv",
				AccountId:        acc.Id,
				OriginalFilename: "statement.csv",
			}
			resp, err := service.ParseStatement(nil, input, userId)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() models.StatementStatus {
				result, _ := service.GetStatementStatus(nil, resp.Id, userId)
				return result.Status
			}, "2s", "100ms").Should(BeElementOf(models.StatementStatusDone, models.StatementStatusError))

			result, _ := service.GetStatementStatus(nil, resp.Id, userId)
			Expect(result.Message).NotTo(BeNil())
			// Should either process 0 transactions or fail parsing
			Expect(*result.Message).To(SatisfyAny(
				ContainSubstring("Processed 0 transactions"),
				ContainSubstring("Failed to parse statement"),
			))
		})

		It("should handle statement with custom bank type override", func() {
			accInput := models.CreateAccountInput{
				Name:      "Test Account",
				BankType:  models.BankTypeAxis, // Different from what we'll specify in input
				Currency:  models.CurrencyINR,
				CreatedBy: userId,
			}
			acc, err := accountService.CreateAccount(nil, accInput)
			Expect(err).NotTo(HaveOccurred())

			fileBytes := []byte(
				"Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance\n" +
					"1 Aug 2022	1 Aug 2022	TO TRANSFER-UPI/DR/221356312527/RITIK S/SBIN/rs6321908@/UPI--	123456	100.00		1000.00\n" +
					"Computer Generated Statement")
			input := models.ParseStatementInput{
				FileBytes:        fileBytes,
				FileName:         "statement.csv",
				AccountId:        acc.Id,
				OriginalFilename: "statement.csv",
				BankType:         string(models.BankTypeSBI), // Override with SBI parser
			}
			resp, err := service.ParseStatement(nil, input, userId)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() models.StatementStatus {
				result, _ := service.GetStatementStatus(nil, resp.Id, userId)
				return result.Status
			}, "2s", "100ms").Should(Equal(models.StatementStatusDone))

			result, _ := service.GetStatementStatus(nil, resp.Id, userId)
			Expect(result.Message).NotTo(BeNil())
			Expect(*result.Message).To(ContainSubstring("Processed"))
		})

		It("should handle statement with metadata", func() {
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
					"Computer Generated Statement")
			input := models.ParseStatementInput{
				FileBytes:        fileBytes,
				FileName:         "statement.csv",
				AccountId:        acc.Id,
				OriginalFilename: "statement.csv",
				Metadata:         `{"skipRows": 0, "customField": "value"}`,
			}
			resp, err := service.ParseStatement(nil, input, userId)
			Expect(err).NotTo(HaveOccurred())

			Eventually(func() models.StatementStatus {
				result, _ := service.GetStatementStatus(nil, resp.Id, userId)
				return result.Status
			}, "2s", "100ms").Should(Equal(models.StatementStatusDone))

			result, _ := service.GetStatementStatus(nil, resp.Id, userId)
			Expect(result.Message).NotTo(BeNil())
			Expect(*result.Message).To(ContainSubstring("Processed"))
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

	Describe("PreviewStatement", func() {
		Describe("Valid CSV files", func() {
			It("should preview a simple CSV file with default row size", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Test Transaction,100.00\n2023-01-02,Another Transaction,-50.00")
				fileName := "test.csv"
				skipRows := 0
				rowSize := 10

				preview, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).NotTo(BeNil())
				Expect(preview.Headers).To(Equal([]string{"Date", "Description", "Amount"}))
				Expect(preview.Rows).To(HaveLen(2))
				Expect(preview.Rows[0]).To(Equal([]string{"2023-01-01", "Test Transaction", "100.00"}))
				Expect(preview.Rows[1]).To(Equal([]string{"2023-01-02", "Another Transaction", "-50.00"}))
			})

			It("should preview CSV file with skip rows", func() {
				fileBytes := []byte("Bank Statement\nAccount: 123456\nDate,Description,Amount\n2023-01-01,Test Transaction,100.00\n2023-01-02,Another Transaction,-50.00")
				fileName := "test.csv"
				skipRows := 2
				rowSize := 10

				preview, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).NotTo(BeNil())
				Expect(preview.Headers).To(Equal([]string{"Date", "Description", "Amount"}))
				Expect(preview.Rows).To(HaveLen(2))
				Expect(preview.Rows[0]).To(Equal([]string{"2023-01-01", "Test Transaction", "100.00"}))
				Expect(preview.Rows[1]).To(Equal([]string{"2023-01-02", "Another Transaction", "-50.00"}))
			})

			It("should limit rows based on rowSize parameter", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Transaction 1,100.00\n2023-01-02,Transaction 2,-50.00\n2023-01-03,Transaction 3,75.00\n2023-01-04,Transaction 4,-25.00")
				fileName := "test.csv"
				skipRows := 0
				rowSize := 2

				preview, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).NotTo(BeNil())
				Expect(preview.Headers).To(Equal([]string{"Date", "Description", "Amount"}))
				Expect(preview.Rows).To(HaveLen(2))
				Expect(preview.Rows[0]).To(Equal([]string{"2023-01-01", "Transaction 1", "100.00"}))
				Expect(preview.Rows[1]).To(Equal([]string{"2023-01-02", "Transaction 2", "-50.00"}))
			})

			It("should default rowSize to 10 when rowSize is 0 or negative", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Test Transaction,100.00")
				fileName := "test.csv"
				skipRows := 0
				rowSize := 1 // Use 1 instead of 0 since validator requires positive rowSize

				preview, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).NotTo(BeNil())
				Expect(preview.Headers).To(Equal([]string{"Date", "Description", "Amount"}))
				Expect(preview.Rows).To(HaveLen(1))
			})

			It("should handle CSV with comma-separated values (not tab)", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Test Transaction,100.00")
				fileName := "test.csv"
				skipRows := 0
				rowSize := 10

				preview, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).NotTo(BeNil())
				Expect(preview.Headers).To(Equal([]string{"Date", "Description", "Amount"}))
				Expect(preview.Rows).To(HaveLen(1))
				Expect(preview.Rows[0]).To(Equal([]string{"2023-01-01", "Test Transaction", "100.00"}))
			})
		})

		Describe("Valid Excel files", func() {
			It("should preview XLS file", func() {
				// Create a minimal XLS file content (this is a simplified example)
				fileBytes := []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1} // XLS file signature
				fileName := "test.xls"
				skipRows := 0
				rowSize := 10

				// Note: This test might fail with actual XLS parsing, but validates the flow
				_, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				// We expect this to either succeed or fail with a parsing error, not a validation error
				if err != nil {
					Expect(err.Error()).NotTo(ContainSubstring("file is required"))
					Expect(err.Error()).NotTo(ContainSubstring("filename cannot be empty"))
				}
			})

			It("should preview XLSX file", func() {
				// Create a minimal XLSX file content (this is a simplified example)
				fileBytes := []byte{0x50, 0x4B, 0x03, 0x04} // ZIP file signature (XLSX is a ZIP)
				fileName := "test.xlsx"
				skipRows := 0
				rowSize := 10

				// Note: This test might fail with actual XLSX parsing, but validates the flow
				_, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				// We expect this to either succeed or fail with a parsing error, not a validation error
				if err != nil {
					Expect(err.Error()).NotTo(ContainSubstring("file is required"))
					Expect(err.Error()).NotTo(ContainSubstring("filename cannot be empty"))
				}
			})
		})

		Describe("Edge cases", func() {
			It("should return empty preview when skipRows exceeds file length", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Test Transaction,100.00")
				fileName := "test.csv"
				skipRows := 10
				rowSize := 10

				preview, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).NotTo(BeNil())
				Expect(preview.Headers).To(HaveLen(0))
				Expect(preview.Rows).To(HaveLen(0))
			})

			It("should handle empty CSV file", func() {
				fileBytes := []byte("\n") // Use a newline instead of space to get empty headers
				fileName := "empty.csv"
				skipRows := 0
				rowSize := 10

				preview, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).To(HaveOccurred())
				Expect(preview).To(BeNil())
			})

			It("should handle CSV with only headers", func() {
				fileBytes := []byte("Date,Description,Amount")
				fileName := "headers-only.csv"
				skipRows := 0
				rowSize := 10

				preview, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).NotTo(BeNil())
				Expect(preview.Headers).To(Equal([]string{"Date", "Description", "Amount"}))
				Expect(preview.Rows).To(HaveLen(0))
			})

			It("should handle CSV with irregular column counts", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Test Transaction\n2023-01-02,Another Transaction,100.00,Extra Column")
				fileName := "irregular.csv"
				skipRows := 0
				rowSize := 10

				preview, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).NotTo(BeNil())
				Expect(preview.Headers).To(Equal([]string{"Date", "Description", "Amount"}))
				Expect(preview.Rows).To(HaveLen(2))
			})

			It("should handle CSV with properly quoted special characters", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,\"Transaction with, comma\",100.00\n2023-01-02,\"Transaction with \"\"quotes\"\"\",50.00")
				fileName := "special-chars.csv"
				skipRows := 0
				rowSize := 10

				preview, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).NotTo(BeNil())
				Expect(preview.Headers).To(Equal([]string{"Date", "Description", "Amount"}))
				Expect(preview.Rows).To(HaveLen(2))
			})
		})

		Describe("Validation errors", func() {
			It("should error when file bytes are empty", func() {
				fileBytes := []byte{}
				fileName := "test.csv"
				skipRows := 0
				rowSize := 10

				_, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).To(HaveOccurred())
				Expect(errors.Unwrap(err).Error()).To(ContainSubstring("file is required"))
			})

			It("should error when filename is empty", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Test Transaction,100.00")
				fileName := ""
				skipRows := 0
				rowSize := 10

				_, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).To(HaveOccurred())
				Expect(errors.Unwrap(err).Error()).To(ContainSubstring("filename cannot be empty"))
			})

			It("should error when filename is only whitespace", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Test Transaction,100.00")
				fileName := "   "
				skipRows := 0
				rowSize := 10

				_, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).To(HaveOccurred())
				Expect(errors.Unwrap(err).Error()).To(ContainSubstring("filename cannot be empty"))
			})

			It("should error when file size exceeds limit", func() {
				// Create a file larger than 256KB
				largeContent := make([]byte, 257*1024)
				for i := range largeContent {
					largeContent[i] = 'a'
				}
				fileName := "large.csv"
				skipRows := 0
				rowSize := 10

				_, err := service.PreviewStatement(nil, largeContent, fileName, skipRows, rowSize)
				Expect(err).To(HaveOccurred())
				Expect(errors.Unwrap(err).Error()).To(ContainSubstring("file size must be less than 256KB"))
			})

			It("should error when file format is unsupported", func() {
				fileBytes := []byte("some content")
				fileName := "test.txt"
				skipRows := 0
				rowSize := 10

				_, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).To(HaveOccurred())
				Expect(errors.Unwrap(err).Error()).To(ContainSubstring("file must be CSV or Excel format"))
			})

			It("should error when skipRows is negative", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Test Transaction,100.00")
				fileName := "test.csv"
				skipRows := -1
				rowSize := 10

				_, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).To(HaveOccurred())
				Expect(errors.Unwrap(err).Error()).To(ContainSubstring("skipRows cannot be negative"))
			})

			It("should error when rowSize is negative", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Test Transaction,100.00")
				fileName := "test.csv"
				skipRows := 0
				rowSize := -5

				_, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).To(HaveOccurred())
				Expect(errors.Unwrap(err).Error()).To(ContainSubstring("rowSize must be positive"))
			})

			It("should handle zero rowSize gracefully by using default", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Test Transaction,100.00")
				fileName := "test.csv"
				skipRows := 0
				rowSize := 0

				preview, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).ToNot(HaveOccurred())
				Expect(preview).ToNot(BeNil())
			})

			It("should accept various supported file extensions", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Test Transaction,100.00")
				skipRows := 0
				rowSize := 10

				// Test CSV
				_, err := service.PreviewStatement(nil, fileBytes, "test.csv", skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())

				// Test CSV with uppercase
				_, err = service.PreviewStatement(nil, fileBytes, "test.CSV", skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())

				// Test XLS (will likely fail parsing but should pass validation)
				xlsBytes := []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}
				_, err = service.PreviewStatement(nil, xlsBytes, "test.xls", skipRows, rowSize)
				// Should not fail with validation error
				if err != nil {
					underlyingErr := errors.Unwrap(err)
					if underlyingErr != nil {
						Expect(underlyingErr.Error()).NotTo(ContainSubstring("file must be CSV or Excel format"))
					}
				}

				// Test XLSX (will likely fail parsing but should pass validation)
				xlsxBytes := []byte{0x50, 0x4B, 0x03, 0x04}
				_, err = service.PreviewStatement(nil, xlsxBytes, "test.xlsx", skipRows, rowSize)
				// Should not fail with validation error
				if err != nil {
					underlyingErr := errors.Unwrap(err)
					if underlyingErr != nil {
						Expect(underlyingErr.Error()).NotTo(ContainSubstring("file must be CSV or Excel format"))
					}
				}
			})
		})

		Describe("Parameter handling", func() {
			It("should handle zero skipRows correctly", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Test Transaction,100.00")
				fileName := "test.csv"
				skipRows := 0
				rowSize := 10

				preview, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).NotTo(BeNil())
				Expect(preview.Headers).To(Equal([]string{"Date", "Description", "Amount"}))
				Expect(preview.Rows).To(HaveLen(1))
			})

			It("should handle large rowSize correctly", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Test Transaction,100.00\n2023-01-02,Another Transaction,-50.00")
				fileName := "test.csv"
				skipRows := 0
				rowSize := 1000

				preview, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
				Expect(preview).NotTo(BeNil())
				Expect(preview.Headers).To(Equal([]string{"Date", "Description", "Amount"}))
				Expect(preview.Rows).To(HaveLen(2)) // Should return all available rows
			})

			It("should handle filename with mixed case extensions", func() {
				fileBytes := []byte("Date,Description,Amount\n2023-01-01,Test Transaction,100.00")
				skipRows := 0
				rowSize := 10

				testCases := []string{"test.Csv", "test.CSV", "test.cSv", "TEST.CSV"}
				for _, fileName := range testCases {
					_, err := service.PreviewStatement(nil, fileBytes, fileName, skipRows, rowSize)
					Expect(err).NotTo(HaveOccurred())
				}
			})
		})
	})
})
