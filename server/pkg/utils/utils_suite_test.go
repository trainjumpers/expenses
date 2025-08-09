package utils

import (
	"reflect"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Utils Suite")
}

type TestStruct struct {
	Name      string
	Age       int
	CreatedAt time.Time
	PtrField  *string
	//lint:ignore U1000 This is a test struct
	unexported string
}

var _ = Describe("Utils", func() {
	Describe("ExtractFields", func() {
		Context("with invalid inputs", func() {
			It("should return error for nil input", func() {
				ptrs, values, fields, err := ExtractFields(nil, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("extractFields: obj is nil"))
				Expect(ptrs).To(BeNil())
				Expect(values).To(BeNil())
				Expect(fields).To(BeNil())
			})

			It("should return error for non-pointer input", func() {
				ptrs, values, fields, err := ExtractFields(TestStruct{}, false)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("extractFields: obj must be a non-nil pointer to a struct"))
				Expect(ptrs).To(BeNil())
				Expect(values).To(BeNil())
				Expect(fields).To(BeNil())
			})
		})

		Context("with valid inputs", func() {
			It("should extract all fields from a valid struct", func() {
				input := &TestStruct{Name: "John", Age: 30}
				ptrs, values, fields, err := ExtractFields(input, false)

				Expect(err).NotTo(HaveOccurred())
				Expect(fields).To(HaveLen(4)) // Name, Age, CreatedAt, PtrField
				Expect(ptrs).To(HaveLen(4))
				Expect(values).To(HaveLen(4))

				Expect(fields).To(ContainElements("Name", "Age", "CreatedAt", "PtrField"))
				Expect(values[0]).To(Equal("John"))
				Expect(values[1]).To(Equal(30))
			})

			It("should skip unexported fields", func() {
				input := &TestStruct{Name: "John", Age: 30}
				_, _, fields, err := ExtractFields(input, false)

				Expect(err).NotTo(HaveOccurred())
				Expect(fields).NotTo(ContainElement("unexported"))
			})

			It("should skip null fields when skipNull is true", func() {
				input := &TestStruct{Name: "John"}
				ptrs, values, fields, err := ExtractFields(input, true)

				Expect(err).NotTo(HaveOccurred())
				Expect(fields).To(HaveLen(1)) // Only Name should be included
				Expect(ptrs).To(HaveLen(1))
				Expect(values).To(HaveLen(1))

				Expect(fields[0]).To(Equal("Name"))
				Expect(values[0]).To(Equal("John"))
			})
		})
	})

	Describe("IsZeroValue", func() {
		var nilPtr *string
		var zeroTime time.Time
		var nonZeroTime = time.Now()

		It("should correctly identify zero values", func() {
			Expect(IsZeroValue(reflect.ValueOf(nilPtr))).To(BeTrue())
			Expect(IsZeroValue(reflect.ValueOf(""))).To(BeTrue())
			Expect(IsZeroValue(reflect.ValueOf("hello"))).To(BeFalse())
			Expect(IsZeroValue(reflect.ValueOf(0))).To(BeTrue())
			Expect(IsZeroValue(reflect.ValueOf(42))).To(BeFalse())
			Expect(IsZeroValue(reflect.ValueOf(zeroTime))).To(BeTrue())
			Expect(IsZeroValue(reflect.ValueOf(nonZeroTime))).To(BeFalse())
			var nilInterface any
			Expect(IsZeroValue(reflect.ValueOf(nilInterface))).To(BeTrue())
		})
	})

	Describe("ConvertStruct", func() {
		type SrcStruct struct {
			Name            string
			Age             int
			Address         string
			unexportedField bool
		}

		type DstStruct struct {
			Name    string
			Age     int
			Address string
		}

		type DstPartialStruct struct {
			Name string
			Age  int
		}

		type DstExtraStruct struct {
			Name    string
			Age     int
			Address string
			Extra   string
		}

		type DstWrongTypeStruct struct {
			Name string
			Age  string // Mismatched type
		}

		var src *SrcStruct

		BeforeEach(func() {
			src = &SrcStruct{
				Name:            "John Doe",
				Age:             30,
				Address:         "123 Main St",
				unexportedField: true,
			}
		})

		Context("with identical structs", func() {
			It("should copy all matching fields", func() {
				dst := &DstStruct{}
				ConvertStruct(src, dst)
				Expect(dst.Name).To(Equal(src.Name))
				Expect(dst.Age).To(Equal(src.Age))
				Expect(dst.Address).To(Equal(src.Address))
			})
		})

		Context("with destination struct having a subset of fields", func() {
			It("should copy only the common fields", func() {
				dst := &DstPartialStruct{}
				ConvertStruct(src, dst)
				Expect(dst.Name).To(Equal(src.Name))
				Expect(dst.Age).To(Equal(src.Age))
			})
		})

		Context("with destination struct having extra fields", func() {
			It("should copy common fields and leave extra fields with their default values", func() {
				dst := &DstExtraStruct{Extra: "initial"}
				ConvertStruct(src, dst)
				Expect(dst.Name).To(Equal(src.Name))
				Expect(dst.Age).To(Equal(src.Age))
				Expect(dst.Address).To(Equal(src.Address))
				Expect(dst.Extra).To(Equal("initial"))
			})
		})

		Context("with fields of different types", func() {
			It("should not copy fields with mismatched types", func() {
				dst := &DstWrongTypeStruct{Name: "old name", Age: "old age"}
				ConvertStruct(src, dst)
				Expect(dst.Name).To(Equal(src.Name))
				Expect(dst.Age).To(Equal("old age"))
			})
		})

		Context("with nil or non-pointer inputs", func() {
			It("should do nothing if src is nil", func() {
				dst := &DstStruct{Name: "test"}
				ConvertStruct(nil, dst)
				Expect(dst.Name).To(Equal("test"))
			})

			It("should do nothing if dst is nil", func() {
				src := &SrcStruct{Name: "test"}
				Expect(func() { ConvertStruct(src, nil) }).ToNot(Panic())
			})

			It("should do nothing if src is a nil pointer", func() {
				var nilSrc *SrcStruct
				dst := &DstStruct{Name: "test"}
				ConvertStruct(nilSrc, dst)
				Expect(dst.Name).To(Equal("test"))
			})

			It("should do nothing if dst is a nil pointer", func() {
				src := &SrcStruct{Name: "test"}
				var nilDst *DstStruct
				Expect(func() { ConvertStruct(src, nilDst) }).ToNot(Panic())
			})

			It("should do nothing for non-pointer src", func() {
				dst := &DstStruct{Name: "test"}
				ConvertStruct(*src, dst)
				Expect(dst.Name).To(Equal("test"))
			})

			It("should do nothing for non-pointer dst", func() {
				dst := DstStruct{Name: "test"}
				ConvertStruct(src, dst)
				Expect(dst.Name).To(Equal("test"))
			})
		})
	})

	Describe("ParseDate", func() {
		Context("with SBI date formats", func() {
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
					result, err := ParseDate(tc.input)
					Expect(err).NotTo(HaveOccurred(), "Failed to parse: %s", tc.input)
					Expect(result).To(Equal(tc.expected), "Mismatch for input: %s", tc.input)
				}
			})
		})

		Context("with HDFC date formats", func() {
			It("should parse HDFC dd/mm/yy and dd/mm/yyyy formats", func() {
				testCases := []struct {
					input    string
					expected time.Time
				}{
					{"01/04/25", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)},
					{"1/4/25", time.Date(2025, 4, 1, 0, 0, 0, 0, time.UTC)},
					{"28/02/2025", time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)},
					{"2/1/2006", time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC)},
				}

				for _, tc := range testCases {
					result, err := ParseDate(tc.input)
					Expect(err).NotTo(HaveOccurred(), "Failed to parse: %s", tc.input)
					Expect(result).To(Equal(tc.expected), "Mismatch for input: %s", tc.input)
				}
			})
		})

		Context("with ICICI date formats", func() {
			It("should parse ICICI DD/MM/YYYY format", func() {
				testCases := []struct {
					input    string
					expected time.Time
				}{
					{"12/07/2025", time.Date(2025, 7, 12, 0, 0, 0, 0, time.UTC)},
					{"08/07/2025", time.Date(2025, 7, 8, 0, 0, 0, 0, time.UTC)},
					{"02/08/2024", time.Date(2024, 8, 2, 0, 0, 0, 0, time.UTC)},
				}

				for _, tc := range testCases {
					result, err := ParseDate(tc.input)
					Expect(err).NotTo(HaveOccurred(), "Failed to parse: %s", tc.input)
					Expect(result).To(Equal(tc.expected), "Mismatch for input: %s", tc.input)
				}
			})
		})

		Context("with standard date formats", func() {
			expectedDate := time.Date(2023, 10, 26, 0, 0, 0, 0, time.UTC)
			expectedDateTime := time.Date(2023, 10, 26, 10, 0, 0, 0, time.UTC)

			// Helper function to compare dates without time
			dateComparator := func(parsedTime, expectedTime time.Time) bool {
				y1, m1, d1 := parsedTime.Date()
				y2, m2, d2 := expectedTime.Date()
				return y1 == y2 && m1 == m2 && d1 == d2
			}

			It("should parse layout 2006-01-02", func() {
				t, err := ParseDate("2023-10-26")
				Expect(err).NotTo(HaveOccurred())
				Expect(dateComparator(t, expectedDate)).To(BeTrue())
			})
			It("should parse layout 01-02-2006", func() {
				t, err := ParseDate("10-26-2023")
				Expect(err).NotTo(HaveOccurred())
				Expect(dateComparator(t, expectedDate)).To(BeTrue())
			})
			It("should parse layout 01/02/2006", func() {
				t, err := ParseDate("10/26/2023")
				Expect(err).NotTo(HaveOccurred())
				Expect(dateComparator(t, expectedDate)).To(BeTrue())
			})
			It("should parse layout Jan 2, 2006", func() {
				t, err := ParseDate("Oct 26, 2023")
				Expect(err).NotTo(HaveOccurred())
				Expect(dateComparator(t, expectedDate)).To(BeTrue())
			})
			It("should parse layout RFC3339", func() {
				t, err := ParseDate("2023-10-26T10:00:00Z")
				Expect(err).NotTo(HaveOccurred())
				Expect(t.Equal(expectedDateTime)).To(BeTrue())
			})
		})

		Context("with invalid date strings", func() {
			It("should return an error for an invalid format", func() {
				_, err := ParseDate("not-a-date")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("unable to parse date: not-a-date"))
			})
			It("should return an error for an empty string", func() {
				_, err := ParseDate("")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("unable to parse date: "))
			})
			It("should return an error for invalid date values", func() {
				_, err := ParseDate("32/01/2025")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("ParseFloat", func() {
		Context("with valid amount strings", func() {
			It("should parse a standard float string", func() {
				f, err := ParseFloat("123.45")
				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(Equal(123.45))
			})
			It("should parse a string with commas", func() {
				f, err := ParseFloat("1,234.56")
				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(Equal(1234.56))
			})
			It("should parse amounts with Indian comma formatting", func() {
				f, err := ParseFloat("2,59,000.00")
				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(Equal(259000.0))
			})
			It("should parse an integer string", func() {
				f, err := ParseFloat("789")
				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(Equal(789.0))
			})
			It("should handle leading/trailing spaces", func() {
				f, err := ParseFloat("  987.65  ")
				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(Equal(987.65))
			})
			It("should handle a negative number", func() {
				f, err := ParseFloat("-50.25")
				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(Equal(-50.25))
			})
			It("should parse zero values", func() {
				f, err := ParseFloat("0.00")
				Expect(err).NotTo(HaveOccurred())
				Expect(f).To(Equal(0.0))
			})
		})

		Context("with invalid amount strings", func() {
			It("should return an error for empty string", func() {
				_, err := ParseFloat("")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("empty amount string"))
			})
			It("should return an error for spaces only", func() {
				_, err := ParseFloat("   ")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("empty amount string"))
			})
			It("should return an error for invalid number format", func() {
				_, err := ParseFloat("not-a-number")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid amount format"))
			})
			It("should return an error for mixed text and numbers", func() {
				_, err := ParseFloat("abc123")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid amount format"))
			})
		})
	})
})
