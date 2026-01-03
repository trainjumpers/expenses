package parser

import (
	"expenses/internal/models"
	"expenses/pkg/utils"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSBIParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Statement Parser Suite")
}

var _ = Describe("SBIParser", func() {
	var parser *SBIParser

	BeforeEach(func() {
		parser = &SBIParser{}
	})

	Describe("parseDate", func() {
		It("should parse SBI date formats correctly", func() {
			testCases := []struct {
				input    string
				expected time.Time
			}{
				{"01/12/2024", time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC)},
				{"15/01/2024", time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
				{"1 Aug 2022", time.Date(2022, 8, 1, 0, 0, 0, 0, time.UTC)},
				{"01 Aug 2022", time.Date(2022, 8, 1, 0, 0, 0, 0, time.UTC)},
				{"2 January 2022", time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)},
			}

			for _, tc := range testCases {
				result, err := utils.ParseDate(tc.input)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(tc.expected))
			}
		})

		It("should error on unsupported date format", func() {
			_, err := utils.ParseDate("invalid-date-format")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("parseAmount", func() {
		It("should parse amounts with commas correctly", func() {
			testCases := []struct {
				input    string
				expected float64
			}{
				{"40.00", 40.0},
				{"1,141.74", 1141.74},
				{"50,000.00", 50000.0},
				{"2,59,000.00", 259000.0},
				{"  2,000.50  ", 2000.50},
			}

			for _, tc := range testCases {
				result, err := utils.ParseFloat(tc.input)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(tc.expected))
			}
		})

		It("should error on empty string", func() {
			_, err := utils.ParseFloat("")
			Expect(err).To(HaveOccurred())
		})

		It("should error on spaces only", func() {
			_, err := utils.ParseFloat("   ")
			Expect(err).To(HaveOccurred())
		})

		It("should error on invalid number format", func() {
			_, err := utils.ParseFloat("abc123")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("parseTransactionRow", func() {
		It("should error on insufficient columns", func() {
			fields := []string{"01/12/2024", "Desc", "123", "100.00"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should error on invalid date", func() {
			fields := []string{"invalid-date", "Desc", "123", "100.00", "", "1000.00"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should error on invalid amount", func() {
			fields := []string{"01/12/2024", "Desc", "123", "abc", "", "1000.00"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should error when both debit and credit are empty", func() {
			fields := []string{"01/12/2024", "Desc", "123", "", "", "1000.00"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should parse valid debit transaction row", func() {
			fields := []string{"01/12/2024", "WDL TFR UPI/DR/123456789/JOHN DOE/SBIN/user123@/UPI--", "REF123", "100.00", "", "1000.00"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Name).To(Equal("UPI to JOHN DOE"))
			Expect(*result.Amount).To(Equal(100.00))
			Expect(result.Description).To(ContainSubstring("Ref: REF123"))
			Expect(result.CategoryIds).To(BeEmpty())
		})

		It("should parse valid credit transaction row", func() {
			fields := []string{"02/12/2024", "DEP TFR NEFT*HDFC0000001*N123456789*COMPANY NAME--", "REF456", "", "200.00", "1200.00"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Name).To(Equal("NEFT from HDFC0000001"))
			Expect(*result.Amount).To(Equal(-200.00))
			Expect(result.Description).To(ContainSubstring("Ref: REF456"))
			Expect(result.CategoryIds).To(BeEmpty())
		})

		It("should parse row with both debit and credit present (prioritizes debit)", func() {
			fields := []string{"03/12/2024", "Desc", "REF789", "150.00", "200.00", "1300.00"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(*result.Amount).To(Equal(150.00))
		})

		It("should parse negative amount in debit", func() {
			fields := []string{"04/12/2024", "Desc", "REF345", "-50.00", "", "1250.00"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(*result.Amount).To(Equal(-50.00))
		})

		It("should parse negative amount in credit", func() {
			fields := []string{"05/12/2024", "Desc", "REF987", "", "-75.00", "1175.00"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(*result.Amount).To(Equal(75.00))
		})
	})

	Describe("Parse", func() {
		It("should parse a valid SBI statement with multiple transactions", func() {
			data := [][]string{
				{"Date", "Details", "Ref No/Cheque No", "Debit", "Credit", "Balance"},
				{"01/12/2024", "WDL TFR UPI/DR/123456789/JOHN DOE/SBIN/user123@/UPI", "REF123", "100.00", "", "1000.00"},
				{"02/12/2024", "DEP TFR NEFT*HDFC0000001*N123456789*COMPANY NAME", "REF456", "", "200.00", "1200.00"},
			}
			fileBytes := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(fileBytes, "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(2))
			Expect(txns[0].Name).To(Equal("UPI to JOHN DOE"))
			Expect(txns[1].Name).To(Equal("NEFT from HDFC0000001"))
			Expect(txns[0].CategoryIds).To(BeEmpty())
			Expect(txns[1].CategoryIds).To(BeEmpty())
		})

		It("should parse row with both debit and credit present (prioritizes debit)", func() {
			data := [][]string{
				{"Date", "Details", "Ref No/Cheque No", "Debit", "Credit", "Balance"},
				{"03/12/2024", "Desc", "REF789", "150.00", "200.00", "1300.00"},
			}
			fileBytes := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(fileBytes, "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
			Expect(*txns[0].Amount).To(Equal(150.00))
		})

		It("should parse row with extra columns", func() {
			data := [][]string{
				{"Date", "Details", "Ref No/Cheque No", "Debit", "Credit", "Balance", "Extra"},
				{"04/12/2024", "Desc", "REF345", "100.00", "", "1250.00", "ExtraValue"},
			}
			fileBytes := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(fileBytes, "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
		})

		It("should parse row with spaces in description", func() {
			data := [][]string{
				{"Date", "Details", "Ref No/Cheque No", "Debit", "Credit", "Balance"},
				{"05/12/2024", "Desc with spaces", "REF987", "100.00", "", "1175.00"},
			}
			fileBytes := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(fileBytes, "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
		})

		It("should parse header row with extra whitespace", func() {
			data := [][]string{
				{"  Date  ", "Details", "Ref No/Cheque No", "Debit", "Credit", "Balance"},
				{"07/12/2024", "नमस्ते", "123456", "100.00", "", "1000.00"},
			}
			fileBytes := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(fileBytes, "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
		})

		It("should parse row with non-ASCII characters in description", func() {
			data := [][]string{
				{"Date", "Details", "Ref No/Cheque No", "Debit", "Credit", "Balance"},
				{"07/12/2024", "नमस्ते", "REF123", "100.00", "", "1000.00"},
			}
			fileBytes := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(fileBytes, "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
			Expect(txns[0].Name).To(ContainSubstring("नमस्ते"))
		})

		It("should error if header row is missing", func() {
			data := [][]string{
				{"No header here"},
				{"Just some text"},
			}
			fileBytes := utils.CreateXLSXFile(data)
			_, err := parser.Parse(fileBytes, "", "")
			Expect(err).To(MatchError(ContainSubstring("transaction header row not found")))
		})

		It("should handle rows with and without RefNo", func() {
			data := [][]string{
				{"Date", "Details", "Ref No/Cheque No", "Debit", "Credit", "Balance"},
				{"01/12/2024", "Desc", "", "100.00", "", "1000.00"},
				{"02/12/2024", "Desc", "REF654", "", "200.00", "1200.00"},
			}
			fileBytes := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(fileBytes, "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(2))
			Expect(txns[0].Description).To(Equal("Desc"))
			Expect(txns[1].Description).To(Equal("Desc (Ref: REF654)"))
		})

		It("should ensure CategoryIds is always empty", func() {
			data := [][]string{
				{"Date", "Details", "Ref No/Cheque No", "Debit", "Credit", "Balance"},
				{"01/12/2024", "Desc", "REF123", "100.00", "", "1000.00"},
			}
			fileBytes := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(fileBytes, "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
			Expect(txns[0].CategoryIds).To(BeEmpty())
		})

		It("should skip lines with insufficient columns", func() {
			data := [][]string{
				{"Date", "Details", "Ref No/Cheque No", "Debit", "Credit", "Balance"},
				{"01/12/2024", "Short", "REF123", "100.00"},
				{"02/12/2024", "Desc", "REF456", "", "200.00", "1200.00"},
			}
			fileBytes := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(fileBytes, "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
		})

		It("should skip lines with both debit and credit empty", func() {
			data := [][]string{
				{"Date", "Details", "Ref No/Cheque No", "Debit", "Credit", "Balance"},
				{"01/12/2024", "Desc", "123456", "abc", "", "1000.00"},
				{"02/12/2024", "Desc", "123456", "", "200.00", "1200.00"},
			}
			fileBytes := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(fileBytes, "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
		})

		It("should skip lines with malformed amount", func() {
			data := [][]string{
				{"Date", "Details", "Ref No/Cheque No", "Debit", "Credit", "Balance"},
				{"01/12/2024", "Desc", "REF123", "abc", "", "1000.00"},
				{"02/12/2024", "Desc", "REF456", "", "200.00", "1200.00"},
			}
			fileBytes := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(fileBytes, "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
		})
	})

	Describe("generateTransactionName", func() {
		It("should generate readable names from descriptions", func() {
			testCases := []struct {
				description string
				isCredit    bool
				expected    string
			}{
				{
					"WDL TFR UPI/DR/123456789/JOHN DOE/SBIN/user123@/UPI",
					false,
					"UPI to JOHN DOE",
				},
				{
					"DEP TFR UPI/CR/123456789/JANE DOE/SBIN/user456@/UPI",
					true,
					"UPI from JANE DOE",
				},
				{
					"DEP TFR NEFT*HDFC0000001*N123456789*COMPANY NAME",
					true,
					"NEFT from HDFC0000001",
				},
				{
					"WDL TFR NEFT*ICIC0000002*N987654321*ANOTHER COMPANY",
					false,
					"NEFT to ICIC0000002",
				},
				{
					"DEBIT-ATMCard AMC  607431 CLASSIC",
					false,
					"ATM Card AMC 607431 (Debit)",
				},
				{
					"CREDIT-ATMCard AMC  607431 CLASSIC",
					true,
					"ATM Card AMC 607431 (Credit)",
				},
				{
					"DEBIT-e-TDR/e-STDR 123456789",
					false,
					"Term Deposit 123456789 (Debit)",
				},
				{
					"CREDIT-e-TDR/e-STDR 123456789",
					true,
					"Term Deposit 123456789 (Credit)",
				},
				{
					"WDL TFR This is a very long description that should be truncated for readability",
					false,
					"Debit: This is a very...",
				},
				{
					"DEP TFR Short Desc",
					true,
					"Credit: Short Desc",
				},
			}

			for _, tc := range testCases {
				result := parser.generateTransactionName(tc.description, tc.isCredit)
				Expect(result).To(Equal(tc.expected))
			}
		})
	})

	Describe("Parser Registry", func() {
		It("should return SBI parser for BankTypeSBI", func() {
			p, ok := GetParser(models.BankTypeSBI)
			Expect(ok).To(BeTrue())
			Expect(p).NotTo(BeNil())
		})

		It("should return false for unknown bank type", func() {
			p, ok := GetParser("UNKNOWN_BANK")
			Expect(ok).To(BeFalse())
			Expect(p).To(BeNil())
		})
	})
})
