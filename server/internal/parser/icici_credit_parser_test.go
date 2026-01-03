package parser

import (
	"expenses/internal/models"
	"expenses/pkg/utils"
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
		It("should parse a valid ICICI credit statement", func() {
			input := `"Accountno:","[ACCOUNT_NUMBER]"
"Customer Name:","[CUSTOMER_NAME]"
"Address:","[ADDRESS]"


"Transaction Details:"
"Date","Sr.No.","Transaction Details","Reward Point Header","Intl.Amount","Amount(in Rs)","BillingAmountSign"
"[CARD_NUMBER]"
"12/07/2025","11600218774","BOOK MY SHOW PAYU PG MUMBAI IN","12","0","612.16",""
"16/07/2025","11627037259","PAYU*SWIGGY Bangalore IN","6","0","327.00",""
"08/07/2025","11571085308","BBPS Payment received","0","0","10412.40","CR"
"04/07/2025","11500123456","UPI-929092781855-Blinkit IN","8","0","385.00",""`

			txns, err := parser.Parse([]byte(input), "", "test.csv", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(4))

			// Check debit transaction
			Expect(txns[0].Name).To(Equal("BOOK MY SHOW PAYU PG MUMBAI IN"))
			Expect(*txns[0].Amount).To(Equal(612.16))
			Expect(txns[0].Date).To(Equal(time.Date(2025, 7, 12, 0, 0, 0, 0, time.UTC)))

			// Check credit transaction
			Expect(txns[2].Name).To(Equal("BBPS Payment received"))
			Expect(*txns[2].Amount).To(Equal(-10412.40))
			Expect(txns[2].Date).To(Equal(time.Date(2025, 7, 8, 0, 0, 0, 0, time.UTC)))

			// Check another debit transaction
			Expect(txns[3].Name).To(Equal("UPI-929092781855-Blinkit IN"))
			Expect(*txns[3].Amount).To(Equal(385.00))
			Expect(txns[3].Date).To(Equal(time.Date(2025, 7, 4, 0, 0, 0, 0, time.UTC)))
		})

		It("should parse another valid ICICI credit statement format", func() {
			input := `"Accountno:","[ACCOUNT_NUMBER]"
"Customer Name:","[CUSTOMER_NAME]"
"Address:","[ADDRESS]"


"Transaction Details:"
"Date","Sr.No.","Transaction Details","Reward Point Header","Intl.Amount","Amount(in Rs)","BillingAmountSign"
"[CARD_NUMBER]"
"02/08/2024","9598980587","GOOGLEPLAY MUMBAI IN","52","0","2600.00",""
"11/08/2024","9647709815","HOTSTAR MUMBAI IN","25","0","1271.10",""`

			txns, err := parser.Parse([]byte(input), "", "test.csv", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(2))

			Expect(txns[0].Name).To(Equal("GOOGLEPLAY MUMBAI IN"))
			Expect(*txns[0].Amount).To(Equal(2600.00))
			Expect(txns[0].Date).To(Equal(time.Date(2024, 8, 2, 0, 0, 0, 0, time.UTC)))

			Expect(txns[1].Name).To(Equal("HOTSTAR MUMBAI IN"))
			Expect(*txns[1].Amount).To(Equal(1271.10))
			Expect(txns[1].Date).To(Equal(time.Date(2024, 8, 11, 0, 0, 0, 0, time.UTC)))
		})

		It("should skip card number rows and other non-transaction data", func() {
			input := `"Accountno:","[ACCOUNT_NUMBER]"
"Customer Name:","[CUSTOMER_NAME]"


"Transaction Details:"
"Date","Sr.No.","Transaction Details","Reward Point Header","Intl.Amount","Amount(in Rs)","BillingAmountSign"
"[CARD_NUMBER]"
"12/07/2025","11600218774","BOOK MY SHOW PAYU PG MUMBAI IN","12","0","612.16",""
"[ANOTHER_CARD_NUMBER]"
"16/07/2025","11627037259","PAYU*SWIGGY Bangalore IN","6","0","327.00",""`

			txns, err := parser.Parse([]byte(input), "", "test.csv", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(2)) // Should skip card number rows
		})

		It("should error if header row is missing", func() {
			input := `"Accountno:","[ACCOUNT_NUMBER]"
"Customer Name:","[CUSTOMER_NAME]"`
			_, err := parser.Parse([]byte(input), "", "", "")
			Expect(err).To(MatchError("transaction header row not found in ICICI credit statement"))
		})

		It("should handle empty transactions gracefully", func() {
			input := `"Accountno:","[ACCOUNT_NUMBER]"
"Customer Name:","[CUSTOMER_NAME]"


"Transaction Details:"
"Date","Sr.No.","Transaction Details","Reward Point Header","Intl.Amount","Amount(in Rs)","BillingAmountSign"`

			txns, err := parser.Parse([]byte(input), "", "test.csv", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(0))
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
