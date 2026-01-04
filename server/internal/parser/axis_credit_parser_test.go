package parser

import (
	"expenses/internal/models"
	"expenses/pkg/utils"
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
				{"14 Nov '25", "AMELIA,MUMBAI", "", "₹ 7,977.00", "Debit"},
				{"13 Nov '25", "Swiggy Limited,Bangalore", "", "₹ 1,179.00", "Debit"},
				{"11 Nov '25", "Swiggy Limited,Bengaluru", "", "₹ 908.00", "Debit"},
				{"02 Nov '25", "AL ARABIAN EXPRESS,NASIK", "", "₹ 729.00", "Debit"},
				{"20 Oct '25", "ING*IRCTC/AUTOPE,WWW.IRCTC.CO.", "", "₹ 8,033.38", "Debit"},
				{"19 Oct '25", "BBPS Payment Received - BD015292BAJAAACU9DB5", "", "₹ 107.08", "Credit"},
			}

			b := utils.CreateXLSXFile(data)
			txns, err := parser.Parse(b, "", "test.xlsx", "")
			Expect(err).NotTo(HaveOccurred())
			Expect(txns).To(HaveLen(6))

			// check first
			Expect(txns[0].Name).To(Equal("AMELIA,MUMBAI"))
			Expect(*txns[0].Amount).To(Equal(7977.00))
			Expect(txns[0].Date).To(Equal(time.Date(2025, 11, 14, 0, 0, 0, 0, time.UTC)))

			// check last credit (should be negative)
			Expect(txns[5].Name).To(Equal("BBPS Payment Received - BD015292BAJAAACU9DB5"))
			Expect(*txns[5].Amount).To(Equal(-107.08))
			Expect(txns[5].Date).To(Equal(time.Date(2025, 10, 19, 0, 0, 0, 0, time.UTC)))
		})

		It("should error if header missing", func() {
			data := [][]string{{"Some Random Header", "", ""}}
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
