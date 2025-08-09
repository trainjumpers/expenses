package parser

import (
	"expenses/internal/models"
	"expenses/pkg/utils"
	"os"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ICICICreditParser", func() {
	var parser *ICICICreditParser

	BeforeEach(func() {
		parser = &ICICICreditParser{}
	})

	Describe("parseDate", func() {
		It("should parse DD/MM/YYYY format correctly", func() {
			expectedTime := time.Date(2025, 7, 12, 0, 0, 0, 0, time.UTC)
			result, err := utils.ParseDate("12/07/2025")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(expectedTime))
		})

		It("should error on unsupported date format", func() {
			_, err := utils.ParseDate("invalid-date-format")
			Expect(err).To(HaveOccurred())
		})

		It("should error on invalid date", func() {
			_, err := utils.ParseDate("32/01/2025")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("parseAmount", func() {
		It("should parse amounts correctly", func() {
			result, err := utils.ParseFloat("612.16")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(612.16))
		})

		It("should parse amounts with commas correctly", func() {
			result, err := utils.ParseFloat("10,412.40")
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(10412.40))
		})

		It("should error on empty string", func() {
			_, err := utils.ParseFloat("")
			Expect(err).To(HaveOccurred())
		})

		It("should error on invalid number format", func() {
			_, err := utils.ParseFloat("invalid-amount")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("generateTransactionName", func() {
		It("should return the trimmed description", func() {
			description := "  BOOK MY SHOW PAYU PG MUMBAI IN  "
			expected := "BOOK MY SHOW PAYU PG MUMBAI IN"
			result := parser.generateTransactionName(description)
			Expect(result).To(Equal(expected))
		})
	})

	Describe("parseTransactionRow", func() {
		It("should parse a valid debit transaction row", func() {
			fields := []string{"12/07/2025", "11600218774", "BOOK MY SHOW PAYU PG MUMBAI IN", "12", "0", "612.16", ""}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Name).To(Equal("BOOK MY SHOW PAYU PG MUMBAI IN"))
			Expect(*result.Amount).To(Equal(612.16))
			Expect(result.Date).To(Equal(time.Date(2025, 7, 12, 0, 0, 0, 0, time.UTC)))
			Expect(result.CategoryIds).To(BeEmpty())
		})

		It("should parse a valid credit transaction row", func() {
			fields := []string{"08/07/2025", "11571085308", "BBPS Payment received", "0", "0", "10412.40", "CR"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(result.Name).To(Equal("BBPS Payment received"))
			Expect(*result.Amount).To(Equal(-10412.40))
			Expect(result.Date).To(Equal(time.Date(2025, 7, 8, 0, 0, 0, 0, time.UTC)))
		})

		It("should return nil for rows with insufficient columns", func() {
			fields := []string{"3747XXXXXXXX4009"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return nil for header rows", func() {
			fields := []string{"Date", "Sr.No.", "Transaction Details", "Reward Point Header", "Intl.Amount", "Amount(in Rs)", "BillingAmountSign"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return nil for rows with unparseable dates", func() {
			fields := []string{"invalid-date", "11600218774", "Description", "12", "0", "612.16", ""}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return an error for rows with unparseable amounts", func() {
			fields := []string{"12/07/2025", "11600218774", "Description", "12", "0", "bad-amount", ""}
			_, err := parser.parseTransactionRow(fields)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Parse", func() {
		It("should parse a valid ICICI credit statement file", func() {
			// Using content from the provided CreditCardStatement.CSV
			filePath := filepath.Join("../../../", "CreditCardStatement.CSV")
			fileBytes, err := os.ReadFile(filePath)
			Expect(err).NotTo(HaveOccurred())

			txns, err := parser.Parse(fileBytes, "", "CreditCardStatement.CSV")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(44)) // 44 transactions in the first file

			// Check a few transactions to ensure correctness
			// Debit
			Expect(txns[0].Name).To(Equal("BOOK MY SHOW PAYU PG MUMBAI IN"))
			Expect(*txns[0].Amount).To(Equal(612.16))
			Expect(txns[0].Date).To(Equal(time.Date(2025, 7, 12, 0, 0, 0, 0, time.UTC)))

			// Credit
			Expect(txns[9].Name).To(Equal("BBPS Payment received"))
			Expect(*txns[9].Amount).To(Equal(-10412.40))
			Expect(txns[9].Date).To(Equal(time.Date(2025, 7, 8, 0, 0, 0, 0, time.UTC)))

			// Another debit from a different card on the same statement
			Expect(txns[28].Name).To(Equal("UPI-929092781855-Blinkit IN"))
			Expect(*txns[28].Amount).To(Equal(385.00))
			Expect(txns[28].Date).To(Equal(time.Date(2025, 7, 4, 0, 0, 0, 0, time.UTC)))
		})

		It("should parse another valid ICICI credit statement file", func() {
			// Using content from the provided CreditCardStatement(1).CSV
			filePath := filepath.Join("../../../", "CreditCardStatement(1).CSV")
			fileBytes, err := os.ReadFile(filePath)
			Expect(err).NotTo(HaveOccurred())

			txns, err := parser.Parse(fileBytes, "", "CreditCardStatement(1).CSV")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(2)) // 2 transactions in the second file

			Expect(txns[0].Name).To(Equal("GOOGLEPLAY MUMBAI IN"))
			Expect(*txns[0].Amount).To(Equal(2600.00))
			Expect(txns[0].Date).To(Equal(time.Date(2024, 8, 2, 0, 0, 0, 0, time.UTC)))

			Expect(txns[1].Name).To(Equal("HOTSTAR MUMBAI IN"))
			Expect(*txns[1].Amount).To(Equal(1271.10))
			Expect(txns[1].Date).To(Equal(time.Date(2024, 8, 11, 0, 0, 0, 0, time.UTC)))
		})

		It("should error if header row is missing", func() {
			input := `"Accountno:","0000000031229212"
"Customer Name:","MR NEDUNGADI PRANAV V"`
			_, err := parser.Parse([]byte(input), "", "")
			Expect(err).To(MatchError("transaction header row not found in ICICI credit statement"))
		})
	})

	Describe("Parser Registry", func() {
		It("should return ICICI credit parser for BankTypeICICICredit", func() {
			p, ok := GetParser(models.BankTypeICICICredit)
			Expect(ok).To(BeTrue())
			Expect(p).NotTo(BeNil())
			_, isCorrectType := p.(*ICICICreditParser)
			Expect(isCorrectType).To(BeTrue())
		})
	})
})
