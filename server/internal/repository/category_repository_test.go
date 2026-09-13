package repository

import (
	"context"
	"fmt"
	"time"

	"expenses/internal/config"
	"expenses/internal/models"
	databasemanager "expenses/pkg/database/manager"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CategoryRepository", func() {
	var (
		ctx    context.Context
		cfg    *config.Config
		db     databasemanager.DatabaseManager
		repo   CategoryRepositoryInterface
		userID int64
	)

	BeforeEach(func() {
		var err error
		cfg, err = config.NewConfig()
		Expect(err).NotTo(HaveOccurred())

		db, err = databasemanager.NewDatabaseManager(cfg)
		Expect(err).NotTo(HaveOccurred())

		repo = NewCategoryRepository(db, cfg)
		ctx = context.Background()
		userID = createRepoTestUser(ctx, db, cfg.DBSchema)
	})

	AfterEach(func() {
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf(`DELETE FROM %s.transaction_category_mapping WHERE transaction_id IN (SELECT id FROM %s.transaction WHERE created_by = $1)`, cfg.DBSchema, cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.transaction WHERE created_by = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.categories WHERE created_by = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.account WHERE created_by = $1", cfg.DBSchema), userID)
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.user WHERE id = $1", cfg.DBSchema), userID)
		Expect(db.Close()).To(Succeed())
	})

	createCategory := func(name string) models.CategoryResponse {
		category, err := repo.CreateCategory(ctx, models.CreateCategoryInput{Name: name, CreatedBy: userID})
		Expect(err).NotTo(HaveOccurred())
		return category
	}

	Describe("CreateCategory", func() {
		It("creates a category with an icon", func() {
			icon := "wallet"
			category, err := repo.CreateCategory(ctx, models.CreateCategoryInput{Name: "Shopping", Icon: icon, CreatedBy: userID})
			Expect(err).NotTo(HaveOccurred())
			Expect(category.Id).NotTo(BeZero())
			Expect(category.Name).To(Equal("Shopping"))
			Expect(category.Icon).NotTo(BeNil())
			Expect(*category.Icon).To(Equal(icon))
		})

		It("rejects a duplicate name for the same user", func() {
			createCategory("Food")

			_, err := repo.CreateCategory(ctx, models.CreateCategoryInput{Name: "Food", CreatedBy: userID})
			expectAuthErrorType(err, "CategoryAlreadyExists")
		})

		It("returns an error for an unknown user", func() {
			_, err := repo.CreateCategory(ctx, models.CreateCategoryInput{Name: "Orphan", CreatedBy: 999999999})
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetCategoryById", func() {
		It("find a category for its owner", func() {
			created := createCategory("Groceries")

			fetched, err := repo.GetCategoryById(ctx, created.Id, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetched.Id).To(Equal(created.Id))
			Expect(fetched.Name).To(Equal("Groceries"))
		})

		It("returns not found for an unknown category", func() {
			_, err := repo.GetCategoryById(ctx, 999999, userID)
			expectAuthErrorType(err, "CategoryNotFound")
		})

		It("hides categories from another user", func() {
			created := createCategory("Private")
			otherUserID := createRepoTestUser(ctx, db, cfg.DBSchema)
			defer func() {
				_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.user WHERE id = $1", cfg.DBSchema), otherUserID)
			}()

			_, err := repo.GetCategoryById(ctx, created.Id, otherUserID)
			expectAuthErrorType(err, "CategoryNotFound")
		})
	})

	Describe("ListCategories", func() {
		It("returns an empty list for a new user", func() {
			categories, err := repo.ListCategories(ctx, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(categories).To(BeEmpty())
		})

		It("lists categories newest first", func() {
			first := createCategory("First")
			second := createCategory("Second")

			categories, err := repo.ListCategories(ctx, userID)
			Expect(err).NotTo(HaveOccurred())
			Expect(categories).To(HaveLen(2))
			Expect(categories[0].Id).To(Equal(second.Id))
			Expect(categories[1].Id).To(Equal(first.Id))
		})
	})

	Describe("UpdateCategory", func() {
		It("updates the name and icon", func() {
			created := createCategory("Old name")
			icon := "tag"

			updated, err := repo.UpdateCategory(ctx, created.Id, userID, models.UpdateCategoryInput{Name: "New name", Icon: &icon})
			Expect(err).NotTo(HaveOccurred())
			Expect(updated.Name).To(Equal("New name"))
			Expect(updated.Icon).NotTo(BeNil())
			Expect(*updated.Icon).To(Equal(icon))
		})

		It("rejects a duplicate name", func() {
			createCategory("Taken")
			created := createCategory("Mine")

			_, err := repo.UpdateCategory(ctx, created.Id, userID, models.UpdateCategoryInput{Name: "Taken"})
			expectAuthErrorType(err, "CategoryAlreadyExists")
		})

		It("returns not found for an unknown category", func() {
			_, err := repo.UpdateCategory(ctx, 999999, userID, models.UpdateCategoryInput{Name: "Ghost"})
			expectAuthErrorType(err, "CategoryNotFound")
		})

		It("rejects an empty update", func() {
			created := createCategory("Empty update")

			_, err := repo.UpdateCategory(ctx, created.Id, userID, models.UpdateCategoryInput{})
			expectAuthErrorType(err, "NoFieldsToUpdate")
		})
	})

	Describe("DeleteCategory", func() {
		It("deletes the category and its transaction mappings", func() {
			created := createCategory("To delete")
			accountID := createRepoTestAccount(ctx, db, cfg.DBSchema, userID)

			var transactionID int64
			Expect(db.FetchOne(ctx,
				fmt.Sprintf("INSERT INTO %s.transaction (name, amount, date, account_id, created_by) VALUES ($1, $2, $3, $4, $5) RETURNING id", cfg.DBSchema),
				"Mapped transaction", 20.0, time.Now().UTC(), accountID, userID,
			).Scan(&transactionID)).To(Succeed())
			Expect(db.FetchOne(ctx,
				fmt.Sprintf("INSERT INTO %s.transaction_category_mapping (category_id, transaction_id) VALUES ($1, $2) RETURNING id", cfg.DBSchema),
				created.Id, transactionID,
			).Scan(new(int64))).To(Succeed())

			Expect(repo.DeleteCategory(ctx, created.Id, userID)).To(Succeed())

			_, err := repo.GetCategoryById(ctx, created.Id, userID)
			expectAuthErrorType(err, "CategoryNotFound")

			var mappingCount int
			Expect(db.FetchOne(ctx,
				fmt.Sprintf("SELECT COUNT(*) FROM %s.transaction_category_mapping WHERE category_id = $1", cfg.DBSchema),
				created.Id,
			).Scan(&mappingCount)).To(Succeed())
			Expect(mappingCount).To(Equal(0))

			err = repo.DeleteCategory(ctx, created.Id, userID)
			expectAuthErrorType(err, "CategoryNotFound")
		})
	})
})
