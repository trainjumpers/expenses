package controller_test

import (
	"bytes"
	"encoding/json"
	"expenses/internal/models"
	"net/http"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CategoryController", func() {
	Describe("CreateCategory", func() {
		It("should create a category successfully with icon", func() {
			input := models.CreateCategoryInput{
				Name: "Food with icon",
				Icon: "burger-icon",
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/category", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Category created successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]interface{})["name"]).To(Equal("Food with icon"))
			Expect(response["data"].(map[string]interface{})["icon"]).To(Equal("burger-icon"))
		})

		It("should create a category successfully with empty icon", func() {
			input := models.CreateCategoryInput{
				Name: "Entertainment without icon",
				Icon: "",
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/category", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Category created successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]interface{})["name"]).To(Equal("Entertainment without icon"))
			Expect(response["data"].(map[string]interface{})["icon"]).To(BeNil())
		})

		It("should return error when category name already exists for same user", func() {
			input := models.CreateCategoryInput{
				Name: "Transportation check existing",
				Icon: "car-icon",
			}
			body, _ := json.Marshal(input)

			// Create first category
			req, err := http.NewRequest(http.MethodPost, baseURL+"/category", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			// Try to create same category again
			req, err = http.NewRequest(http.MethodPost, baseURL+"/category", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusConflict))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("category with this name already exists for this user"))
		})

		It("should allow same category name for different users", func() {
			input := models.CreateCategoryInput{
				Name: "Shoppingfor different user",
				Icon: "shopping-icon",
			}
			body, _ := json.Marshal(input)

			// Create category for first user
			req, err := http.NewRequest(http.MethodPost, baseURL+"/category", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			// Create category with same name for second user
			req, err = http.NewRequest(http.MethodPost, baseURL+"/category", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken1)
			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		})

		It("should return error for invalid authorization", func() {
			input := models.CreateCategoryInput{
				Name: "Invalid Auth Category",
				Icon: "error-icon",
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/category", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer invalid-token")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})

		It("should return error for missing name", func() {
			input := models.CreateCategoryInput{
				Icon: "burger-icon",
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/category", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for invalid JSON", func() {
			req, err := http.NewRequest(http.MethodPost, baseURL+"/category", bytes.NewBuffer([]byte("invalid json")))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})
	})

	Describe("ListCategories", func() {
		It("should list categories for authenticated user", func() {
			req, err := http.NewRequest(http.MethodGet, baseURL+"/category", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Categories retrieved successfully"))
			Expect(response["data"]).To(BeAssignableToTypeOf([]interface{}{}))

			categories := response["data"].([]interface{})
			Expect(len(categories)).To(BeNumerically(">=", 5))
		})

		It("should return empty list for user with no categories", func() {
			req, err := http.NewRequest(http.MethodGet, baseURL+"/category", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken2)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Categories retrieved successfully"))
			Expect(response["data"]).To(BeNil())
		})

		It("should return error for invalid authorization", func() {
			req, err := http.NewRequest(http.MethodGet, baseURL+"/category", nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer invalid-token")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("GetCategory", func() {
		var categoryId int64 = 1 // From seed data
		It("should get category by id successfully", func() {
			url := baseURL + "/category/" + strconv.FormatInt(categoryId, 10)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Category retrieved successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]interface{})["name"]).To(Equal("Food"))
		})

		It("should return error for non-existent category id", func() {
			url := baseURL + "/category/9999"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("category not found"))
		})

		It("should return error when accessing category of different user", func() {
			url := baseURL + "/category/" + strconv.FormatInt(categoryId, 10)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken1)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("category not found"))
		})

		It("should return error for invalid category id", func() {
			url := baseURL + "/category/invalid"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("invalid category id"))
		})

		It("should return error for invalid authorization", func() {
			url := baseURL + "/category/" + strconv.FormatInt(categoryId, 10)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer invalid-token")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("UpdateCategory", func() {
		var categoryId int64 = 2

		It("should update category name successfully", func() {
			update := models.UpdateCategoryInput{Name: "Updated Category Name"}
			body, _ := json.Marshal(update)
			url := baseURL + "/category/" + strconv.FormatInt(categoryId, 10)
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Category updated successfully"))
			Expect(response["data"].(map[string]interface{})["name"]).To(Equal("Updated Category Name"))
		})

		It("should update category icon successfully", func() {
			newIcon := "new-icon"
			update := models.UpdateCategoryInput{Icon: &newIcon}
			body, _ := json.Marshal(update)
			url := baseURL + "/category/" + strconv.FormatInt(categoryId, 10)
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Category updated successfully"))
			Expect(response["data"].(map[string]interface{})["icon"]).To(Equal(newIcon))
		})

		It("should update both name and icon successfully", func() {
			newIcon := "updated-icon"
			update := models.UpdateCategoryInput{
				Name: "Complete Update",
				Icon: &newIcon,
			}
			body, _ := json.Marshal(update)
			url := baseURL + "/category/" + strconv.FormatInt(categoryId, 10)
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("Category updated successfully"))
			Expect(response["data"].(map[string]interface{})["name"]).To(Equal("Complete Update"))
			Expect(response["data"].(map[string]interface{})["icon"]).To(Equal(newIcon))
		})

		It("should return error for non-existent category id", func() {
			update := models.UpdateCategoryInput{Name: "Updated Name"}
			body, _ := json.Marshal(update)
			url := baseURL + "/category/9999"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("category not found"))
		})

		It("should return error when updating category of different user", func() {
			update := models.UpdateCategoryInput{Name: "Updated Name"}
			body, _ := json.Marshal(update)
			url := baseURL + "/category/" + strconv.FormatInt(categoryId, 10)
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken1)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("category not found"))
		})

		It("should return error when updating to duplicate name for same user", func() {
			// Create another category
			input := models.CreateCategoryInput{
				Name: "Unique Category",
				Icon: "unique-icon",
			}
			body, _ := json.Marshal(input)
			req, err := http.NewRequest(http.MethodPost, baseURL+"/category", bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)
			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			// Try to update first category to have same name
			update := models.UpdateCategoryInput{Name: "Unique Category"}
			body, _ = json.Marshal(update)
			url := baseURL + "/category/" + strconv.FormatInt(categoryId, 10)
			req, err = http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusConflict))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("category with this name already exists for this user"))
		})

		It("should return error for invalid category id", func() {
			update := models.UpdateCategoryInput{Name: "Updated Name"}
			body, _ := json.Marshal(update)
			url := baseURL + "/category/invalid"
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("invalid category id"))
		})

		It("should return error for invalid JSON", func() {
			url := baseURL + "/category/" + strconv.FormatInt(categoryId, 10)
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer([]byte("invalid json")))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for invalid authorization", func() {
			update := models.UpdateCategoryInput{Name: "Updated Name"}
			body, _ := json.Marshal(update)
			url := baseURL + "/category/" + strconv.FormatInt(categoryId, 10)
			req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(body))
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer invalid-token")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("DeleteCategory", func() {
		var categoryId int64 = 5

		It("should delete category successfully", func() {
			url := baseURL + "/category/" + strconv.FormatInt(categoryId, 10)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			// Verify category is deleted by trying to get it
			req, err = http.NewRequest(http.MethodGet, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err = client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should return error for non-existent category id", func() {
			url := baseURL + "/category/9999"
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("category not found"))
		})

		It("should return error when deleting category of different user", func() {
			url := baseURL + "/category/" + strconv.FormatInt(categoryId, 10)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken1)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("category not found"))
		})

		It("should return error for invalid category id", func() {
			url := baseURL + "/category/invalid"
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer "+accessToken)

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			response, err := decodeJSON(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			Expect(response["message"]).To(Equal("invalid category id"))
		})

		It("should return error for invalid authorization", func() {
			url := baseURL + "/category/" + strconv.FormatInt(categoryId, 10)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			Expect(err).NotTo(HaveOccurred())
			req.Header.Set("Authorization", "Bearer invalid-token")

			resp, err := client.Do(req)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})
})
