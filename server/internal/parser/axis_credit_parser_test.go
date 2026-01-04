package parser

import (
	"expenses/internal/models"
	"expenses/pkg/utils"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AxisCreditParser", func() {
	var parser *AxisCreditParser

	BeforeEach(func() {
		parser = &AxisCreditParser{}
	})

	Describe("Parse XLSX", func() {
		It("should parse axis credit xlsx statement correctly", func() {
			data := [][]string{
				{"Selected Statement Month", "Nov 2025", "", "", ""},
				{"", "", "", "", ""},
				{"Transaction Summary", "", "", "", ""},
				{"Date", "Transaction Details", "", "Amount (INR)", "Debit/Credit"},
				{"14 Nov '25", "Restro,MUMBAI", "", "₹ 7,977.00", "Debit"},
				{"13 Nov '25", "Swiggy Limited,Bangalore", "", "₹ 1,179.00", "Debit"},
				{"11 Nov '25", "Swiggy Limited,Bengaluru", "", "₹ 908.00", "Debit"},
				{"02 Nov '25", "AL ARABIAN EXPRESS,NASIK", "", "₹ 729.00", "Debit"},
				{"20 Oct '25", "ING*IRCTC/AUTOPE,WWW.IRCTC.CO.", "", "₹ 8,033.38", "Debit"},
				{"19 Oct '25", "BBPS Payment Received - SOME NUMBER", "", "₹ 107.08", "Credit"},
			}

			b := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(b, "", "test.xlsx", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(6))

			// check first
			Expect(txns[0].Name).To(Equal("Restro,MUMBAI"))
			Expect(*txns[0].Amount).To(Equal(7977.00))
			Expect(txns[0].Date).To(Equal(time.Date(2025, 11, 14, 0, 0, 0, 0, time.UTC)))

			// check last credit (should be negative)
			Expect(txns[5].Name).To(Equal("BBPS Payment Received - SOME NUMBER"))
			Expect(*txns[5].Amount).To(Equal(-107.08))
			Expect(txns[5].Date).To(Equal(time.Date(2025, 10, 19, 0, 0, 0, 0, time.UTC)))
		})

		It("should error if header missing", func() {
			data := [][]string{{"Some Random Header", "", ""}}
			b := utils.CreateXLSXFile(data)
			_, err := parser.Parse(b, "", "test.xlsx", "")
			Expect(err).To(MatchError("transaction header row not found in Axis credit statement"))
		})

		It("parses various amount formats and sign inference", func() {
			data := [][]string{
				{"Selected Statement Month", "Nov 2025", "", "", ""},
				{"Transaction Summary", "", "", "", ""},
				{"Date", "Transaction Details", "", "Amount (INR)", "Debit/Credit"},
				{"01 Nov '25", "Shop A", "", "₹ 1,972.06", "Debit"},
				{"02 Nov '25", "Shop B", "", "₹ 52.46", ""},
				{"03 Nov '25", "Refund", "", "(₹ 1,000.00)", ""},
				{"04 Nov '25", "Txn CR", "", "1,000.00 CR", ""},
				{"05 Nov '25", "Txn DR", "", "1,000.00 DR", ""},
				{"06 Nov '25", "Explicit Credit", "", "1,000.00", "Credit"},
				{"07 Nov '25", "Conflicting", "", "(1,234.00)", "Debit"},
			}

			b := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(b, "", "test.xlsx", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(7))

			Expect(*txns[0].Amount).To(Equal(1972.06))
			Expect(*txns[1].Amount).To(Equal(52.46))
			Expect(*txns[2].Amount).To(Equal(-1000.00))
			Expect(*txns[3].Amount).To(Equal(-1000.00))
			Expect(*txns[4].Amount).To(Equal(1000.00))
			Expect(*txns[5].Amount).To(Equal(-1000.00))
			Expect(*txns[6].Amount).To(Equal(1234.00))
		})

		It("uses description from the next column when missing", func() {
			data := [][]string{
				{"Date", "Transaction Details", "Alt Details", "Amount (INR)", "Debit/Credit"},
				{"10 Nov '25", "", "Fallback Desc", "₹ 500.00", "Debit"},
			}
			b := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(b, "", "test.xlsx", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
			Expect(txns[0].Name).To(Equal("Fallback Desc"))
		})

		It("falls back to numeric cell after description when amount column empty", func() {
			data := [][]string{
				{"Date", "Transaction Details", "Misc", "Amount (INR)", "Debit/Credit"},
				{"11 Nov '25", "StoreX", "", "", "", "₹ 1,234.00"},
			}
			b := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(b, "", "test.xlsx", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
			Expect(*txns[0].Amount).To(Equal(1234.00))
		})

		It("skips rows with malformed amounts and keeps valid rows", func() {
			data := [][]string{
				{"Date", "Transaction Details", "", "Amount (INR)", "Debit/Credit"},
				{"12 Nov '25", "BadAmt", "", "abc", "Debit"},
				{"13 Nov '25", "GoodAmt", "", "₹ 200.00", "Debit"},
			}
			b := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(b, "", "test.xlsx", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
			Expect(txns[0].Name).To(Equal("GoodAmt"))
		})

		It("supports header variations with spaced 'Debit / Credit' label", func() {
			data := [][]string{
				{"Date", "Transaction Details", "", "Amount", "Debit / Credit"},
				{"14 Nov '25", "HeaderVar", "", "₹ 50.00", "Debit"},
			}
			b := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(b, "", "test.xlsx", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
			Expect(txns[0].Name).To(Equal("HeaderVar"))
		})

		It("parseAmount returns error on invalid input", func() {
			_, _, err := parser.parseAmount("   ")
			Expect(err).To(HaveOccurred())
			_, _, err = parser.parseAmount("abc123")
			Expect(err).To(HaveOccurred())
			_, _, err = parser.parseAmount("₹")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("empty amount after cleaning"))
		})

		It("truncates long description to 40 chars in transaction name", func() {
			longDesc := strings.Repeat("LONGTEXT-", 6) + "COMPANYNAME"
			row := []string{"01 Nov '25", longDesc, "", "₹ 1,234.00", "Debit"}
			txn, err := parser.parseTransactionRow(row, 0, 1, 3, 4)
			Expect(err).NotTo(HaveOccurred())
			Expect(txn).NotTo(BeNil())
			Expect(len(txn.Name)).To(Equal(40))
			Expect(strings.HasSuffix(txn.Name, "...")).To(BeTrue())
		})

		It("returns error when amount string is not found in row", func() {
			row := []string{"01 Nov '25", "ShopX", "", "", ""}
			txn, err := parser.parseTransactionRow(row, 0, 1, 3, 4)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("amount not found"))
			Expect(txn).To(BeNil())
		})

		It("returns nil for non-transaction/invalid rows", func() {
			row := []string{"NotADate", "desc", "", "₹ 100.00", "Debit"}
			txn, err := parser.parseTransactionRow(row, 0, 1, 3, 4)
			Expect(err).To(BeNil())
			Expect(txn).To(BeNil())
		})

		It("returns error when header exists but no valid transaction rows found", func() {
			data := [][]string{
				{"Date", "Transaction Details", "", "Amount (INR)", "Debit/Credit"},
				{"", "", "", "", ""},
				{"", "", "", "", ""},
			}
			b := utils.CreateXLSXFile(data)
			_, err := parser.Parse(b, "", "test.xlsx", "")
			Expect(err).To(MatchError("transaction header row not found in Axis credit statement"))
		})
	})

	Describe("Parser Registry", func() {
		It("should return axis credit parser for BankTypeAxisCredit", func() {
			p, ok := GetParser(models.BankTypeAxisCredit)
			Expect(ok).To(BeTrue())
			Expect(p).NotTo(BeNil())
			_, isCorrectType := p.(*AxisCreditParser)
			Expect(isCorrectType).To(BeTrue())
		})
	})
})
