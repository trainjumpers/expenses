package controller_test

import (
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
			resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(response["message"]).To(Equal("Category created successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]any)["name"]).To(Equal("Food with icon"))
			Expect(response["data"].(map[string]any)["icon"]).To(Equal("burger-icon"))
		})

		It("should create a category successfully with empty icon", func() {
			input := models.CreateCategoryInput{
				Name: "Entertainment without icon",
				Icon: "",
			}
			resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(response["message"]).To(Equal("Category created successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]any)["name"]).To(Equal("Entertainment without icon"))
			Expect(response["data"].(map[string]any)["icon"]).To(BeNil())
		})

		It("should trim whitespace from category name and create successfully", func() {
			input := models.CreateCategoryInput{
				Name: "  Whitespace Category  ", // Name with leading and trailing whitespace
				Icon: "space-icon",
			}
			resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(response["message"]).To(Equal("Category created successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]any)["name"]).To(Equal("Whitespace Category")) // Should be trimmed
			Expect(response["data"].(map[string]any)["icon"]).To(Equal("space-icon"))
		})

		It("should trim complex whitespace characters from category name", func() {
			input := models.CreateCategoryInput{
				Name: "\t  Complex Tab Category  \n", // Name with tabs and newlines
				Icon: "tab-icon",
			}
			resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(response["message"]).To(Equal("Category created successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]any)["name"]).To(Equal("Complex Tab Category")) // Should be trimmed
			Expect(response["data"].(map[string]any)["icon"]).To(Equal("tab-icon"))
		})

		It("should trim whitespace from icon field", func() {
			input := models.CreateCategoryInput{
				Name: "Icon Whitespace Test",
				Icon: "  trimmed-icon  ", // Icon with whitespace
			}
			resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			Expect(response["message"]).To(Equal("Category created successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]any)["name"]).To(Equal("Icon Whitespace Test"))
			Expect(response["data"].(map[string]any)["icon"]).To(Equal("trimmed-icon")) // Should be trimmed
		})

		It("should return error for whitespace-only category name", func() {
			input := models.CreateCategoryInput{
				Name: "   ", // Only whitespace - will become empty after trimming
				Icon: "error-icon",
			}
			resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(ContainSubstring("required"))
		})

		It("should return error for tabs and newlines only in category name", func() {
			input := models.CreateCategoryInput{
				Name: "\t\n  \r  ", // Only various whitespace characters
				Icon: "error-icon",
			}
			resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(ContainSubstring("required"))
		})

		It("should return error when category name already exists for same user", func() {
			input := models.CreateCategoryInput{
				Name: "Transportation check existing",
				Icon: "car-icon",
			}
			// Create first category
			resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			// Try to create same category again
			resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusConflict))
			Expect(response["message"]).To(Equal("category with this name already exists for this user"))
		})

		It("should allow same category name for different users", func() {
			input := models.CreateCategoryInput{
				Name: "Shopping for different user",
				Icon: "shopping-icon",
			}
			// Create category for first user
			resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			// Create category with same name for second user
			resp, _ = testHelperUser2.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
		})

		It("should return error for invalid authorization", func() {
			input := models.CreateCategoryInput{
				Name: "Invalid Auth Category",
				Icon: "error-icon",
			}
			resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})

		It("should return error for missing name", func() {
			input := models.CreateCategoryInput{
				Icon: "burger-icon",
			}
			resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for invalid JSON", func() {
			resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/category", "invalid json")
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		Context("with malformed tokens", func() {
			It("should return unauthorized or bad request for malformed tokens on create", func() {
				input := models.CreateCategoryInput{
					Name: "Malformed Token Category",
					Icon: "icon",
				}
				checkMalformedTokens(testHelperUser1, http.MethodPost, "/category", input)
			})
			It("should return unauthorized or bad request for malformed tokens on list", func() {
				checkMalformedTokens(testHelperUser1, http.MethodGet, "/category", nil)
			})
			It("should return unauthorized or bad request for malformed tokens on get", func() {
				url := "/category/1"
				checkMalformedTokens(testHelperUser1, http.MethodGet, url, nil)
			})
			It("should return unauthorized or bad request for malformed tokens on update", func() {
				update := models.UpdateCategoryInput{Name: "Malformed Update"}
				url := "/category/1"
				checkMalformedTokens(testHelperUser1, http.MethodPatch, url, update)
			})
			It("should return unauthorized or bad request for malformed tokens on delete", func() {
				url := "/category/1"
				checkMalformedTokens(testHelperUser1, http.MethodDelete, url, nil)
			})
		})
	})

	Describe("ListCategories", func() {
		It("should list categories for authenticated user", func() {
			resp, response := testHelperUser1.MakeRequest(http.MethodGet, "/category", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Categories retrieved successfully"))
			Expect(response["data"]).To(BeAssignableToTypeOf([]any{}))
			categories := response["data"].([]any)
			Expect(len(categories)).To(BeNumerically(">=", 5))
		})

		It("should return empty list for user with no categories", func() {
			resp, response := testHelperUser3.MakeRequest(http.MethodGet, "/category", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Categories retrieved successfully"))
			Expect(len(response["data"].([]any))).To(Equal(0))
		})

		It("should return error for invalid authorization", func() {
			resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodGet, "/category", nil)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("GetCategory", func() {
		var categoryId int64 = 1 // From seed data
		It("should get category by id successfully", func() {
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Category retrieved successfully"))
			Expect(response["data"]).To(HaveKey("id"))
			Expect(response["data"].(map[string]any)["name"]).To(Equal("Food"))
		})

		It("should return error for non-existent category id", func() {
			url := "/category/9999"
			resp, response := testHelperUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(Equal("category not found"))
		})

		It("should return error when accessing category of different user", func() {
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, response := testHelperUser2.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(Equal("category not found"))
		})

		It("should return error for invalid category id", func() {
			url := "/category/invalid"
			resp, response := testHelperUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid category id"))
		})

		It("should return error for invalid authorization", func() {
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("UpdateCategory", func() {
		var categoryId int64 = 2

		It("should update category name successfully", func() {
			update := models.UpdateCategoryInput{Name: "Updated Category Name"}
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Category updated successfully"))
			Expect(response["data"].(map[string]any)["name"]).To(Equal("Updated Category Name"))
		})

		It("should trim whitespace from category name during update", func() {
			update := models.UpdateCategoryInput{Name: "  Trimmed Update Name  "} // Name with whitespace
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Category updated successfully"))
			Expect(response["data"].(map[string]any)["name"]).To(Equal("Trimmed Update Name")) // Should be trimmed
		})

		It("should trim complex whitespace from category name during update", func() {
			update := models.UpdateCategoryInput{Name: "\t  Complex Update Name  \n"} // Name with tabs and newlines
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Category updated successfully"))
			Expect(response["data"].(map[string]any)["name"]).To(Equal("Complex Update Name")) // Should be trimmed
		})

		It("should return error for whitespace-only category name during update", func() {
			update := models.UpdateCategoryInput{Name: "   "} // Only whitespace - will become empty after trimming
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("no fields to update"))
		})

		It("should update category icon successfully", func() {
			newIcon := "new-icon"
			update := models.UpdateCategoryInput{Icon: &newIcon}
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Category updated successfully"))
			Expect(response["data"].(map[string]any)["icon"]).To(Equal(newIcon))
		})

		It("should trim whitespace from icon during update", func() {
			newIcon := "  trimmed-update-icon  " // Icon with whitespace
			update := models.UpdateCategoryInput{Icon: &newIcon}
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Category updated successfully"))
			Expect(response["data"].(map[string]any)["icon"]).To(Equal("trimmed-update-icon")) // Should be trimmed
		})

		It("should update both name and icon successfully", func() {
			newIcon := "updated-icon"
			update := models.UpdateCategoryInput{
				Name: "Complete Update",
				Icon: &newIcon,
			}
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Category updated successfully"))
			Expect(response["data"].(map[string]any)["name"]).To(Equal("Complete Update"))
			Expect(response["data"].(map[string]any)["icon"]).To(Equal(newIcon))
		})

		It("should trim whitespace from both name and icon during update", func() {
			newIcon := "  complete-trimmed-icon  " // Icon with whitespace
			update := models.UpdateCategoryInput{
				Name: "  Complete Whitespace Update  ", // Name with whitespace
				Icon: &newIcon,
			}
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			Expect(response["message"]).To(Equal("Category updated successfully"))
			Expect(response["data"].(map[string]any)["name"]).To(Equal("Complete Whitespace Update")) // Should be trimmed
			Expect(response["data"].(map[string]any)["icon"]).To(Equal("complete-trimmed-icon"))      // Should be trimmed
		})

		It("should return error for non-existent category id", func() {
			update := models.UpdateCategoryInput{Name: "Updated Name"}
			url := "/category/9999"
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(Equal("category not found"))
		})

		It("should return error when updating category of different user", func() {
			update := models.UpdateCategoryInput{Name: "Updated Name"}
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, response := testHelperUser2.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(Equal("category not found"))
		})

		It("should return error when updating to duplicate name for same user", func() {
			// Create another category
			input := models.CreateCategoryInput{
				Name: "Unique Category",
				Icon: "unique-icon",
			}
			resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))

			// Try to update first category to have same name
			update := models.UpdateCategoryInput{Name: "Unique Category"}
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusConflict))
			Expect(response["message"]).To(Equal("category with this name already exists for this user"))
		})

		It("should return error for invalid category id", func() {
			update := models.UpdateCategoryInput{Name: "Updated Name"}
			url := "/category/invalid"
			resp, response := testHelperUser1.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid category id"))
		})

		It("should return error for invalid JSON", func() {
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, _ := testHelperUser1.MakeRequest(http.MethodPatch, url, "invalid json")
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
		})

		It("should return error for invalid authorization", func() {
			update := models.UpdateCategoryInput{Name: "Updated Name"}
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodPatch, url, update)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})

	Describe("DeleteCategory", func() {
		var categoryId int64 = 2

		It("should delete category successfully", func() {
			// Create a new category first
			input := models.CreateCategoryInput{
				Name: "Category to Delete",
				Icon: "delete-icon",
			}
			resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/category", input)
			Expect(resp.StatusCode).To(Equal(http.StatusCreated))
			categoryId := int64(response["data"].(map[string]any)["id"].(float64))

			// Delete the category
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, _ = testHelperUser1.MakeRequest(http.MethodDelete, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

			// Verify category is deleted by trying to get it
			resp, _ = testHelperUser1.MakeRequest(http.MethodGet, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
		})

		It("should return error for non-existent category id", func() {
			url := "/category/9999"
			resp, response := testHelperUser1.MakeRequest(http.MethodDelete, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(Equal("category not found"))
		})

		It("should return error when deleting category of different user", func() {
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, response := testHelperUser2.MakeRequest(http.MethodDelete, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			Expect(response["message"]).To(Equal("category not found"))
		})

		It("should return error for invalid category id", func() {
			url := "/category/invalid"
			resp, response := testHelperUser1.MakeRequest(http.MethodDelete, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			Expect(response["message"]).To(Equal("invalid category id"))
		})

		It("should return error for invalid authorization", func() {
			url := "/category/" + strconv.FormatInt(categoryId, 10)
			resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodDelete, url, nil)
			Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		})
	})
})
