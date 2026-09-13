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

var _ = Describe("UserRepository", func() {
	var (
		ctx  context.Context
		cfg  *config.Config
		db   databasemanager.DatabaseManager
		repo UserRepositoryInterface
	)

	BeforeEach(func() {
		var err error
		cfg, err = config.NewConfig()
		Expect(err).NotTo(HaveOccurred())

		db, err = databasemanager.NewDatabaseManager(cfg)
		Expect(err).NotTo(HaveOccurred())

		repo = NewUserRepository(db, cfg)
		ctx = context.Background()
	})

	AfterEach(func() {
		_, _ = db.ExecuteQuery(ctx, fmt.Sprintf("DELETE FROM %s.user WHERE email LIKE $1", cfg.DBSchema), "user-repository-%@example.com")
		Expect(db.Close()).To(Succeed())
	})

	newUser := func(name string) models.CreateUserInput {
		return models.CreateUserInput{
			Email:    fmt.Sprintf("user-repository-%d@example.com", time.Now().UnixNano()),
			Name:     name,
			Password: "hashed-password",
		}
	}

	Describe("CreateUser", func() {
		It("creates and returns a user", func() {
			input := newUser("Created User")

			created, err := repo.CreateUser(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(created.Id).NotTo(BeZero())
			Expect(created.Email).To(Equal(input.Email))
			Expect(created.Name).To(Equal(input.Name))
		})

		It("rejects a duplicate email", func() {
			input := newUser("Duplicate User")
			_, err := repo.CreateUser(ctx, input)
			Expect(err).NotTo(HaveOccurred())

			_, err = repo.CreateUser(ctx, input)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("GetUserByEmailWithPassword", func() {
		It("finds an active user with the password", func() {
			input := newUser("Email Lookup")
			created, err := repo.CreateUser(ctx, input)
			Expect(err).NotTo(HaveOccurred())

			found, err := repo.GetUserByEmailWithPassword(ctx, input.Email)
			Expect(err).NotTo(HaveOccurred())
			Expect(found.Id).To(Equal(created.Id))
			Expect(found.Password).To(Equal(input.Password))
		})

		It("returns not found for an unknown email", func() {
			_, err := repo.GetUserByEmailWithPassword(ctx, "missing@example.com")
			expectAuthErrorType(err, "UserNotFound")
		})

		It("excludes deleted users", func() {
			input := newUser("Deleted Lookup")
			created, err := repo.CreateUser(ctx, input)
			Expect(err).NotTo(HaveOccurred())
			Expect(repo.DeleteUser(ctx, created.Id)).To(Succeed())

			_, err = repo.GetUserByEmailWithPassword(ctx, input.Email)
			expectAuthErrorType(err, "UserNotFound")
		})
	})

	Describe("GetUserByIdWithPassword and GetUserById", func() {
		It("finds an active user", func() {
			input := newUser("Id Lookup")
			created, err := repo.CreateUser(ctx, input)
			Expect(err).NotTo(HaveOccurred())

			withPassword, err := repo.GetUserByIdWithPassword(ctx, created.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(withPassword.Password).To(Equal(input.Password))

			withoutPassword, err := repo.GetUserById(ctx, created.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(withoutPassword.Email).To(Equal(input.Email))
		})

		It("returns not found for an unknown id", func() {
			_, err := repo.GetUserByIdWithPassword(ctx, 999999)
			expectAuthErrorType(err, "UserNotFound")

			_, err = repo.GetUserById(ctx, 999999)
			expectAuthErrorType(err, "UserNotFound")
		})
	})

	Describe("DeleteUser", func() {
		It("soft deletes a user once", func() {
			created, err := repo.CreateUser(ctx, newUser("Delete Me"))
			Expect(err).NotTo(HaveOccurred())

			Expect(repo.DeleteUser(ctx, created.Id)).To(Succeed())

			err = repo.DeleteUser(ctx, created.Id)
			expectAuthErrorType(err, "UserNotFound")
		})
	})

	Describe("UpdateUser", func() {
		It("updates the user name", func() {
			created, err := repo.CreateUser(ctx, newUser("Before Update"))
			Expect(err).NotTo(HaveOccurred())

			updated, err := repo.UpdateUser(ctx, created.Id, models.UpdateUserInput{Name: "After Update"})
			Expect(err).NotTo(HaveOccurred())
			Expect(updated.Name).To(Equal("After Update"))
		})

		It("returns not found for an unknown user", func() {
			_, err := repo.UpdateUser(ctx, 999999, models.UpdateUserInput{Name: "Ghost"})
			expectAuthErrorType(err, "UserNotFound")
		})

		It("rejects an empty update", func() {
			created, err := repo.CreateUser(ctx, newUser("Empty Update"))
			Expect(err).NotTo(HaveOccurred())

			_, err = repo.UpdateUser(ctx, created.Id, models.UpdateUserInput{})
			expectAuthErrorType(err, "NoFieldsToUpdate")
		})
	})

	Describe("UpdateUserPassword", func() {
		It("updates the password", func() {
			created, err := repo.CreateUser(ctx, newUser("Password Update"))
			Expect(err).NotTo(HaveOccurred())

			updated, err := repo.UpdateUserPassword(ctx, created.Id, "new-hashed-password")
			Expect(err).NotTo(HaveOccurred())
			Expect(updated.Id).To(Equal(created.Id))

			found, err := repo.GetUserByIdWithPassword(ctx, created.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(found.Password).To(Equal("new-hashed-password"))
		})

		It("returns not found for an unknown user", func() {
			_, err := repo.UpdateUserPassword(ctx, 999999, "new-password")
			expectAuthErrorType(err, "UserNotFound")
		})
	})
})
