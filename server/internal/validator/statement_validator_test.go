package validator

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StatementValidator", func() {
	var validator *StatementValidator

	BeforeEach(func() {
		validator = NewStatementValidator()
	})

	Describe("ValidateStatementUpload", func() {
		var (
			accountId int64
			fileBytes []byte
			fileName  string
		)

		BeforeEach(func() {
			accountId = 123
			fileBytes = []byte("test,data\n1,2")
			fileName = "test.csv"
		})

		Context("with valid input", func() {
			It("should validate successfully", func() {
				err := validator.ValidateStatementUpload(accountId, fileBytes, fileName)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with invalid account_id", func() {
			It("should return error for zero account_id", func() {
				err := validator.ValidateStatementUpload(0, fileBytes, fileName)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for negative account_id", func() {
				err := validator.ValidateStatementUpload(-1, fileBytes, fileName)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with invalid file", func() {
			It("should return error for nil file bytes", func() {
				err := validator.ValidateStatementUpload(accountId, nil, fileName)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for empty file bytes", func() {
				err := validator.ValidateStatementUpload(accountId, []byte{}, fileName)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for file too large", func() {
				largeFile := make([]byte, 5*1024*1024+1)
				err := validator.ValidateStatementUpload(accountId, largeFile, fileName)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for invalid file type", func() {
				err := validator.ValidateStatementUpload(accountId, fileBytes, "test.png")
				Expect(err).To(HaveOccurred())
			})

			It("should return error for empty filename", func() {
				err := validator.ValidateStatementUpload(accountId, fileBytes, "")
				Expect(err).To(HaveOccurred())
			})

			It("should return error for filename with only spaces", func() {
				err := validator.ValidateStatementUpload(accountId, fileBytes, "   ")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with different valid file types", func() {
			It("should accept .csv files", func() {
				err := validator.ValidateStatementUpload(accountId, fileBytes, "test.csv")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should accept .xls files", func() {
				err := validator.ValidateStatementUpload(accountId, fileBytes, "test.xls")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should accept .xlsx files", func() {
				err := validator.ValidateStatementUpload(accountId, fileBytes, "test.xlsx")
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with edge cases", func() {
			It("should accept file exactly at 5MB limit", func() {
				limitFile := make([]byte, 5*1024*1024)
				err := validator.ValidateStatementUpload(accountId, limitFile, fileName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should accept uppercase file extensions", func() {
				err := validator.ValidateStatementUpload(accountId, fileBytes, "test.CSV")
				Expect(err).NotTo(HaveOccurred())
				err = validator.ValidateStatementUpload(accountId, fileBytes, "test.XLS")
				Expect(err).NotTo(HaveOccurred())
				err = validator.ValidateStatementUpload(accountId, fileBytes, "test.XLSX")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should handle filenames with leading/trailing spaces", func() {
				err := validator.ValidateStatementUpload(accountId, fileBytes, " test.csv ")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should reject file with no extension", func() {
				err := validator.ValidateStatementUpload(accountId, fileBytes, "statement")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("ValidateStatementPreview", func() {
		var (
			fileBytes []byte
			fileName  string
			skipRows  int
			rowSize   int
		)

		BeforeEach(func() {
			fileBytes = []byte("header1,header2\ndata1,data2")
			fileName = "preview.csv"
			skipRows = 0
			rowSize = 10
		})

		Context("with valid input", func() {
			It("should validate successfully", func() {
				err := validator.ValidateStatementPreview(fileBytes, fileName, skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with invalid file", func() {
			It("should return error for nil file bytes", func() {
				err := validator.ValidateStatementPreview(nil, fileName, skipRows, rowSize)
				Expect(err).To(HaveOccurred())
			})
			It("should return error for empty filename", func() {
				err := validator.ValidateStatementPreview(fileBytes, "", skipRows, rowSize)
				Expect(err).To(HaveOccurred())
			})
			It("should return error for file too large", func() {
				largeFile := make([]byte, 5*1024*1024+1)
				err := validator.ValidateStatementPreview(largeFile, fileName, skipRows, rowSize)
				Expect(err).To(HaveOccurred())
			})
			It("should return error for invalid file type (.txt)", func() {
				err := validator.ValidateStatementPreview(fileBytes, "test.txt", skipRows, rowSize)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with invalid parameters", func() {
			It("should return error for negative skipRows", func() {
				err := validator.ValidateStatementPreview(fileBytes, fileName, -1, rowSize)
				Expect(err).To(HaveOccurred())
			})
			It("should return error for zero rowSize", func() {
				err := validator.ValidateStatementPreview(fileBytes, fileName, skipRows, 0)
				Expect(err).To(HaveOccurred())
			})
			It("should return error for negative rowSize", func() {
				err := validator.ValidateStatementPreview(fileBytes, fileName, skipRows, -1)
				Expect(err).To(HaveOccurred())
			})
		})
		Context("with edge cases for filenames", func() {
			It("should accept filename with multiple dots", func() {
				err := validator.ValidateStatementPreview([]byte("data"), "report.final.csv", skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should accept unicode/special characters in filename", func() {
				err := validator.ValidateStatementPreview([]byte("data"), "résumé.csv", skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
				err = validator.ValidateStatementPreview([]byte("data"), "data_测试.xlsx", skipRows, rowSize)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should reject filename with only extension", func() {
				err := validator.ValidateStatementPreview([]byte("data"), ".csv", skipRows, rowSize)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with edge cases for file content", func() {
			It("should reject file with whitespace-only content", func() {
				err := validator.ValidateStatementPreview([]byte("   "), "test.csv", skipRows, rowSize)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with edge cases for preview parameters", func() {
			It("should accept large rowSize if allowed", func() {
				err := validator.ValidateStatementPreview([]byte("header1,header2\ndata1,data2"), "preview.csv", 0, 1000)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should accept skipRows exceeding file line count", func() {
				err := validator.ValidateStatementPreview([]byte("header1,header2\ndata1,data2"), "preview.csv", 100, 10)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
})
