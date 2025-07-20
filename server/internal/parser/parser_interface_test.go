package parser

import (
	"expenses/internal/models"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestParserInterface(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Parser Interface Suite")
}

var _ = Describe("Parser Interface with ParseOptions", func() {
	Describe("SBI Parser with ParseOptions", func() {
		var parser *SBIParser

		BeforeEach(func() {
			parser = &SBIParser{}
		})

		Context("with empty ParseOptions", func() {
			It("should parse normally with empty options", func() {
				input := `Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance
1 Aug 2022	1 Aug 2022	TO TRANSFER-UPI/DR/221356312527/RITIK  S/SBIN/rs6321908@/UPI--	123456	100.00		1000.00
Computer Generated Statement`
				
				options := models.NewParseOptions()
				txns, err := parser.Parse([]byte(input), options)
				
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
				Expect(txns[0].Name).To(Equal("UPI to RITIK  S"))
				Expect(*txns[0].Amount).To(Equal(100.00))
			})

			It("should ignore ParseOptions when they are empty", func() {
				input := `Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance
1 Aug 2022	1 Aug 2022	Test Transaction	123456	100.00		1000.00
2 Aug 2022	2 Aug 2022	Another Transaction	654321		200.00	1200.00
Computer Generated Statement`
				
				options := models.ParseOptions{
					SkipRows: 0,
					Mappings: []models.ColumnMapping{},
				}
				
				txns, err := parser.Parse([]byte(input), options)
				
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(2))
			})
		})

		Context("with non-empty ParseOptions", func() {
			It("should ignore custom mappings for SBI parser", func() {
				input := `Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance
1 Aug 2022	1 Aug 2022	Test Transaction	123456	100.00		1000.00
Computer Generated Statement`
				
				options := models.ParseOptions{
					SkipRows: 0,
					Mappings: []models.ColumnMapping{
						{SourceColumn: "Description", TargetField: "name"},
						{SourceColumn: "Debit", TargetField: "amount"},
					},
				}
				
				txns, err := parser.Parse([]byte(input), options)
				
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
				// SBI parser should use its own logic, not the mappings
				Expect(txns[0].Name).To(Equal("Test Transaction"))
			})

			It("should ignore skip rows for SBI parser", func() {
				input := `Header Line 1
Header Line 2
Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance
1 Aug 2022	1 Aug 2022	Test Transaction	123456	100.00		1000.00
Computer Generated Statement`
				
				options := models.ParseOptions{
					SkipRows: 2, // Should be ignored
					Mappings: []models.ColumnMapping{},
				}
				
				txns, err := parser.Parse([]byte(input), options)
				
				// SBI parser should find its own header, regardless of skip rows
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
			})
		})

		Context("error handling with ParseOptions", func() {
			It("should handle invalid ParseOptions gracefully", func() {
				input := `Txn Date	Value Date	Description	Ref No.	Debit	Credit	Balance
1 Aug 2022	1 Aug 2022	Test Transaction	123456	100.00		1000.00
Computer Generated Statement`
				
				options := models.ParseOptions{
					SkipRows: -1, // Invalid but should be ignored
					Mappings: []models.ColumnMapping{
						{SourceColumn: "", TargetField: ""}, // Invalid but should be ignored
					},
				}
				
				txns, err := parser.Parse([]byte(input), options)
				
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
			})
		})
	})

	Describe("Custom CSV Parser with ParseOptions", func() {
		var parser *CustomCSVParser

		BeforeEach(func() {
			parser = &CustomCSVParser{}
		})

		Context("with valid ParseOptions", func() {
			It("should use skip rows from ParseOptions", func() {
				input := `Header Row 1
Header Row 2
Date,Description,Amount
2022-08-01,Test Transaction,100.00
2022-08-02,Another Transaction,-200.00`
				
				options := models.ParseOptions{
					SkipRows: 2,
					Mappings: []models.ColumnMapping{
						{SourceColumn: "Date", TargetField: "date"},
						{SourceColumn: "Description", TargetField: "name"},
						{SourceColumn: "Amount", TargetField: "amount"},
					},
				}
				
				txns, err := parser.Parse([]byte(input), options)
				
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(2))
				Expect(txns[0].Name).To(Equal("Test Transaction"))
				Expect(*txns[0].Amount).To(Equal(100.00))
			})

			It("should use column mappings from ParseOptions", func() {
				input := `Transaction Date,Transaction Description,Debit Amount,Credit Amount
2022-08-01,Purchase at Store,100.00,
2022-08-02,Salary Credit,,2000.00`
				
				options := models.ParseOptions{
					SkipRows: 0,
					Mappings: []models.ColumnMapping{
						{SourceColumn: "Transaction Date", TargetField: "date"},
						{SourceColumn: "Transaction Description", TargetField: "name"},
						{SourceColumn: "Debit Amount", TargetField: "debit"},
						{SourceColumn: "Credit Amount", TargetField: "credit"},
					},
				}
				
				txns, err := parser.Parse([]byte(input), options)
				
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(2))
				Expect(txns[0].Name).To(Equal("Purchase at Store"))
				Expect(*txns[0].Amount).To(Equal(100.00))
				Expect(txns[1].Name).To(Equal("Salary Credit"))
				Expect(*txns[1].Amount).To(Equal(-2000.00))
			})

			It("should handle both skip rows and mappings together", func() {
				input := `Bank Statement
Account: 123456789
Date,Desc,Amount
2022-08-01,Test Transaction,100.00`
				
				options := models.ParseOptions{
					SkipRows: 2,
					Mappings: []models.ColumnMapping{
						{SourceColumn: "Date", TargetField: "date"},
						{SourceColumn: "Desc", TargetField: "name"},
						{SourceColumn: "Amount", TargetField: "amount"},
					},
				}
				
				txns, err := parser.Parse([]byte(input), options)
				
				Expect(err).NotTo(HaveOccurred())
				Expect(txns).To(HaveLen(1))
				Expect(txns[0].Name).To(Equal("Test Transaction"))
			})
		})

		Context("with invalid ParseOptions", func() {
			It("should error when required mappings are missing", func() {
				input := `Date,Description,Amount
2022-08-01,Test Transaction,100.00`
				
				options := models.ParseOptions{
					SkipRows: 0,
					Mappings: []models.ColumnMapping{
						{SourceColumn: "Date", TargetField: "date"},
						// Missing name and amount mappings
					},
				}
				
				_, err := parser.Parse([]byte(input), options)
				
				Expect(err).To(HaveOccurred())
			})

			It("should error when skip rows exceeds file length", func() {
				input := `Date,Description,Amount
2022-08-01,Test Transaction,100.00`
				
				options := models.ParseOptions{
					SkipRows: 10, // More than available rows
					Mappings: []models.ColumnMapping{
						{SourceColumn: "Date", TargetField: "date"},
						{SourceColumn: "Description", TargetField: "name"},
						{SourceColumn: "Amount", TargetField: "amount"},
					},
				}
				
				_, err := parser.Parse([]byte(input), options)
				
				Expect(err).To(HaveOccurred())
			})

			It("should error when mapped columns don't exist", func() {
				input := `Date,Description,Amount
2022-08-01,Test Transaction,100.00`
				
				options := models.ParseOptions{
					SkipRows: 0,
					Mappings: []models.ColumnMapping{
						{SourceColumn: "Date", TargetField: "date"},
						{SourceColumn: "NonExistentColumn", TargetField: "name"},
						{SourceColumn: "Amount", TargetField: "amount"},
					},
				}
				
				_, err := parser.Parse([]byte(input), options)
				
				Expect(err).To(HaveOccurred())
			})
		})

		Context("edge cases with ParseOptions", func() {
			It("should handle empty ParseOptions", func() {
				input := `Date,Description,Amount
2022-08-01,Test Transaction,100.00`
				
				options := models.NewParseOptions()
				
				_, err := parser.Parse([]byte(input), options)
				
				// Should error because no mappings provided
				Expect(err).To(HaveOccurred())
			})

			It("should validate ParseOptions helper methods", func() {
				emptyOptions := models.NewParseOptions()
				Expect(emptyOptions.IsEmpty()).To(BeTrue())
				Expect(emptyOptions.HasCustomMappings()).To(BeFalse())
				Expect(emptyOptions.HasRowSkipping()).To(BeFalse())
				
				fullOptions := models.ParseOptions{
					SkipRows: 2,
					Mappings: []models.ColumnMapping{
						{SourceColumn: "Date", TargetField: "date"},
					},
				}
				Expect(fullOptions.IsEmpty()).To(BeFalse())
				Expect(fullOptions.HasCustomMappings()).To(BeTrue())
				Expect(fullOptions.HasRowSkipping()).To(BeTrue())
			})
		})
	})

	Describe("Parser Registry with ParseOptions", func() {
		It("should return parsers that support ParseOptions interface", func() {
			sbiParser, ok := GetParser(models.BankTypeSBI)
			Expect(ok).To(BeTrue())
			Expect(sbiParser).To(BeAssignableToTypeOf(&SBIParser{}))
			
			// Verify the parser implements the interface with ParseOptions
			options := models.NewParseOptions()
			_, err := sbiParser.Parse([]byte("invalid"), options)
			Expect(err).To(HaveOccurred()) // Should error on invalid input, but method should exist
		})

		It("should handle unknown bank types gracefully", func() {
			parser, ok := GetParser("UNKNOWN_BANK")
			Expect(ok).To(BeFalse())
			Expect(parser).To(BeNil())
		})
	})
})