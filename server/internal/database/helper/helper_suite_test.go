package helper_test

import (
	"expenses/internal/database/helper"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestHelper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Helper Suite")
}

type TestStruct struct {
	Id        int     `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Age       int     `json:"age"`
	Score     float64 `json:"score"`
}

var _ = Describe("Database Helper Utils", func() {
	Describe("CreateUpdateParams", func() {
		It("should generate correct update parameters for a struct", func() {
			obj := &TestStruct{
				Id:        1,
				FirstName: "John",
				LastName:  "Doe",
				Age:       30,
				Score:     95.5,
			}

			clause, values, paramCount, err := helper.CreateUpdateParams(obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(clause).To(ContainSubstring("id = $1"))
			Expect(clause).To(ContainSubstring("first_name = $2"))
			Expect(clause).To(ContainSubstring("last_name = $3"))
			Expect(clause).To(ContainSubstring("age = $4"))
			Expect(clause).To(ContainSubstring("score = $5"))
			Expect(values).To(HaveLen(5))
			Expect(paramCount).To(Equal(6))
		})

		It("should return error for nil pointer", func() {
			var obj *TestStruct
			_, _, _, err := helper.CreateUpdateParams(obj)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-pointer", func() {
			obj := TestStruct{}
			_, _, _, err := helper.CreateUpdateParams(obj)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when trying to update with no fields", func() {
			obj := &TestStruct{}

			_, _, _, err := helper.CreateUpdateParams(obj)
			Expect(err).To(HaveOccurred())
		})

	})

	Describe("CreateInsertQuery", func() {
		It("should generate correct insert query for a struct", func() {
			insertObj := &TestStruct{
				FirstName: "John",
				LastName:  "Doe",
				Age:       30,
				Score:     95.5,
			}
			outputObj := &TestStruct{}

			query, values, ptrs, err := helper.CreateInsertQuery(insertObj, outputObj, "test_table", "public")
			Expect(err).NotTo(HaveOccurred())
			Expect(query).To(ContainSubstring("INSERT INTO public.test_table"))
			Expect(query).To(ContainSubstring("RETURNING"))
			Expect(values).To(HaveLen(4))
			Expect(ptrs).To(HaveLen(5))
			expectedQuery := `INSERT INTO public.test_table (first_name, last_name, age, score) VALUES ($1, $2, $3, $4) RETURNING id, first_name, last_name, age, score;`
			Expect(strings.TrimSpace(query)).To(Equal(expectedQuery))
		})

		It("should return error for empty struct", func() {
			insertObj := &TestStruct{}
			outputObj := &TestStruct{}
			_, _, _, err := helper.CreateInsertQuery(insertObj, outputObj, "test_table", "public")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetDbFieldsFromObject", func() {
		It("should return correct field pointers and names", func() {
			obj := &TestStruct{
				Id:        1,
				FirstName: "John",
				LastName:  "Doe",
			}

			ptrs, fields, err := helper.GetDbFieldsFromObject(obj)
			Expect(err).NotTo(HaveOccurred())
			Expect(ptrs).To(HaveLen(5))
			Expect(fields).To(ContainElements("id", "first_name", "last_name", "age", "score"))
		})

		It("should return error for nil pointer", func() {
			var obj *TestStruct
			_, _, err := helper.GetDbFieldsFromObject(obj)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("toSnakeCase", func() {
		It("should handle PascalCase and camelCase", func() {
			Expect(helper.ToSnakeCase("FirstName")).To(Equal("first_name"))
			Expect(helper.ToSnakeCase("firstName")).To(Equal("first_name"))
			Expect(helper.ToSnakeCase("first_name")).To(Equal("first_name"))
		})

		It("should handle acronyms and special cases", func() {
			Expect(helper.ToSnakeCase("HTTPRequest")).To(Equal("http_request"))
			Expect(helper.ToSnakeCase("HTTPServer")).To(Equal("http_server"))
			Expect(helper.ToSnakeCase("URLParser")).To(Equal("url_parser"))
			Expect(helper.ToSnakeCase("APIClient")).To(Equal("api_client"))
			Expect(helper.ToSnakeCase("RESTAPI")).To(Equal("restapi"))
		})

		It("should handle mixed cases", func() {
			Expect(helper.ToSnakeCase("UserIDAndName")).To(Equal("user_id_and_name"))
			Expect(helper.ToSnakeCase("testCase")).To(Equal("test_case"))
			Expect(helper.ToSnakeCase("TestCase")).To(Equal("test_case"))
			Expect(helper.ToSnakeCase("Test Case")).To(Equal("test_case"))
			Expect(helper.ToSnakeCase(" Test Case")).To(Equal("test_case"))
			Expect(helper.ToSnakeCase("Test Case ")).To(Equal("test_case"))
			Expect(helper.ToSnakeCase(" Test Case ")).To(Equal("test_case"))
			Expect(helper.ToSnakeCase("test")).To(Equal("test"))
			Expect(helper.ToSnakeCase("test_case")).To(Equal("test_case"))
			Expect(helper.ToSnakeCase("Test")).To(Equal("test"))
			Expect(helper.ToSnakeCase("")).To(Equal(""))
			Expect(helper.ToSnakeCase("ManyManyWords")).To(Equal("many_many_words"))
			Expect(helper.ToSnakeCase("manyManyWords")).To(Equal("many_many_words"))
			Expect(helper.ToSnakeCase("AnyKind of_string")).To(Equal("any_kind_of_string"))
			Expect(helper.ToSnakeCase("numbers2and55with000")).To(Equal("numbers_2_and_55_with_000"))
			Expect(helper.ToSnakeCase("JSONData")).To(Equal("json_data"))
			Expect(helper.ToSnakeCase("userID")).To(Equal("user_id"))
			Expect(helper.ToSnakeCase("AAAbbb")).To(Equal("aa_abbb"))
			Expect(helper.ToSnakeCase("1A2")).To(Equal("1_a_2"))
			Expect(helper.ToSnakeCase("A1B")).To(Equal("a_1_b"))
			Expect(helper.ToSnakeCase("A1A2A3")).To(Equal("a_1_a_2_a_3"))
			Expect(helper.ToSnakeCase("A1 A2 A3")).To(Equal("a_1_a_2_a_3"))
			Expect(helper.ToSnakeCase("AB1AB2AB3")).To(Equal("ab_1_ab_2_ab_3"))
			Expect(helper.ToSnakeCase("AB1 AB2 AB3")).To(Equal("ab_1_ab_2_ab_3"))
			Expect(helper.ToSnakeCase("some string")).To(Equal("some_string"))
			Expect(helper.ToSnakeCase(" some string")).To(Equal("some_string"))
		})
	})
})
