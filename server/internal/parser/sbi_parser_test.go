package parser

import (
	"os"
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

	Describe("Parse", func() {
		Context("with valid SBI statement file", func() {
			It("should parse transactions correctly", func() {
				// Read the sample file
				fileBytes, err := os.ReadFile("../../../1751650544162TRs1vfpDTiaqywTz.xls")
				Expect(err).NotTo(HaveOccurred())

				// Parse the file
				transactions, err := parser.Parse(fileBytes)
				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).NotTo(BeEmpty())

				// Verify first transaction
				firstTx := transactions[0]
				Expect(firstTx.Name).NotTo(BeEmpty())
				Expect(firstTx.Amount).NotTo(BeNil())
				Expect(*firstTx.Amount).To(Equal(-40.0)) // Should be negative for debit
				Expect(firstTx.Date).To(Equal(time.Date(2022, 8, 1, 0, 0, 0, 0, time.UTC)))
				Expect(firstTx.Description).To(ContainSubstring("UPI"))

				// Print some transactions for manual verification
				GinkgoWriter.Printf("Parsed %d transactions\n", len(transactions))
				for i, tx := range transactions[:5] { // Print first 5
					GinkgoWriter.Printf("Transaction %d: %s, Amount: %.2f, Date: %s\n", 
						i+1, tx.Name, *tx.Amount, tx.Date.Format("2006-01-02"))
				}
			})
		})

		Context("with second sample file", func() {
			It("should parse transactions correctly", func() {
				// Read the second sample file
				fileBytes, err := os.ReadFile("../../../1751650557652aDsDgtRwzn4kgopx.xls")
				Expect(err).NotTo(HaveOccurred())

				// Parse the file
				transactions, err := parser.Parse(fileBytes)
				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).NotTo(BeEmpty())

				// Print some transactions for manual verification
				GinkgoWriter.Printf("Parsed %d transactions from second file\n", len(transactions))
				for i, tx := range transactions[:3] { // Print first 3
					GinkgoWriter.Printf("Transaction %d: %s, Amount: %.2f, Date: %s\n", 
						i+1, tx.Name, *tx.Amount, tx.Date.Format("2006-01-02"))
				}
			})
		})
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
			}

			for _, tc := range testCases {
				result, err := parser.parseDate(tc.input)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(tc.expected))
			}
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
			}

			for _, tc := range testCases {
				result, err := parser.parseAmount(tc.input)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(tc.expected))
			}
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
					"BY TRANSFER-NEFT*HDFC0000001*N215222062454075*QURIATE TECHNOLO--",
					true,
					"NEFT from HDFC0000001",
				},
				{
					"DEBIT-ATMCard AMC  607431*3795 CLASSIC--",
					false,
					"ATM Card AMC",
				},
			}

			for _, tc := range testCases {
				result := parser.generateTransactionName(tc.description, tc.isCredit)
				Expect(result).To(Equal(tc.expected))
			}
		})
	})
})
