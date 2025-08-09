package parser

import (
	"expenses/internal/models"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("HDFCParser", func() {
	var parser *HDFCParser

	BeforeEach(func() {
		parser = &HDFCParser{}
	})

	Describe("parseDate", func() {
		It("parses typical HDFC dd/mm/yy and dd/mm/yyyy formats", func() {
			testCases := []struct {
				input    string
				expected time.Time
			}{
				{"01/04/25", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)},
				{"1/4/25", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)},
				{"28/02/2025", time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)},
			}
			for _, tc := range testCases {
				result, err := parser.parseDate(tc.input)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(tc.expected))
			}
		})

		It("errors on unsupported date format", func() {
			_, err := parser.parseDate("2025-04-01")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("parseAmount", func() {
		It("parses numbers with commas and spaces", func() {
			testCases := []struct {
				input    string
				expected float64
			}{
				{"40.00", 40.0},
				{"1,141.74", 1141.74},
				{"2,59,000.00", 259000.0},
				{"  2,000.50  ", 2000.50},
			}
			for _, tc := range testCases {
				result, err := parser.parseAmount(tc.input)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(tc.expected))
			}
		})

		It("errors on invalid inputs", func() {
			_, err := parser.parseAmount("")
			Expect(err).To(HaveOccurred())
			_, err = parser.parseAmount("   ")
			Expect(err).To(HaveOccurred())
			_, err = parser.parseAmount("abc123")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("parseTransactionRow", func() {
		It("errors on insufficient columns", func() {
			fields := []string{"01/04/25", "Narr", "01/04/25", "192.36"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("parses a debit row", func() {
			fields := []string{"01/04/25", "UPI-ABC-XYZ-UPI", "01/04/25", "192.36", "0.00", "REF123", "138.00"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(*result.Amount).To(Equal(192.36))
			Expect(result.Description).To(ContainSubstring("Ref:"))
		})

		It("parses a credit row", func() {
			fields := []string{"28/04/25", "NEFT CR-ABCD-ORG NAME-XYZ", "28/04/25", "0.00", "11000.00", "NEFTREF123", "11120.30"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(*result.Amount).To(Equal(-11000.00))
		})

		It("errors when both debit and credit are zero/empty", func() {
			fields := []string{"01/04/25", "Something", "01/04/25", "0.00", "0.00", "0", "0.00"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("prioritizes debit when both debit and credit are present", func() {
			fields := []string{"02/02/25", "UPI-JOHN DOE-XYZ", "02/02/25", "490.00", "100.00", "REFDUAL", "0.00"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeNil())
			Expect(*result.Amount).To(Equal(490.00))
		})

		It("handles negative amounts for debit and credit", func() {
			// Negative debit remains negative
			fieldsDebit := []string{"04/08/25", "Adj", "04/08/25", "-50.00", "0.00", "REFNEG1", "0.00"}
			res1, err := parser.parseTransactionRow(fieldsDebit)
			Expect(err).NotTo(HaveOccurred())
			Expect(*res1.Amount).To(Equal(-50.00))

			// Negative credit becomes positive after sign flip
			fieldsCredit := []string{"05/08/25", "Adj", "05/08/25", "0.00", "-75.00", "REFNEG2", "0.00"}
			res2, err := parser.parseTransactionRow(fieldsCredit)
			Expect(err).NotTo(HaveOccurred())
			Expect(*res2.Amount).To(Equal(75.00))
		})

		It("omits Ref when placeholder zeros are used", func() {
			fields := []string{"01/07/25", "INTEREST PAID TILL 30-JUN-2025", "30/06/25", "0.00", "1.00", "000000000000000", "1.00"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).NotTo(HaveOccurred())
			Expect(result.Description).To(Equal("INTEREST PAID TILL 30-JUN-2025"))
		})

		It("errors on malformed amount", func() {
			fields := []string{"01/04/25", "Bad Amount", "01/04/25", "abc", "", "REFBAD", "0.00"}
			result, err := parser.parseTransactionRow(fields)
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})
	})

	Describe("Parse", func() {
		It("parses a small snippet similar to HDFC export", func() {
			input := "  Date     ,Narration                                                ,Value Dat,Debit Amount       ,Credit Amount      ,Chq/Ref Number   ,Closing Balance\n" +
				"  01/04/25  ,UPI-ABC-XYZ-UPI                                        ,01/04/25 ,         192.36     ,           0.00     ,REF123                 ,         138.00  \n" +
				"  28/04/25  ,NEFT CR-ABCD-ORG NAME-XYZ                             ,28/04/25 ,           0.00     ,       11000.00     ,NEFTREF123             ,       11120.30  \n"
			txns, err := parser.Parse([]byte(input), "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(2))
			Expect(*txns[0].Amount).To(Equal(192.36))
			Expect(*txns[1].Amount).To(Equal(-11000.00))
		})

		It("ensures CategoryIds is empty", func() {
			input := "Date,Narration,Value Dat,Debit Amount,Credit Amount,Chq/Ref Number,Closing Balance\n" +
				"01/07/25,INTEREST PAID TILL 30-JUN-2025,30/06/25,0.00,1.00,000000000000000,1.00\n"
			txns, err := parser.Parse([]byte(input), "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
			Expect(txns[0].CategoryIds).To(BeEmpty())
		})

		It("returns error when header row is missing", func() {
			_, err := parser.Parse([]byte("No header here\nJust text"), "", "")
			Expect(err).To(HaveOccurred())
		})

		It("fails when scanner encounters an overly long line", func() {
			// bufio.Scanner returns ErrTooLong when a token exceeds the buffer size
			veryLong := make([]byte, 70*1024)
			for i := range veryLong {
				veryLong[i] = 'A'
			}
			_, err := parser.Parse(veryLong, "", "")
			Expect(err).To(HaveOccurred())
		})

		It("skips empty lines between transactions", func() {
			input := "Date,Narration,Value Dat,Debit Amount,Credit Amount,Chq/Ref Number,Closing Balance\n\n" +
				"01/10/24,POS SOME MERCHANT,30/09/24,2.00,0.00,0,20185.00\n\n"
			txns, err := parser.Parse([]byte(input), "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
		})

		It("skips lines with insufficient columns", func() {
			input := "Date,Narration,Value Dat,Debit Amount,Credit Amount,Chq/Ref Number,Closing Balance\n" +
				"01/10/24,Short\n" +
				"01/10/24,POS SOME MERCHANT,30/09/24,2.00,0.00,0,20185.00\n"
			txns, err := parser.Parse([]byte(input), "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
		})

		It("skips rows with invalid date formats", func() {
			input := "Date,Narration,Value Dat,Debit Amount,Credit Amount,Chq/Ref Number,Closing Balance\n" +
				"bad-date,POS SOME MERCHANT,30/09/24,2.00,0.00,0,20185.00\n" +
				"01/10/24,POS SOME MERCHANT,30/09/24,2.00,0.00,0,20185.00\n"
			txns, err := parser.Parse([]byte(input), "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
		})

		It("skips rows with non-numeric amount", func() {
			input := "Date,Narration,Value Dat,Debit Amount,Credit Amount,Chq/Ref Number,Closing Balance\n" +
				"01/10/24,POS SOME MERCHANT,30/09/24,abc,0.00,0,20185.00\n" +
				"01/10/24,POS SOME MERCHANT,30/09/24,2.00,0.00,0,20185.00\n"
			txns, err := parser.Parse([]byte(input), "", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(1))
		})
	})

	Describe("generateTransactionName", func() {
		It("truncates overly long descriptions and prefixes Debit/Credit", func() {
			name := parser.generateTransactionName("Some very long description that should be truncated for readability", false)
			Expect(name).To(Equal("Debit: Some very long des..."))
		})
	})

	Describe("Parser Registry", func() {
		It("returns HDFC parser for BankTypeHDFC", func() {
			p, ok := GetParser(models.BankTypeHDFC)
			Expect(ok).To(BeTrue())
			Expect(p).NotTo(BeNil())
		})
	})
})
