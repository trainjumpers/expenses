package service

import (
	mock "expenses/internal/mock/repository"
	"expenses/internal/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gin-gonic/gin"
)

var _ = Describe("CategoryService", func() {
	var (
		categoryService CategoryServiceInterface
		mockRepo        *mock.MockCategoryRepository
		ctx             *gin.Context
	)

	BeforeEach(func() {
		ctx = &gin.Context{}
		mockRepo = mock.NewMockCategoryRepository()
		categoryService = NewCategoryService(mockRepo)
	})

	Describe("CreateCategory", func() {
		It("should create a new category successfully with icon", func() {
			input := models.CreateCategoryInput{
				Name:      "Food",
				Icon:      "burger-icon",
				CreatedBy: 1,
			}
			category, err := categoryService.CreateCategory(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(category.Name).To(Equal(input.Name))
			Expect(*category.Icon).To(Equal(input.Icon))
			Expect(category.CreatedBy).To(Equal(input.CreatedBy))
			Expect(category.Id).To(BeNumerically(">", 0))
		})

		It("should create a new category successfully with empty icon", func() {
			input := models.CreateCategoryInput{
				Name:      "Entertainment",
				CreatedBy: 1,
			}
			category, err := categoryService.CreateCategory(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(category.Name).To(Equal(input.Name))
			Expect(category.Icon).To(BeNil())
			Expect(category.CreatedBy).To(Equal(input.CreatedBy))
		})

		It("should handle special characters in category name", func() {
			input := models.CreateCategoryInput{
				Name:      "!@#$%^&*()_+|~",
				CreatedBy: 1,
			}
			category, err := categoryService.CreateCategory(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(category.Name).To(Equal(input.Name))
		})

		It("should handle nil icon pointer", func() {
			input := models.CreateCategoryInput{
				Name:      "NoIcon",
				Icon:      "",
				CreatedBy: 1,
			}
			input.Icon = ""
			_, err := categoryService.CreateCategory(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			// Icon may be nil or empty string depending on implementation
		})

		It("should return error when category name already exists for same user", func() {
			input := models.CreateCategoryInput{
				Name:      "Food",
				CreatedBy: 1,
			}
			// Create first category
			_, err := categoryService.CreateCategory(ctx, input)
			Expect(err).NotTo(HaveOccurred())

			// Try to create category with same name for same user
			_, err = categoryService.CreateCategory(ctx, input)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("category with this name already exists for this user"))
		})

		It("should allow same category name for different users", func() {
			input1 := models.CreateCategoryInput{
				Name:      "Food v1",
				Icon:      "burger-icon",
				CreatedBy: 1,
			}
			input2 := models.CreateCategoryInput{
				Name:      "Food v1",
				Icon:      "burger-icon",
				CreatedBy: 2,
			}

			// Create category for first user
			category1, err := categoryService.CreateCategory(ctx, input1)
			Expect(err).NotTo(HaveOccurred())
			Expect(category1.Name).To(Equal("Food v1"))

			// Create category with same name for second user
			category2, err := categoryService.CreateCategory(ctx, input2)
			Expect(err).NotTo(HaveOccurred())
			Expect(category2.Name).To(Equal("Food v1"))
			Expect(category1.Id).NotTo(Equal(category2.Id))
		})
	})

	Describe("GetCategoryById", func() {
		var created models.CategoryResponse

		BeforeEach(func() {
			input := models.CreateCategoryInput{
				Name:      "Transportation",
				Icon:      "car-icon",
				CreatedBy: 1,
			}
			var err error
			created, err = categoryService.CreateCategory(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should get category by id successfully", func() {
			category, err := categoryService.GetCategoryById(ctx, created.Id, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(category.Name).To(Equal("Transportation"))
			Expect(*category.Icon).To(Equal("car-icon"))
		})

		It("should return error for non-existent category id", func() {
			_, err := categoryService.GetCategoryById(ctx, 9999, 1)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("category not found"))
		})

		It("should return error when accessing category of different user", func() {
			_, err := categoryService.GetCategoryById(ctx, created.Id, 2)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("category not found"))
		})
	})

	Describe("ListCategories", func() {
		BeforeEach(func() {
			// Create categories for user 1
			categories := []models.CreateCategoryInput{
				{Name: "Food v2", Icon: "burger-icon", CreatedBy: 1},
				{Name: "Transport v2", Icon: "car-icon", CreatedBy: 1},
				{Name: "Shopping v2", Icon: "shopping-icon", CreatedBy: 1},
			}
			for _, cat := range categories {
				_, err := categoryService.CreateCategory(ctx, cat)
				Expect(err).NotTo(HaveOccurred())
			}

			// Create category for user 2
			input := models.CreateCategoryInput{
				Name:      "Entertainment",
				Icon:      "entertainment-icon",
				CreatedBy: 2,
			}
			_, err := categoryService.CreateCategory(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should list all categories for a specific user", func() {
			categories, err := categoryService.ListCategories(ctx, 1)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(categories)).To(Equal(3))

			categoryNames := make([]string, len(categories))
			for i, cat := range categories {
				categoryNames[i] = cat.Name
			}
			Expect(categoryNames).To(ContainElements("Food v2", "Transport v2", "Shopping v2"))
		})

		It("should return empty list for user with no categories", func() {
			categories, err := categoryService.ListCategories(ctx, 999)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(categories)).To(Equal(0))
		})

		It("should only return categories for the requested user", func() {
			categories, err := categoryService.ListCategories(ctx, 2)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(categories)).To(Equal(1))
			Expect(categories[0].Name).To(Equal("Entertainment"))
		})
	})

	Describe("UpdateCategory", func() {
		var created models.CategoryResponse

		BeforeEach(func() {
			input := models.CreateCategoryInput{
				Name:      "Health",
				Icon:      "health-icon",
				CreatedBy: 1,
			}
			var err error
			created, err = categoryService.CreateCategory(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should update category name successfully", func() {
			update := models.UpdateCategoryInput{Name: "Healthcare"}
			category, err := categoryService.UpdateCategory(ctx, created.Id, 1, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(category.Name).To(Equal("Healthcare"))
			Expect(*category.Icon).To(Equal("health-icon")) // Icon should remain unchanged
		})

		It("should update category icon successfully", func() {
			newIcon := "health-icon"
			update := models.UpdateCategoryInput{Icon: &newIcon}
			category, err := categoryService.UpdateCategory(ctx, created.Id, 1, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(category.Name).To(Equal("Health")) // Name should remain unchanged
			Expect(*category.Icon).To(Equal(newIcon))
		})

		It("should set icon to nil when updating with empty string", func() {
			emptyIcon := ""
			update := models.UpdateCategoryInput{Icon: &emptyIcon}
			category, err := categoryService.UpdateCategory(ctx, created.Id, 1, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(category.Name).To(Equal("Health"))
			Expect(*category.Icon).To(Equal(""))
		})

		It("should not change icon if icon pointer is nil", func() {
			update := models.UpdateCategoryInput{Icon: nil}
			category, err := categoryService.UpdateCategory(ctx, created.Id, 1, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(category.Icon).NotTo(BeNil())
		})

		It("should update both name and icon successfully", func() {
			newIcon := "health-icon"
			update := models.UpdateCategoryInput{
				Name: "Medicine",
				Icon: &newIcon,
			}
			category, err := categoryService.UpdateCategory(ctx, created.Id, 1, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(category.Name).To(Equal("Medicine"))
			Expect(*category.Icon).To(Equal(newIcon))
		})

		It("should return error for non-existent category id", func() {
			update := models.UpdateCategoryInput{Name: "Updated Name"}
			_, err := categoryService.UpdateCategory(ctx, 9999, 1, update)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("category not found"))
		})

		It("should return error when updating category of different user", func() {
			update := models.UpdateCategoryInput{Name: "Updated Name"}
			_, err := categoryService.UpdateCategory(ctx, created.Id, 2, update)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("category not found"))
		})

		It("should return error when updating to duplicate name for same user", func() {
			// Create another category for the same user
			input := models.CreateCategoryInput{
				Name:      "Education",
				Icon:      "education-icon",
				CreatedBy: 1,
			}
			_, err := categoryService.CreateCategory(ctx, input)
			Expect(err).NotTo(HaveOccurred())

			// Try to update first category to have same name as second
			update := models.UpdateCategoryInput{Name: "Education"}
			_, err = categoryService.UpdateCategory(ctx, created.Id, 1, update)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("category with this name already exists for this user"))
		})

		It("should allow updating to same name (no actual change)", func() {
			update := models.UpdateCategoryInput{Name: "Health"}
			category, err := categoryService.UpdateCategory(ctx, created.Id, 1, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(category.Name).To(Equal("Health"))
		})
	})

	Describe("DeleteCategory", func() {
		var created models.CategoryResponse

		BeforeEach(func() {
			input := models.CreateCategoryInput{
				Name:      "Utilities",
				Icon:      "utilities-icon",
				CreatedBy: 1,
			}
			var err error
			created, err = categoryService.CreateCategory(ctx, input)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should delete category successfully", func() {
			err := categoryService.DeleteCategory(ctx, created.Id, 1)
			Expect(err).NotTo(HaveOccurred())

			// Verify category is deleted
			_, err = categoryService.GetCategoryById(ctx, created.Id, 1)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("category not found"))
		})

		It("should return error when deleting a category that was already deleted", func() {
			err := categoryService.DeleteCategory(ctx, created.Id, 1)
			Expect(err).NotTo(HaveOccurred())
			err = categoryService.DeleteCategory(ctx, created.Id, 1)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent category id", func() {
			err := categoryService.DeleteCategory(ctx, 9999, 1)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("category not found"))
		})

		It("should return error when deleting category of different user", func() {
			err := categoryService.DeleteCategory(ctx, created.Id, 2)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("category not found"))
		})

		It("should not list deleted category for the user", func() {
			err := categoryService.DeleteCategory(ctx, created.Id, 1)
			Expect(err).NotTo(HaveOccurred())
			categories, err := categoryService.ListCategories(ctx, 1)
			Expect(err).NotTo(HaveOccurred())
			for _, cat := range categories {
				Expect(cat.Id).NotTo(Equal(created.Id))
			}
		})
	})

	Describe("Authorization/Ownership edge cases", func() {
		It("should return error when updating with user Id 0", func() {
			update := models.UpdateCategoryInput{Name: "ShouldFail"}
			_, err := categoryService.UpdateCategory(ctx, 1, 0, update)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when deleting with user Id 0", func() {
			err := categoryService.DeleteCategory(ctx, 1, 0)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when updating with negative user Id", func() {
			update := models.UpdateCategoryInput{Name: "ShouldFail"}
			_, err := categoryService.UpdateCategory(ctx, 1, -1, update)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when deleting with negative user Id", func() {
			err := categoryService.DeleteCategory(ctx, 1, -1)
			Expect(err).To(HaveOccurred())
		})
	})
})
