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

var _ = Describe("Mapper", func() {
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
			var nilInterface interface{}
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
	})
})
