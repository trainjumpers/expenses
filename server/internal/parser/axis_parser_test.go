package parser

import (
	"math"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AxisParser", func() {
	var parser *AxisParser

	BeforeEach(func() {
		parser = &AxisParser{}
	})

	Describe("parseTransactionRow", func() {
		It("should error on insufficient columns", func() {
			fields := []string{"01-12-2025", "Short", "X"}
			res, err := parser.parseTransactionRow(fields)
			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("should error on invalid date", func() {
			fields := []string{"invalid", "-", "IMPS/XXX", "100.00", "", "1000.00"}
			res, err := parser.parseTransactionRow(fields)
			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
		})

		It("should parse debit row", func() {
			fields := []string{"31-03-2025", "-", "UPI/P2M/509038927105/Yes Bank Partner Sell/Accoun/YesBank_Yespay", "1.00", "", "131999.00"}
			res, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(*res.Amount).To(BeNumerically("==", 1.00))
			Expect(res.Name).To(ContainSubstring("UPI to"))
		})

		It("should parse credit row", func() {
			fields := []string{"31-03-2025", "-", "IMPS/P2A/509017158423/Nedungad/Remitter/salary/9177359940927139000", "", "132000.00", "132000.00"}
			res, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			Expect(*res.Amount).To(BeNumerically("==", -132000.00))
			Expect(res.Name).To(ContainSubstring("IMPS from"))
		})
	})

	Describe("Parse", func() {
		It("should parse a synthetic axis csv and return multiple transactions without PII", func() {
			// Build a synthetic CSV content (no PII) that mimics Axis statement structure
			csvContent := `Name :- SAMPLE USER
Statement of Account No - 000000000000 for the period (From : 01-03-2025 To : 04-01-2026)

Tran Date,CHQNO,PARTICULARS,DR,CR,BAL,SOL
31-03-2025,-,IMPS/P2A/000000000000/SENDER/Remitter/salary/0000000000, ,132000.00,132000.00,4806
31-03-2025,-,UPI/P2M/000000000000/MERCHANT1/UPI/,1.00, ,131999.00,4806
01-04-2025,-,NEFT/ICIC0000001/ACME_CORP,5000.00, ,127010.00,4806
30-04-2025,-,RTGS/REF0001/LARGE_PAYMENT_BANK, ,223771.00,248108.82,248
30-05-2025,-,UPI/P2A/000000000001/CUSTOMER/UPI/,100000.00, ,316304.42,4806
`

			fileBytes := []byte(csvContent)

			txns, err := parser.Parse(fileBytes, "", "test.csv", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).ToNot(BeEmpty())

			// ensure at least one large amount exists and patterns were recognized
			foundLarge := false
			foundUPI := false
			for _, t := range txns {
				if t.Amount != nil && math.Abs(*t.Amount) >= 100000 {
					foundLarge = true
				}
				if strings.Contains(strings.ToLower(t.Description), "upi") || strings.Contains(strings.ToLower(t.Name), "upi") {
					foundUPI = true
				}
			}
			Expect(foundLarge).To(BeTrue())
			Expect(foundUPI).To(BeTrue())
		})

		It("should return error for CSV without a recognizable header", func() {
			csvContent := `random,values,only
31-03-2025,1,2,3,4,5`
			fileBytes := []byte(csvContent)
			txns, err := parser.Parse(fileBytes, "", "bad.csv", "")
			Expect(err).To(HaveOccurred())
			Expect(txns).To(BeNil())
		})

		It("should skip rows with fewer than 6 columns", func() {
			csvContent := `Tran Date,CHQNO,PARTICULARS,DR,CR,BAL
31-03-2025,-,SHORT,1.00`
			fileBytes := []byte(csvContent)
			txns, err := parser.Parse(fileBytes, "", "short.csv", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(BeEmpty())
		})

		It("should handle very long descriptions by truncating the generated name", func() {
			longDesc := strings.Repeat("LONGTEXT-", 10) + "COMPANY/EXTRA"
			fields := []string{"31-03-2025", "-", longDesc, "10.00", "", "1000.00"}
			res, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(res).NotTo(BeNil())
			// Name should be prefixed with Debit and truncated to 40 runes (37 + "...")
			Expect(strings.HasPrefix(res.Name, "Debit: ")).To(BeTrue())
			Expect(strings.HasSuffix(res.Name, "...")).To(BeTrue())
			Expect(len([]rune(res.Name))).To(Equal(40))
		})

		It("should parse different debit patterns (NEFT, RTGS) correctly", func() {
			neft := []string{"31-03-2025", "-", "NEFT/ICIC0000001/ACME_CORP", "5000.00", "", "127010.00"}
			res, err := parser.parseTransactionRow(neft)
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Name).To(Equal("NEFT to ICIC0000001"))

			rtgs := []string{"30-04-2025", "-", "RTGS/REF0001/LARGE_PAYMENT_BANK/", "1000.00", "", "248108.82"}
			res2, err2 := parser.parseTransactionRow(rtgs)
			Expect(err2).NotTo(HaveOccurred())
			Expect(res2.Name).To(Equal("RTGS to LARGE_PAYMENT_BANK"))
		})

		It("should return error when transaction date is empty", func() {
			fields := []string{"", "-", "NEFT/ICIC0000001/ACME_CORP", "5000.00", "", "127010.00"}
			res, err := parser.parseTransactionRow(fields)
			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("empty transaction date"))
		})

		It("should error when credit is not a float", func() {
			fields := []string{"31-03-2025", "-", "IMPS/P2A/509017158423/Example", "", "notanumber", "132000.00"}
			res, err := parser.parseTransactionRow(fields)
			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to parse credit amount"))
		})

		It("should error when both credit and debit are empty", func() {
			fields := []string{"31-03-2025", "-", "IMPS/P2A/509017158423/Example", "", "", "132000.00"}
			res, err := parser.parseTransactionRow(fields)
			Expect(err).To(HaveOccurred())
			Expect(res).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("both debit and credit amounts are empty"))
		})
	})
})
