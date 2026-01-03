package parser

import (
	"time"

	"expenses/internal/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CustomParser", func() {
	var p *CustomParser

	BeforeEach(func() {
		p = &CustomParser{}
		RegisterParser(models.BankTypeOthers, &CustomParser{})
	})

	Describe("Preview", func() {
		Context("with a valid CSV file", func() {
			csvContent := `Header1,Header2,Header3
Row1Col1,Row1Col2,Row1Col3
Row2Col1,Row2Col2,Row2Col3
Row3Col1,Row3Col2,Row3Col3`
			fileBytes := []byte(csvContent)

			It("should parse correctly with no rows skipped and all rows returned", func() {
				preview, err := p.Preview(fileBytes, "test.csv", 0, -1, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(preview.Headers).To(Equal([]string{"Header1", "Header2", "Header3"}))
				Expect(preview.Rows).To(HaveLen(3))
				Expect(preview.Rows[0]).To(Equal([]string{"Row1Col1", "Row1Col2", "Row1Col3"}))
			})

			It("should handle skipping metadata rows before the header", func() {
				csvWithMeta := `Bank Statement
Generated on 2023-10-27
Header1,Header2,Header3
Row1Col1,Row1Col2,Row1Col3`
				preview, err := p.Preview([]byte(csvWithMeta), "test.csv", 2, -1, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(preview.Headers).To(Equal([]string{"Header1", "Header2", "Header3"}))
				Expect(preview.Rows).To(HaveLen(1))
			})

			It("should limit the number of data rows when rowSize is specified", func() {
				preview, err := p.Preview(fileBytes, "test.csv", 0, 2, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(preview.Rows).To(HaveLen(2))
			})

			It("should trim leading and trailing spaces from headers and fields", func() {
				csvWithSpaces := ` Header1,  Header2  ,Header3
  Row1Col1  ,Row1Col2  , Row1Col3 `
				preview, err := p.Preview([]byte(csvWithSpaces), "test.csv", 0, -1, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(preview.Headers).To(Equal([]string{"Header1", "Header2", "Header3"}))
				Expect(preview.Rows[0]).To(Equal([]string{"Row1Col1", "Row1Col2", "Row1Col3"}))
			})
		})

		Context("with a valid XLS file (treated as TSV)", func() {
			It("should parse a simple file correctly", func() {
				xlsContent := "Header1\tHeader2\tHeader3\nRow1Col1\tRow1Col2\tRow1Col3"
				fileBytes := []byte(xlsContent)
				preview, err := p.Preview(fileBytes, "test.xls", 0, -1, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(preview.Headers).To(Equal([]string{"Header1", "Header2", "Header3"}))
				Expect(preview.Rows).To(HaveLen(1))
			})

			It("should parse correctly, respecting skipRows and rowSize", func() {
				xlsContent := `Bank Statement
Generated on 2023-10-27
Header1	Header2	Header3
Row1Col1	Row1Col2	Row1Col3
Row2Col1	Row2Col2	Row2Col3
Row3Col1	Row3Col2	Row3Col3`
				fileBytes := []byte(xlsContent)

				preview, err := p.Preview(fileBytes, "test.xls", 2, 2, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(preview.Headers).To(Equal([]string{"Header1", "Header2", "Header3"}))
				Expect(preview.Rows).To(HaveLen(2))
				Expect(preview.Rows[0]).To(Equal([]string{"Row1Col1", "Row1Col2", "Row1Col3"}))
				Expect(preview.Rows[1]).To(Equal([]string{"Row2Col1", "Row2Col2", "Row2Col3"}))
			})
		})

		Context("with edge cases", func() {
			It("should return an error for an unsupported file type", func() {
				_, err := p.Preview([]byte("content"), "statement.pdf", 0, -1, "")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unsupported file type for preview"))
			})

			It("should handle an empty file", func() {
				preview, err := p.Preview([]byte(""), "test.csv", 0, -1, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(preview.Headers).To(BeEmpty())
				Expect(preview.Rows).To(BeEmpty())
			})

			It("should handle a file with only a header row", func() {
				preview, err := p.Preview([]byte("Header1,Header2"), "test.csv", 0, -1, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(preview.Headers).To(Equal([]string{"Header1", "Header2"}))
				Expect(preview.Rows).To(BeEmpty())
			})

			It("should handle when skipRows is greater than the number of available rows", func() {
				preview, err := p.Preview([]byte("row1\nrow2"), "test.csv", 5, -1, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(preview.Headers).To(BeEmpty())
				Expect(preview.Rows).To(BeEmpty())
			})

			It("should handle rowSize being zero, returning headers but no rows", func() {
				preview, err := p.Preview([]byte("H1,H2\nR1,R2"), "test.csv", 0, 0, "")
				Expect(err).NotTo(HaveOccurred())
				Expect(preview.Headers).To(Equal([]string{"H1", "H2"}))
				Expect(preview.Rows).To(BeEmpty())
			})
		})
	})

	Describe("Parse", func() {
		Context("with valid metadata and file", func() {
			It("should parse transactions correctly with a single amount column", func() {
				csvContent := `Date,Payee,Amount
2024-01-15,Supermarket,150.75`
				metadata := `{
					"skip_rows": 0,
					"column_mapping": { "txn_date": "Date", "name": "Payee", "amount": "Amount" }
				}`
				transactions, err := p.Parse([]byte(csvContent), metadata, "test.csv", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(HaveLen(1))
				Expect(transactions[0].Name).To(Equal("Supermarket"))
				Expect(transactions[0].Date.Year()).To(Equal(2024))
				Expect(transactions[0].Date.Month()).To(Equal(time.January))
				Expect(transactions[0].Date.Day()).To(Equal(15))
				Expect(*transactions[0].Amount).To(Equal(150.75))
			})

			It("should parse transactions correctly with separate credit and debit columns", func() {
				csvContent := `Date,Description,Credit,Debit
2024-01-17,Salary,5000.00,
2024-01-18,Groceries,,75.20`
				metadata := `{
					"skip_rows": 0,
					"column_mapping": { "txn_date": "Date", "name": "Description", "credit": "Credit", "debit": "Debit" }
				}`
				transactions, err := p.Parse([]byte(csvContent), metadata, "test.csv", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(HaveLen(2))
				Expect(*transactions[0].Amount).To(Equal(5000.00))
				Expect(*transactions[1].Amount).To(Equal(-75.20))
			})

			It("should correctly handle an optional description field when present", func() {
				csvWithDesc := `Date,Payee,Desc,Amount
2024-01-15,Supermarket,Weekly Groceries,150.75`
				metadata := `{
					"skip_rows": 0,
					"column_mapping": { "txn_date": "Date", "name": "Payee", "description": "Desc", "amount": "Amount" }
				}`
				transactions, err := p.Parse([]byte(csvWithDesc), metadata, "test.csv", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(HaveLen(1))
				Expect(transactions[0].Description).To(Equal("Weekly Groceries"))
			})

			It("should correctly skip rows as specified in metadata", func() {
				csvContent := `Bank Statement
Account: 12345
Date,Payee,Amount
2024-01-15,Supermarket,150.75`
				metadata := `{
					"skip_rows": 2,
					"column_mapping": { "txn_date": "Date", "name": "Payee", "amount": "Amount" }
				}`
				transactions, err := p.Parse([]byte(csvContent), metadata, "test.csv", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(HaveLen(1))
				Expect(transactions[0].Name).To(Equal("Supermarket"))
			})

			It("should parse correctly when optional description field is not mapped", func() {
				csvContent := `Date,Payee,Amount
2024-01-15,Supermarket,150.75`
				metadata := `{
					"skip_rows": 0,
					"column_mapping": { "txn_date": "Date", "name": "Payee", "amount": "Amount" }
				}`
				transactions, err := p.Parse([]byte(csvContent), metadata, "test.csv", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(HaveLen(1))
				Expect(transactions[0].Description).To(BeEmpty())
			})

			It("should gracefully skip empty or malformed lines", func() {
				csvContent := `Date,Payee,Amount

2024-01-15,Supermarket,150.75

malformed-line-without-enough-columns
2024-01-16,Gas Station,45.50
`
				metadata := `{
					"skip_rows": 0,
					"column_mapping": { "txn_date": "Date", "name": "Payee", "amount": "Amount" }
				}`
				transactions, err := p.Parse([]byte(csvContent), metadata, "test.csv", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(HaveLen(2))
				Expect(transactions[0].Name).To(Equal("Supermarket"))
				Expect(transactions[1].Name).To(Equal("Gas Station"))
			})
		})

		Context("with invalid or problematic data", func() {
			It("should return an error if metadata is an empty string", func() {
				_, err := p.Parse([]byte(""), "", "test.csv", "")
				Expect(err).To(MatchError("metadata is required for custom parser"))
			})

			It("should return an error if metadata is not valid JSON", func() {
				_, err := p.Parse([]byte(""), "{not-json}", "test.csv", "")
				Expect(err.Error()).To(ContainSubstring("failed to unmarshal metadata"))
			})

			It("should return an error if a mapped column is not found in the header", func() {
				csvContent := `Date,Payee,Value
2024-01-15,Supermarket,150.75`
				metadata := `{
					"skip_rows": 0,
					"column_mapping": { "txn_date": "Date", "name": "Payee", "amount": "Amount" }
				}`
				_, err := p.Parse([]byte(csvContent), metadata, "test.csv", "")
				Expect(err).To(MatchError("mapped column 'Amount' not found in statement header"))
			})

			It("should return an error for insufficient amount information in metadata", func() {
				csvContent := `Date,Payee,Credit
2024-01-15,Supermarket,150.75`
				metadata := `{
					"skip_rows": 0,
					"column_mapping": { "txn_date": "Date", "name": "Payee", "credit": "Credit" }
				}`
				_, err := p.Parse([]byte(csvContent), metadata, "test.csv", "")
				Expect(err).To(MatchError("insufficient amount information in metadata: map either 'amount' or both 'credit' and 'debit'"))
			})

			It("should return an error if a required field is not mapped in metadata", func() {
				csvContent := `Date,Amount
2024-01-15,150.75`
				metadata := `{
					"skip_rows": 0,
					"column_mapping": { "txn_date": "Date", "amount": "Amount" }
				}`
				_, err := p.Parse([]byte(csvContent), metadata, "test.csv", "")
				Expect(err).To(MatchError("required field 'name' is not mapped in metadata"))
			})

			It("should skip rows with parsing errors and continue processing valid rows", func() {
				csvContent := `Date,Payee,Amount
not-a-date,Supermarket,150.75
2024-01-16,Gas Station,not-a-number
2024-01-17,Restaurant,120.00
2024-01-18,Coffee Shop`
				metadata := `{
					"skip_rows": 0,
					"column_mapping": { "txn_date": "Date", "name": "Payee", "amount": "Amount" }
				}`
				transactions, err := p.Parse([]byte(csvContent), metadata, "test.csv", "")
				Expect(err).NotTo(HaveOccurred())
				Expect(transactions).To(HaveLen(1))
				Expect(transactions[0].Name).To(Equal("Restaurant"))
				Expect(*transactions[0].Amount).To(Equal(120.00))
			})
		})
	})
})
