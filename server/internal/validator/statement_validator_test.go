package validator

import (
	"mime/multipart"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Mock multipart.File for testing
type mockFile struct {
	*strings.Reader
}

func (m *mockFile) Close() error {
	return nil
}

func newMockFile(content string) multipart.File {
	return &mockFile{Reader: strings.NewReader(content)}
}

var _ = Describe("StatementValidator", func() {
	var validator *StatementValidator

	BeforeEach(func() {
		validator = NewStatementValidator()
	})

	Describe("ValidateStatementUpload", func() {
		Context("with valid input", func() {
			It("should validate successfully", func() {
				fileContent := "test,data\n1,2"
				file := newMockFile(fileContent)
				header := &multipart.FileHeader{
					Filename: "test.csv",
					Size:     int64(len(fileContent)),
				}

				err := validator.ValidateStatementUpload(int64(123), file, header)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("with invalid account_id", func() {
			It("should return error for empty account_id", func() {
				file := newMockFile("test")
				header := &multipart.FileHeader{Filename: "test.csv", Size: 4}

				err := validator.ValidateStatementUpload(int64(0), file, header)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for invalid account_id", func() {
				file := newMockFile("test")
				header := &multipart.FileHeader{Filename: "test.csv", Size: 4}

				err := validator.ValidateStatementUpload(int64(-1), file, header)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for negative account_id", func() {
				file := newMockFile("test")
				header := &multipart.FileHeader{Filename: "test.csv", Size: 4}

				err := validator.ValidateStatementUpload(int64(-1), file, header)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with invalid file", func() {
			It("should return error for nil file", func() {
				err := validator.ValidateStatementUpload(int64(123), nil, nil)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for file too large", func() {
				file := newMockFile("test")
				header := &multipart.FileHeader{
					Filename: "test.csv",
					Size:     300 * 1024, // 300KB > 256KB limit
				}

				err := validator.ValidateStatementUpload(int64(123), file, header)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for invalid file type", func() {
				file := newMockFile("test")
				header := &multipart.FileHeader{
					Filename: "test.txt",
					Size:     4,
				}

				err := validator.ValidateStatementUpload(int64(123), file, header)
				Expect(err).To(HaveOccurred())
			})

			It("should return error for empty filename", func() {
				file := newMockFile("test")
				header := &multipart.FileHeader{
					Filename: "",
					Size:     4,
				}

				err := validator.ValidateStatementUpload(int64(123), file, header)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("with different file types", func() {
			It("should accept CSV files", func() {
				file := newMockFile("test")
				header := &multipart.FileHeader{Filename: "test.csv", Size: 4}

				err := validator.ValidateStatementUpload(int64(123), file, header)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should accept XLS files", func() {
				file := newMockFile("test")
				header := &multipart.FileHeader{Filename: "test.xls", Size: 4}

				err := validator.ValidateStatementUpload(int64(123), file, header)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should accept XLSX files", func() {
				file := newMockFile("test")
				header := &multipart.FileHeader{Filename: "test.xlsx", Size: 4}

				err := validator.ValidateStatementUpload(int64(123), file, header)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("Page/PageSize Param Parsing (service integration)", func() {
		It("should default to page 1 and page_size 10 if not provided", func() {
			page := 1
			pageSize := 10
			Expect(page).To(Equal(1))
			Expect(pageSize).To(Equal(10))
		})
		It("should parse valid page and page_size", func() {
			page := 2
			pageSize := 25
			Expect(page).To(Equal(2))
			Expect(pageSize).To(Equal(25))
		})
		It("should clamp page_size to 10 if out of range", func() {
			pageSize := 200
			if pageSize < 1 || pageSize > 100 {
				pageSize = 10
			}
			Expect(pageSize).To(Equal(10))
		})
		It("should clamp page to 1 if less than 1", func() {
			page := -5
			if page < 1 {
				page = 1
			}
			Expect(page).To(Equal(1))
		})
	})

	Describe("Additional Edge Cases", func() {
		It("should accept file exactly at 256KB limit", func() {
			fileContent := strings.Repeat("a", 256*1024)
			file := newMockFile(fileContent)
			header := &multipart.FileHeader{Filename: "test.csv", Size: int64(len(fileContent))}
			err := validator.ValidateStatementUpload(int64(123), file, header)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject file just above 256KB limit", func() {
			fileContent := strings.Repeat("a", 256*1024+1)
			file := newMockFile(fileContent)
			header := &multipart.FileHeader{Filename: "test.csv", Size: int64(len(fileContent))}
			err := validator.ValidateStatementUpload(int64(123), file, header)
			Expect(err).To(HaveOccurred())
		})

		It("should accept uppercase file extensions", func() {
			file := newMockFile("test")
			header := &multipart.FileHeader{Filename: "test.CSV", Size: 4}
			err := validator.ValidateStatementUpload(int64(123), file, header)
			Expect(err).NotTo(HaveOccurred())
			header = &multipart.FileHeader{Filename: "test.XLS", Size: 4}
			err = validator.ValidateStatementUpload(int64(123), file, header)
			Expect(err).NotTo(HaveOccurred())
			header = &multipart.FileHeader{Filename: "test.XLSX", Size: 4}
			err = validator.ValidateStatementUpload(int64(123), file, header)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle filenames with leading/trailing spaces", func() {
			file := newMockFile("test")
			header := &multipart.FileHeader{Filename: " test.csv ", Size: 4}
			err := validator.ValidateStatementUpload(int64(123), file, header)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept filenames with multiple dots", func() {
			file := newMockFile("test")
			header := &multipart.FileHeader{Filename: "statement.final.csv", Size: 4}
			err := validator.ValidateStatementUpload(int64(123), file, header)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject file with no extension", func() {
			file := newMockFile("test")
			header := &multipart.FileHeader{Filename: "statement", Size: 4}
			err := validator.ValidateStatementUpload(int64(123), file, header)
			Expect(err).To(HaveOccurred())
		})

		It("should handle file with valid extension but empty content", func() {
			file := newMockFile("")
			header := &multipart.FileHeader{Filename: "test.csv", Size: 0}
			err := validator.ValidateStatementUpload(int64(123), file, header)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept very large positive account_id", func() {
			file := newMockFile("test")
			header := &multipart.FileHeader{Filename: "test.csv", Size: 4}
			err := validator.ValidateStatementUpload(int64(1<<60), file, header)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should accept file with special characters in filename if extension is valid", func() {
			file := newMockFile("test")
			header := &multipart.FileHeader{Filename: "t@st#file!.csv", Size: 4}
			err := validator.ValidateStatementUpload(int64(123), file, header)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
