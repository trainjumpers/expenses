package parser

import (
	"expenses/internal/models"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSBIParser(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SBI Parser Suite")
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
				{"1 Aug 2022", time.Date(2022, 8, 1, 0, 0, 0, 0, time.UTC)},
				{"01 Aug 2022", time.Date(2022, 8, 1, 0, 0, 0, 0, time.UTC)},
				{"15 Jan 2022", time.Date(2022, 1, 15, 0, 0, 0, 0, time.UTC)},
				{"2 January 2022", time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)},
				{"02 January 2022", time.Date(2022, 1, 2, 0, 0, 0, 0, time.UTC)},
			}

			for _, tc := range testCases {
				result, err := parser.parseDate(tc.input)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(tc.expected))
			}
		})

		It("should error on unsupported date format", func() {
			_, err := parser.parseDate("2022/01/15")
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
				result, err := parser.parseAmount(tc.input)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(tc.expected))
			}
		})

		It("should error on empty string", func() {
			_, err := parser.parseAmount("")
			Expect(err).To(HaveOccurred())
		})

		It("should error on spaces only", func() {
			_, err := parser.parseAmount("   ")
			Expect(err).To(HaveOccurred())
		})

		It("should error on invalid number format", func() {
			_, err := parser.parseAmount("abc123")
			Expect(err).To(HaveOccurred())
		})
		Describe("parseTransactionRow", func() {
			It("should error on insufficient columns", func() {
				fields := []string{"1 Aug 2022", "1 Aug 2022", "Desc", "123", "100.00"}
				result, err := parser.parseTransactionRow(fields)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should error on invalid date", func() {
				fields := []string{"invalid-date", "1 Aug 2022", "Desc", "123", "100.00", "", "1000.00"}
				result, err := parser.parseTransactionRow(fields)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should error on invalid amount", func() {
				fields := []string{"1 Aug 2022", "1 Aug 2022", "Desc", "123", "abc", "", "1000.00"}
				result, err := parser.parseTransactionRow(fields)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should error when both debit and credit are empty", func() {
				fields := []string{"1 Aug 2022", "1 Aug 2022", "Desc", "123", "", "", "1000.00"}
				result, err := parser.parseTransactionRow(fields)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})

			It("should parse valid debit transaction row", func() {
				fields := []string{"1 Aug 2022", "1 Aug 2022", "TO TRANSFER-UPI/DR/221356312527/RITIK  S/SBIN/rs6321908@/UPI--", "123456", "100.00", "", "1000.00"}
				result, err := parser.parseTransactionRow(fields)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Name).To(Equal("UPI to RITIK  S"))
				Expect(*result.Amount).To(Equal(100.00))
				Expect(result.Description).To(ContainSubstring("Ref: 123456"))
				Expect(result.CategoryIds).To(BeEmpty())
			})

			It("should parse valid credit transaction row", func() {
				fields := []string{"2 Aug 2022", "2 Aug 2022", "BY TRANSFER-NEFT*HDFC0000001*N215222062454075*QURIATE TECHNOLO--", "654321", "", "200.00", "1200.00"}
				result, err := parser.parseTransactionRow(fields)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(result.Name).To(Equal("NEFT from HDFC0000001"))
				Expect(*result.Amount).To(Equal(-200.00))
				Expect(result.Description).To(ContainSubstring("Ref: 654321"))
				Expect(result.CategoryIds).To(BeEmpty())
			})

			It("should parse row with both debit and credit present (prioritizes debit)", func() {
				fields := []string{"3 Aug 2022", "3 Aug 2022", "Desc", "789012", "150.00", "200.00", "1300.00"}
				result, err := parser.parseTransactionRow(fields)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(*result.Amount).To(Equal(150.00))
			})

			It("should parse negative amount in debit", func() {
				fields := []string{"4 Aug 2022", "4 Aug 2022", "Desc", "345678", "-50.00", "", "1250.00"}
				result, err := parser.parseTransactionRow(fields)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(*result.Amount).To(Equal(-50.00))
			})

			It("should parse negative amount in credit", func() {
				fields := []string{"5 Aug 2022", "5 Aug 2022", "Desc", "987654", "", "-75.00", "1175.00"}
				result, err := parser.parseTransactionRow(fields)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeNil())
				Expect(*result.Amount).To(Equal(75.00)) // -(-75.00)
			})
		})

		Describe("Parse", func() {
			It("should parse a valid SBI statement with multiple transactions", func() {
				input := `Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance
1 Aug 2022	1 Aug 2022	TO TRANSFER-UPI/DR/221356312527/RITIK  S/SBIN/rs6321908@/UPI--	123456	100.00		1000.00
2 Aug 2022	2 Aug 2022	BY TRANSFER-NEFT*HDFC0000001*N215222062454075*QURIATE TECHNOLO--	654321		200.00	1200.00
Computer Generated Statement`
				txns, err := parser.Parse([]byte(input), "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(2))
				Expect(txns[0].Name).To(Equal("UPI to RITIK  S"))
				Expect(txns[1].Name).To(Equal("NEFT from HDFC0000001"))
				Expect(txns[0].CategoryIds).To(BeEmpty())
				Expect(txns[1].CategoryIds).To(BeEmpty())
			})

			It("should parse row with both debit and credit present (prioritizes debit)", func() {
				input := `Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance
3 Aug 2022	3 Aug 2022	Desc	789012	150.00	200.00	1300.00
Computer Generated Statement`
				txns, err := parser.Parse([]byte(input), "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
				Expect(*txns[0].Amount).To(Equal(150.00))
			})

			It("should parse row with extra columns", func() {
				input := `Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance	Extra
4 Aug 2022	4 Aug 2022	Desc	345678	100.00		1250.00	ExtraValue
Computer Generated Statement`
				txns, err := parser.Parse([]byte(input), "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
			})

			It("should parse row with tabs in description", func() {
				input := "Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance\n" +
					"5 Aug 2022	5 Aug 2022	Desc with	tab	987654	100.00		1175.00\nComputer Generated Statement"
				txns, err := parser.Parse([]byte(input), "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
			})

			It("should parse header row with extra whitespace", func() {
				input := "Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance\n" +
					"7 Aug 2022	7 Aug 2022	नमस्ते	123456	100.00		1000.00\nComputer Generated Statement"
				txns, err := parser.Parse([]byte(input), "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
			})

			It("should parse row with non-ASCII characters in description", func() {
				input := "Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance\n" +
					"7 Aug 2022	7 Aug 2022	नमस्ते	123456	100.00		1000.00\nComputer Generated Statement"
				txns, err := parser.Parse([]byte(input), "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
				Expect(txns[0].Name).To(ContainSubstring("नमस्ते"))
			})

			It("should error if header row is missing", func() {
				input := "No header here\nJust some text"
				_, err := parser.Parse([]byte(input), "", "")
				Expect(err).To(MatchError(ContainSubstring("transaction header row not found")))
			})

			It("should stop parsing at 'computer generated' line", func() {
				input := `Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance
1 Aug 2022	1 Aug 2022	Desc	123456	100.00		1000.00
Computer Generated Statement
2 Aug 2022	2 Aug 2022	Desc	123456	100.00		1000.00`
				txns, err := parser.Parse([]byte(input), "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
			})

			It("should skip empty lines between transactions", func() {
				input := `Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance

 	1 Aug 2022	1 Aug 2022	Desc	123456	100.00		1000.00

 	2 Aug 2022	2 Aug 2022	Desc	123456		200.00	1200.00

 	Computer Generated Statement`
				txns, err := parser.Parse([]byte(input), "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(2))
			})

			It("should skip lines with insufficient columns", func() {
				input := `Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance
1 Aug 2022	1 Aug 2022	Short	123456	100.00
2 Aug 2022	2 Aug 2022	Desc	123456		200.00	1200.00
Computer Generated Statement`
				txns, err := parser.Parse([]byte(input), "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
			})

			It("should skip lines with both debit and credit empty", func() {
				input := `Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance
1 Aug 2022	1 Aug 2022	Desc	123456	abc		1000.00
2 Aug 2022	2 Aug 2022	Desc	123456		200.00	1200.00
Computer Generated Statement`
				txns, err := parser.Parse([]byte(input), "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
			})

			It("should skip lines with malformed amount", func() {
				input := `Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance
 	1 Aug 2022	1 Aug 2022	Desc	123456	abc		1000.00
 	2 Aug 2022	2 Aug 2022	Desc	123456		200.00	1200.00
 	Computer Generated Statement`
				txns, err := parser.Parse([]byte(input), "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
			})

			It("should handle rows with and without RefNo", func() {
				input := `Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance
1 Aug 2022	1 Aug 2022	Desc		100.00		1000.00
2 Aug 2022	2 Aug 2022	Desc	654321		200.00	1200.00
Computer Generated Statement`
				txns, err := parser.Parse([]byte(input), "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(2))
				Expect(txns[0].Description).To(Equal("Desc"))
				Expect(txns[1].Description).To(Equal("Desc (Ref: 654321)"))
			})

			It("should ensure CategoryIds is always empty", func() {
				input := `Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance
 	1 Aug 2022	1 Aug 2022	Desc	123456	100.00		1000.00
 	Computer Generated Statement`
				txns, err := parser.Parse([]byte(input), "", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
				Expect(txns[0].CategoryIds).To(BeEmpty())
			})
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
					"TO TRANSFER-UPI/DR/221356312527/RITIK  S/SBIN/rs6321908@/UPI--",
					false,
					"UPI to RITIK  S",
				},
				{
					"BY TRANSFER-UPI/CR/221356312527/RAHUL/SBIN/rs6321908@/UPI--",
					true,
					"UPI from RAHUL",
				},
				{
					"BY TRANSFER-NEFT*HDFC0000001*N215222062454075*QURIATE TECHNOLO--",
					true,
					"NEFT from HDFC0000001",
				},
				{
					"TO TRANSFER-NEFT*ICIC0000002*N215222062454075*QURIATE TECHNOLO--",
					false,
					"NEFT to ICIC0000002",
				},
				{
					"DEBIT-ATMCard AMC  607431*3795 CLASSIC--",
					false,
					"ATM Card AMC 607431 (Debit)",
				},
				{
					"CREDIT-ATMCard AMC  607431*3795 CLASSIC--",
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
					"TO TRANSFER-This is a very long description that should be truncated for readability",
					false,
					"Debit: This is a very...",
				},
				{
					"BY TRANSFER-Short Desc",
					true,
					"Credit: Short Desc",
				},
			}

			for _, tc := range testCases {
				result := parser.generateTransactionName(tc.description, tc.isCredit)
				Expect(result).To(Equal(tc.expected))
			}
		})

		// Registry tests
		var _ = Describe("Parser Registry", func() {
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
})
