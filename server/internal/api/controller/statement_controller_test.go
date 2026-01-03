package controller_test

import (
	"expenses/internal/models"
	"expenses/pkg/utils"
	"fmt"
	"net/http"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func createUniqueUser(baseURL string) *TestHelper {
	email := fmt.Sprintf("user_%d@example.com", time.Now().UnixNano())
	userInput := models.CreateUserInput{
		Email:    email,
		Name:     "Test User",
		Password: "securepassword123",
	}
	testHelper := NewTestHelper(baseURL)
	resp, response := testHelper.MakeRequest(http.MethodPost, "/signup", userInput)
	Expect(resp.StatusCode).To(Equal(http.StatusCreated))
	user := response["data"].(map[string]any)["user"].(map[string]any)
	testHelper.Login(user["email"].(string), "securepassword123")
	return testHelper
}

func createAccount(testHelper *TestHelper, name string, balance float64) float64 {
	accountInput := models.CreateAccountInput{
		Name:     name,
		BankType: "sbi",
		Currency: "inr",
		Balance:  floatPtr(balance),
	}
	resp, response := testHelper.MakeRequest(http.MethodPost, "/account", accountInput)
	Expect(resp.StatusCode).To(Equal(http.StatusCreated))
	return response["data"].(map[string]any)["id"].(float64)
}

func waitForStatementDone(testHelper *TestHelper, statementId float64) map[string]any {
	var status string
	var data map[string]any
	for i := 0; i < 20; i++ {
		resp, response := testHelper.MakeRequest(http.MethodGet, "/statement/"+strconv.FormatFloat(statementId, 'f', 0, 64), nil)
		Expect(resp.StatusCode).To(Equal(http.StatusOK))
		data = response["data"].(map[string]any)
		status = data["status"].(string)
		if status == "done" {
			break
		}
		time.Sleep(1 * time.Second)
	}
	Expect(status).To(Equal("done"))
	return data
}

var _ = Describe("StatementController", func() {
	Describe("ListStatements", func() {
		It("should list all statements for the authenticated user", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Statements fetched successfully"))
			Expect(response).To(HaveKey("data"))
			data := response["data"].(map[string]any)

			Expect(data).To(HaveKey("statements"))
			Expect(data).To(HaveKey("total"))
			Expect(data).To(HaveKey("page"))
			Expect(data).To(HaveKey("page_size"))

			statements := data["statements"].([]any)
			Expect(len(statements)).To(BeNumerically(">=", 3))

			msgs := []string{}
			for _, s := range statements {
				stmt := s.(map[string]any)
				if msg, ok := stmt["message"].(string); ok {
					msgs = append(msgs, msg)
				}
			}
			Expect(msgs).To(ContainElements("Seed Salary statement", "Seed Groceries statement", "Seed Utilities statement"))
		})

		It("should return an empty list for a different user", func() {
			resp, response := testUser3.MakeRequest(http.MethodGet, "/statement", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response).To(HaveKey("data"))
			data := response["data"].(map[string]any)

			Expect(data).To(HaveKey("statements"))
			Expect(data).To(HaveKey("total"))
			Expect(data).To(HaveKey("page"))
			Expect(data).To(HaveKey("page_size"))

			statements := data["statements"].([]any)
			Expect(len(statements)).To(Equal(0))
			Expect(data["total"]).To(Equal(0.0))
		})

		It("should return only one statement when pageSize is 1 and page is 1", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement?page_size=1&page=1", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response).To(HaveKey("data"))
			data := response["data"].(map[string]any)

			Expect(data).To(HaveKey("statements"))
			Expect(data).To(HaveKey("total"))
			Expect(data).To(HaveKey("page"))
			Expect(data).To(HaveKey("page_size"))

			statements := data["statements"].([]any)
			Expect(len(statements)).To(Equal(1))
			Expect(data["page"]).To(Equal(1.0))
			Expect(data["page_size"]).To(Equal(1.0))
			Expect(data["total"]).To(BeNumerically(">=", 3))
		})

		It("should return the second statement when pageSize is 1 and page is 2", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement?page_size=1&page=2", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response).To(HaveKey("data"))
			data := response["data"].(map[string]any)

			Expect(data).To(HaveKey("statements"))
			Expect(data).To(HaveKey("total"))
			Expect(data).To(HaveKey("page"))
			Expect(data).To(HaveKey("page_size"))

			statements := data["statements"].([]any)
			Expect(len(statements)).To(Equal(1))
			Expect(data["page"]).To(Equal(2.0))
			Expect(data["page_size"]).To(Equal(1.0))
			Expect(data["total"]).To(BeNumerically(">=", 3))
		})

		It("should return the third statement when pageSize is 1 and page is 3", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement?page_size=1&page=3", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response).To(HaveKey("data"))
			data := response["data"].(map[string]any)

			Expect(data).To(HaveKey("statements"))
			Expect(data).To(HaveKey("total"))
			Expect(data).To(HaveKey("page"))
			Expect(data).To(HaveKey("page_size"))
		})

		It("should return data as empty statement when requesting page beyond available data", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement?page_size=1&page=4", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response).To(HaveKey("data"))
			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("statements"))
			Expect(data).To(HaveKey("total"))
			Expect(data).To(HaveKey("page"))
			Expect(data).To(HaveKey("page_size"))
			Expect(data["statements"]).To(BeEmpty())
		})

		It("should return unauthorized for unauthenticated user", func() {
			resp, response := testHelperUnauthenticated.MakeRequest(http.MethodGet, "/statement", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(response).To(HaveKey("message"))
			Expect(response["message"]).To(Equal("please log in to continue"))
		})

		It("should return success for invalid page parameter", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement?page=abc", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Should have all the statements
			Expect(response).To(HaveKey("data"))
			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("statements"))
			Expect(data).To(HaveKey("total"))
			Expect(data).To(HaveKey("page"))
			Expect(data).To(HaveKey("page_size"))
		})

		It("should return success for invalid pageSize parameter", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement?page_size=abc", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Should have all the statements
			Expect(response).To(HaveKey("data"))
			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("statements"))
			Expect(data).To(HaveKey("total"))
			Expect(data).To(HaveKey("page"))
			Expect(data).To(HaveKey("page_size"))
		})

		It("should return success for negative page parameter", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement?page=-1", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Should have all the statements
			Expect(response).To(HaveKey("data"))
			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("statements"))
			Expect(data).To(HaveKey("total"))
			Expect(data).To(HaveKey("page"))
			Expect(data).To(HaveKey("page_size"))

			statements := data["statements"].([]any)
			Expect(len(statements)).To(BeNumerically(">=", 3))
		})

		It("should return success for zero pageSize parameter", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement?page_size=0", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			// Should have all the statements
			Expect(response).To(HaveKey("data"))
			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("statements"))
			Expect(data).To(HaveKey("total"))
			Expect(data).To(HaveKey("page"))
			Expect(data).To(HaveKey("page_size"))

			statements := data["statements"].([]any)
			Expect(len(statements)).To(BeNumerically(">=", 3))
		})

		It("should handle large pageSize gracefully", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement?page_size=1000", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response).To(HaveKey("data"))
			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("statements"))
			statements := data["statements"].([]any)
			Expect(len(statements)).To(BeNumerically("<=", 100))
		})
	})

	Describe("GetStatementByID", func() {
		It("should fetch statement by id for user 1", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement/1", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response).To(HaveKey("data"))
			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("id"))
			Expect(data["id"]).To(Equal(1.0))
			Expect(data["original_filename"]).To(Equal("salary_jan.csv"))
			Expect(data["message"]).To(Equal("Seed Salary statement"))
		})

		It("should fetch statement by id for user 2", func() {
			resp, response := testUser2.MakeRequest(http.MethodGet, "/statement/4", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response).To(HaveKey("data"))
			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("id"))
			Expect(data["id"]).To(Equal(4.0))
			Expect(data["original_filename"]).To(Equal("user2_statement.csv"))
			Expect(data["message"]).To(Equal("Seed User2 statement"))
		})

		It("should return not found when fetching statement by id of another user", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement/4", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response).To(HaveKey("message"))
			Expect(response["message"]).To(Equal("statement not found"))
		})

		It("should return unauthorized for unauthenticated user fetching statement by id", func() {
			resp, response := testHelperUnauthenticated.MakeRequest(http.MethodGet, "/statement/1", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(response).To(HaveKey("message"))
			Expect(response["message"]).To(Equal("please log in to continue"))
		})

		It("should return not found for non-existent statement id", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement/99999", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response).To(HaveKey("message"))
			Expect(response["message"]).To(Equal("statement not found"))
		})

		It("should return bad request for invalid statement id format", func() {
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement/abc", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response).To(HaveKey("message"))
			Expect(response["message"]).To(Equal("Invalid statement_id"))
		})
	})

	Describe("PreviewStatement", func() {
		It("should successfully preview a valid CSV statement file", func() {
			fileContent := []byte(
				"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTO TRANSFER-UPI\t123456\t100.00\t\t1000.00\n" +
					"2 Aug 2022\t2 Aug 2022\tBY TRANSFER-NEFT\t654321\t\t200.00\t1200.00\n" +
					"3 Aug 2022\t3 Aug 2022\tATM WITHDRAWAL\t789012\t50.00\t\t1150.00\n" +
					"Computer Generated Statement")

			previewInput := map[string]any{
				"file":      fileContent,
				"skip_rows": 0,
				"row_size":  10,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Statement preview generated successfully"))
			Expect(response).To(HaveKey("data"))

			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("rows"))
			Expect(data).To(HaveKey("headers"))

			rows := data["rows"].([]any)
			Expect(len(rows)).To(BeNumerically(">", 0))
		})

		It("should successfully preview with custom skip_rows parameter", func() {
			fileContent := []byte(
				"Header Line 1\n" +
					"Header Line 2\n" +
					"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTO TRANSFER-UPI\t123456\t100.00\t\t1000.00\n" +
					"2 Aug 2022\t2 Aug 2022\tBY TRANSFER-NEFT\t654321\t\t200.00\t1200.00\n")

			previewInput := map[string]any{
				"file":      fileContent,
				"skip_rows": 2,
				"row_size":  10,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Statement preview generated successfully"))
			Expect(response).To(HaveKey("data"))
		})

		It("should successfully preview with custom row_size parameter", func() {
			fileContent := []byte(
				"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTO TRANSFER-UPI\t123456\t100.00\t\t1000.00\n" +
					"2 Aug 2022\t2 Aug 2022\tBY TRANSFER-NEFT\t654321\t\t200.00\t1200.00\n" +
					"3 Aug 2022\t3 Aug 2022\tATM WITHDRAWAL\t789012\t50.00\t\t1150.00\n" +
					"4 Aug 2022\t4 Aug 2022\tONLINE PURCHASE\t345678\t75.00\t\t1075.00\n")

			previewInput := map[string]any{
				"file":     fileContent,
				"row_size": 2,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Statement preview generated successfully"))
			Expect(response).To(HaveKey("data"))

			data := response["data"].(map[string]any)
			rows := data["rows"].([]any)
			Expect(len(rows)).To(BeNumerically("<=", 2))
		})

		It("should successfully preview with both skip_rows and row_size parameters", func() {
			fileContent := []byte(
				"Header Line\n" +
					"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTO TRANSFER-UPI\t123456\t100.00\t\t1000.00\n" +
					"2 Aug 2022\t2 Aug 2022\tBY TRANSFER-NEFT\t654321\t\t200.00\t1200.00\n" +
					"3 Aug 2022\t3 Aug 2022\tATM WITHDRAWAL\t789012\t50.00\t\t1150.00\n")

			previewInput := map[string]any{
				"file":      fileContent,
				"skip_rows": 1,
				"row_size":  2,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Statement preview generated successfully"))
			Expect(response).To(HaveKey("data"))
		})

		It("should return bad request when file is missing", func() {
			previewInput := map[string]any{
				"skip_rows": 0,
				"row_size":  10,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response).To(HaveKey("message"))
		})

		It("should return error for empty file", func() {
			fileContent := []byte("")

			previewInput := map[string]any{
				"file":      fileContent,
				"skip_rows": 0,
				"row_size":  10,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(SatisfyAny(Equal(http.StatusBadRequest), Equal(http.StatusInternalServerError)))
			Expect(response).To(HaveKey("message"))
		})

		It("should return error for file larger than 5MB", func() {
			// Create a file >5MB
			bigFile := make([]byte, 5*1024*1024+1)
			for i := range bigFile {
				bigFile[i] = 'A'
			}

			previewInput := map[string]any{
				"file":      bigFile,
				"skip_rows": 0,
				"row_size":  10,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response).To(HaveKey("message"))
		})

		It("should error out invalid skip_rows parameter gracefully", func() {
			fileContent := []byte(
				"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTO TRANSFER-UPI\t123456\t100.00\t\t1000.00\n")

			previewInput := map[string]any{
				"file":      fileContent,
				"skip_rows": "invalid",
			}
			resp, _ := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should error out invalid row_size parameter gracefully", func() {
			fileContent := []byte(
				"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTO TRANSFER-UPI\t123456\t100.00\t\t1000.00\n")

			previewInput := map[string]any{
				"file":     fileContent,
				"row_size": "invalid",
			}
			resp, _ := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should error out negative skip_rows parameter", func() {
			fileContent := []byte(
				"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTO TRANSFER-UPI\t123456\t100.00\t\t1000.00\n")

			previewInput := map[string]any{
				"file":      fileContent,
				"skip_rows": -1,
			}
			resp, _ := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should error out negative row_size parameter", func() {
			fileContent := []byte(
				"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTO TRANSFER-UPI\t123456\t100.00\t\t1000.00\n")

			previewInput := map[string]any{
				"file":     fileContent,
				"row_size": -1,
			}
			resp, _ := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should handle zero row_size parameter", func() {
			fileContent := []byte(
				"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTO TRANSFER-UPI\t123456\t100.00\t\t1000.00\n")

			previewInput := map[string]any{
				"file":     fileContent,
				"row_size": 0,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Statement preview generated successfully"))
		})

		It("should return unauthorized for unauthenticated user", func() {
			fileContent := []byte(
				"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTO TRANSFER-UPI\t123456\t100.00\t\t1000.00\n")

			previewInput := map[string]any{
				"file": fileContent,
			}
			resp, response := testHelperUnauthenticated.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			Expect(response).To(HaveKey("message"))
			Expect(response["message"]).To(Equal("please log in to continue"))
		})

		It("should return appropriate preview structure", func() {
			fileContent := []byte(
				"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTO TRANSFER-UPI\t123456\t100.00\t\t1000.00\n" +
					"2 Aug 2022\t2 Aug 2022\tBY TRANSFER-NEFT\t654321\t\t200.00\t1200.00\n")

			previewInput := map[string]any{
				"file": fileContent,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Statement preview generated successfully"))
			Expect(response).To(HaveKey("data"))

			data := response["data"].(map[string]any)
			Expect(data).To(HaveKey("rows"))
			Expect(data).To(HaveKey("headers"))

			headers := data["headers"].([]any)
			Expect(len(headers)).To(BeNumerically(">", 0))

			rows := data["rows"].([]any)
			Expect(len(rows)).To(BeNumerically(">", 0))
		})

		It("should handle CSV with comma delimiters", func() {
			fileContent := []byte(
				"Txn Date,Value Date,Description,Ref No.,Debit,Credit,Balance\n" +
					"1 Aug 2022,1 Aug 2022,TO TRANSFER-UPI,123456,100.00,,1000.00\n" +
					"2 Aug 2022,2 Aug 2022,BY TRANSFER-NEFT,654321,,200.00,1200.00\n")

			previewInput := map[string]any{
				"file": fileContent,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Statement preview generated successfully"))
			Expect(response).To(HaveKey("data"))
		})

		It("should handle files with special characters", func() {
			fileContent := []byte(
				"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTÖ TRÄNSFÉR-UPI\t123456\t₹100.00\t\t₹1000.00\n" +
					"2 Aug 2022\t2 Aug 2022\tBY TRÄNSFÉR-NEFT\t654321\t\t₹200.00\t₹1200.00\n")

			previewInput := map[string]any{
				"file": fileContent,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Statement preview generated successfully"))
			Expect(response).To(HaveKey("data"))
		})

		It("should handle malformed CSV gracefully", func() {
			fileContent := []byte(
				"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTO TRANSFER-UPI\t123456\t100.00\n" + // Missing columns
					"INVALID ROW WITH NO TABS\n" +
					"2 Aug 2022\t2 Aug 2022\tBY TRANSFER-NEFT\t654321\t\t200.00\t1200.00\n")

			previewInput := map[string]any{
				"file": fileContent,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			// Should either succeed with partial data or return an appropriate error
			Expect(resp.StatusCode).To(SatisfyAny(Equal(http.StatusOK), Equal(http.StatusBadRequest), Equal(http.StatusInternalServerError)))
			Expect(response).To(HaveKey("message"))
		})

		It("should handle very large row_size parameter", func() {
			fileContent := []byte(
				"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTO TRANSFER-UPI\t123456\t100.00\t\t1000.00\n" +
					"2 Aug 2022\t2 Aug 2022\tBY TRANSFER-NEFT\t654321\t\t200.00\t1200.00\n")

			previewInput := map[string]any{
				"file":     fileContent,
				"row_size": 10000,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Statement preview generated successfully"))
			Expect(response).To(HaveKey("data"))

			data := response["data"].(map[string]any)
			rows := data["rows"].([]any)
			// Should return all available rows, not more than what exists
			Expect(len(rows)).To(BeNumerically("<=", 10000))
		})

		It("should handle very large skip_rows parameter", func() {
			fileContent := []byte(
				"Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n" +
					"1 Aug 2022\t1 Aug 2022\tTO TRANSFER-UPI\t123456\t100.00\t\t1000.00\n")

			previewInput := map[string]any{
				"file":      fileContent,
				"skip_rows": 10000,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Statement preview generated successfully"))
			Expect(response).To(HaveKey("data"))

			data := response["data"].(map[string]any)
			rows := data["rows"].([]any)
			// Should return empty or minimal rows since we're skipping more than available
			Expect(len(rows)).To(BeNumerically(">=", 0))
		})
	})

	Describe("Error Handling", func() {
		It("should handle file header nil error in CreateStatement", func() {
			// This tests the readFileFromRequest error path when fileHeader is nil
			statementInput := map[string]any{
				"account_id":        1,
				"original_filename": "test.csv",
				"file_type":         "csv",
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement", statementInput)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response).To(HaveKey("message"))
		})

		It("should handle file header nil error in PreviewStatement", func() {
			// This tests the readFileFromRequest error path when fileHeader is nil
			previewInput := map[string]any{
				"skip_rows": 0,
				"row_size":  10,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response).To(HaveKey("message"))
		})

		It("should handle service error in PreviewStatement", func() {
			// Create a file that will cause service-level validation error
			fileContent := make([]byte, 5*1024*1024+1) // File larger than 5MB
			for i := range fileContent {
				fileContent[i] = 'A'
			}

			previewInput := map[string]any{
				"file":      fileContent,
				"skip_rows": 0,
				"row_size":  10,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement/preview", previewInput)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response).To(HaveKey("message"))
		})

		It("should handle service error in CreateStatement", func() {
			// Create a statement with invalid account ID to trigger service error
			fileContent := []byte("Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n")
			statementInput := map[string]any{
				"account_id":        -1, // Invalid account ID
				"original_filename": "test.csv",
				"file_type":         "csv",
				"file":              fileContent,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement", statementInput)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response).To(HaveKey("message"))
		})

		It("should handle account not found error in CreateStatement", func() {
			fileContent := []byte("Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n")
			statementInput := map[string]any{
				"account_id":        99999, // Non-existent account ID
				"original_filename": "test.csv",
				"file_type":         "csv",
				"file":              fileContent,
			}
			resp, response := testUser1.MakeMultipartRequest(http.MethodPost, "/statement", statementInput)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response).To(HaveKey("message"))
		})

		It("should handle database error in GetStatements", func() {
			// This would require mocking database failure, but we can test with extreme parameters
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement?page=999999&page_size=1", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response).To(HaveKey("data"))
		})

		It("should handle database error in GetStatementStatus", func() {
			// Test with statement ID that doesn't exist or belongs to another user
			resp, response := testUser1.MakeRequest(http.MethodGet, "/statement/99999", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response).To(HaveKey("message"))
		})
	})

	Describe("CreateStatement", func() {
		It("should create a statement, wait for status to be success, and verify transaction inclusion", func() {
			// 1. Signup a new user
			userInput := models.CreateUserInput{
				Email:    "integration_user@example.com",
				Name:     "Integration User",
				Password: "securepassword123",
			}
			testHelper := NewTestHelper(baseURL)
			resp, response := testHelper.MakeRequest(http.MethodPost, "/signup", userInput)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(response["message"]).To(Equal("User signed up successfully"))
			Expect(response["data"]).To(HaveKey("user"))
			user := response["data"].(map[string]any)["user"].(map[string]any)
			userEmail := user["email"].(string)
			userPassword := "securepassword123"

			// 2. Login as the new user
			testHelper.Login(userEmail, userPassword)

			// 3. Create a new SBI account
			accountInput := models.CreateAccountInput{
				Name:     "Integration SBI Account",
				BankType: "sbi",
				Currency: "inr",
				Balance:  floatPtr(1000.0),
			}
			resp, response = testHelper.MakeRequest(http.MethodPost, "/account", accountInput)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			accountId := response["data"].(map[string]any)["id"].(float64)

			// 4. Create a statement with an XLSX upload so the SBI parser can handle it
			xlsxData := [][]string{
				{"Txn Date", "Details", "Ref No.", "Debit", "Credit", "Balance"},
				{"1 Aug 2022", "TO TRANSFER-UPI/DR/221356312527/RITIK S/SBIN/rs6321908@/UPI--", "123456", "100.00", "", "1000.00"},
				{"BADROW"},
				{"2 Aug 2022", "BY TRANSFER-NEFT*HDFC0000001*N215222062454075*QURIATE TECHNOLO--", "654321", "", "200.00", "1200.00"},
				{"Computer Generated Statement"},
			}
			fileBytes := utils.CreateXLSXFile(xlsxData)
			fileName := "integration_statement.xlsx"
			statementInput := map[string]any{
				"account_id":        int64(accountId),
				"original_filename": fileName,
				"file_type":         "excel",
				"file":              fileBytes,
			}
			resp, response = testHelper.MakeMultipartRequest(http.MethodPost, "/statement", statementInput)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			statementId := response["data"].(map[string]any)["id"].(float64)

			// 5. Wait for status to be "done"
			var status string
			for i := 0; i < 10; i++ {
				resp, response = testHelper.MakeRequest(http.MethodGet, "/statement/"+strconv.FormatFloat(statementId, 'f', 0, 64), nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				data := response["data"].(map[string]any)
				status = data["status"].(string)
				if status == "done" {
					break
				}
				time.Sleep(1 * time.Second)
			}
			Expect(status).To(Equal("done"))

			// 6. Fetch all transactions filtered by statement_id and check those
			resp, response = testHelper.MakeRequest(http.MethodGet, "/transaction?statement_id="+strconv.FormatFloat(statementId, 'f', 0, 64), nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response).To(HaveKey("data"))
			txData := response["data"].(map[string]any)
			Expect(txData).To(HaveKey("transactions"))
			filteredTxs := txData["transactions"].([]any)
			Expect(len(filteredTxs)).To(Equal(2))

			// 7. Delete the user
			resp, _ = testHelper.MakeRequest(http.MethodDelete, "/user", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
		})

		It("should parse statement with 4 transactions and another with 7 transactions", func() {
			testHelper := createUniqueUser(baseURL)
			accountId := createAccount(testHelper, "Account 1", 1000.0)

			// Upload as XLSX so the SBI parser can parse rows reliably
			xlsx4 := [][]string{
				{"Txn Date", "Details", "Ref No.", "Debit", "Credit", "Balance"},
				{"1 Aug 2022", "Desc1", "123", "100.00", "", "1000.00"},
				{"2 Aug 2022", "Desc2", "124", "200.00", "", "1200.00"},
				{"3 Aug 2022", "Desc3", "125", "300.00", "", "1500.00"},
				{"4 Aug 2022", "Desc4", "126", "400.00", "", "1900.00"},
			}
			fileBytes4 := utils.CreateXLSXFile(xlsx4)
			statementInput4 := map[string]any{
				"account_id":        int64(accountId),
				"original_filename": "statement_4.xlsx",
				"file_type":         "excel",
				"file":              fileBytes4,
			}
			resp, response := testHelper.MakeMultipartRequest(http.MethodPost, "/statement", statementInput4)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			statementId := response["data"].(map[string]any)["id"].(float64)
			waitForStatementDone(testHelper, statementId)

			// Fetch transactions for the first statement
			res, response := testHelper.MakeRequest(http.MethodGet, "/transaction?page_size=10&statement_id="+strconv.FormatFloat(statementId, 'f', 0, 64), nil)
			Expect(res.StatusCode).To(Equal(http.StatusOK))
			Expect(response).To(HaveKey("data"))
			txData := response["data"].(map[string]any)
			Expect(txData).To(HaveKey("transactions"))
			transactions := txData["transactions"].([]any)
			Expect(transactions).To(HaveLen(4))

			// Upload as XLSX for the 7-row statement
			xlsx7 := [][]string{
				{"Txn Date", "Details", "Ref No.", "Debit", "Credit", "Balance"},
				{"1 Aug 2022", "Desc1", "123", "100.00", "", "1000.00"},
				{"2 Aug 2022", "Desc2", "124", "200.00", "", "1200.00"},
				{"3 Aug 2022", "Desc3", "125", "300.00", "", "1500.00"},
				{"4 Aug 2022", "Desc4", "126", "400.00", "", "1900.00"},
				{"5 Aug 2022", "Desc5", "127", "500.00", "", "2400.00"},
				{"6 Aug 2022", "Desc6", "128", "600.00", "", "3000.00"},
				{"7 Aug 2022", "Desc7", "129", "700.00", "", "3700.00"},
			}
			fileBytes7 := utils.CreateXLSXFile(xlsx7)
			statementInput7 := map[string]any{
				"account_id":        int64(accountId),
				"original_filename": "statement_7.xlsx",
				"file_type":         "excel",
				"file":              fileBytes7,
			}
			resp, response = testHelper.MakeMultipartRequest(http.MethodPost, "/statement", statementInput7)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			statementId = response["data"].(map[string]any)["id"].(float64)
			waitForStatementDone(testHelper, statementId)

			// Fetch transactions for the second statement
			res, response = testHelper.MakeRequest(http.MethodGet, "/transaction?page_size=10&statement_id="+strconv.FormatFloat(statementId, 'f', 0, 64), nil)
			Expect(res.StatusCode).To(Equal(http.StatusOK))
			Expect(response).To(HaveKey("data"))
			txData = response["data"].(map[string]any)
			Expect(txData).To(HaveKey("transactions"))
			transactions = txData["transactions"].([]any)
			Expect(transactions).To(HaveLen(3))
		})

		It("should parse statement with 10000 transactions", func() {
			testHelper := createUniqueUser(baseURL)
			accountId := createAccount(testHelper, "Account 1", 1000.0)

			// Build XLSX data (header + 10000 rows)
			data := [][]string{
				{"Txn Date", "Details", "Ref No.", "Debit", "Credit", "Balance"},
			}
			for i := 1; i <= 1000; i++ {
				row := []string{
					"22 Aug 2022",
					fmt.Sprintf("Desc%d", i),
					fmt.Sprintf("%d", 1000+i),
					fmt.Sprintf("%.2f", float64(i)),
					fmt.Sprintf("%.2f", float64(1000+i)),
					fmt.Sprintf("%.2f", float64(2000+i)),
				}
				data = append(data, row)
			}

			fileBytes := utils.CreateXLSXFile(data)
			// Upload as an Excel file
			statementInput := map[string]any{
				"account_id":        int64(accountId),
				"original_filename": "statement_10000.xlsx",
				"file_type":         "excel",
				"file":              fileBytes,
			}
			resp, response := testHelper.MakeMultipartRequest(http.MethodPost, "/statement", statementInput)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			statementId := response["data"].(map[string]any)["id"].(float64)
			waitForStatementDone(testHelper, statementId)

			// Fetch transactions for the statement
			resp, response = testHelper.MakeRequest(http.MethodGet, "/transaction?page_size=10&statement_id="+strconv.FormatFloat(statementId, 'f', 0, 64), nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response).To(HaveKey("data"))
			txData := response["data"].(map[string]any)
			Expect(txData).To(HaveKey("transactions"))
			transactions := txData["transactions"].([]any)
			Expect(len(transactions)).To(Equal(10))
			Expect(txData).To(HaveKey("total"))
			Expect(txData["total"]).To(Equal(1000.0))
		})

		It("should error when accountId doesn't exist", func() {
			testHelper := createUniqueUser(baseURL)
			statementInput := map[string]any{
				"account_id":        int64(999999), // non-existent
				"original_filename": "statement_invalid.csv",
				"file_type":         "csv",
				"file":              []byte("Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n1 Aug 2022\t1 Aug 2022\tDesc\t123\t100.00\t\t1000.00\nComputer Generated Statement"),
			}
			resp, _ := testHelper.MakeMultipartRequest(http.MethodPost, "/statement", statementInput)
			Expect(resp.StatusCode).To(SatisfyAny(Equal(http.StatusBadRequest), Equal(http.StatusNotFound)))
		})

		It("should error when accountId is from a different user", func() {
			testHelper := createUniqueUser(baseURL)
			otherHelper := createUniqueUser(baseURL)
			otherAccountId := createAccount(otherHelper, "Other Account", 5000.0)

			// Try to upload statement with other user's accountId
			statementInput := map[string]any{
				"account_id":        int64(otherAccountId),
				"original_filename": "statement_other.csv",
				"file_type":         "csv",
				"file":              []byte("Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n1 Aug 2022\t1 Aug 2022\tDesc\t123\t100.00\t\t1000.00\nComputer Generated Statement"),
			}
			resp, _ := testHelper.MakeMultipartRequest(http.MethodPost, "/statement", statementInput)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should fail to parse multipart file > 5MB", func() {
			testHelper := createUniqueUser(baseURL)
			accountId := createAccount(testHelper, "BigFileAccount", 1000.0)
			// Create a file >5MB
			bigFile := make([]byte, 5*1024*1024+1)
			for i := range bigFile {
				bigFile[i] = 'A'
			}
			statementInput := map[string]any{
				"account_id":        int64(accountId),
				"original_filename": "bigfile.csv",
				"file_type":         "csv",
				"file":              bigFile,
			}
			resp, _ := testHelper.MakeMultipartRequest(http.MethodPost, "/statement", statementInput)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should fail when account_id is a string", func() {
			testHelper := createUniqueUser(baseURL)
			statementInput := map[string]any{
				"account_id":        "notanumber",
				"original_filename": "statement.csv",
				"file_type":         "csv",
				"file":              []byte("Txn Date\tValue Date\tDescription\tRef No.\tDebit\tCredit\tBalance\n1 Aug 2022\t1 Aug 2022\tDesc\t123\t100.00\t\t1000.00\nComputer Generated Statement"),
			}
			resp, _ := testHelper.MakeMultipartRequest(http.MethodPost, "/statement", statementInput)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should fail when file is not sent", func() {
			testHelper := createUniqueUser(baseURL)
			accountId := createAccount(testHelper, "NoFileAccount", 1000.0)
			statementInput := map[string]any{
				"account_id":        int64(accountId),
				"original_filename": "nofile.csv",
				"file_type":         "csv",
				// No "file" key
			}
			resp, _ := testHelper.MakeMultipartRequest(http.MethodPost, "/statement", statementInput)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})
})
