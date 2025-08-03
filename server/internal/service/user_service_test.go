package service

import (
	"context"
	"expenses/internal/errors"
	mock "expenses/internal/mock/repository"
	"expenses/internal/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserService", func() {
	var (
		userService UserServiceInterface
		mockRepo    *mock.MockUserRepository
		ctx         context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		mockRepo = mock.NewMockUserRepository()
		userService = NewUserService(mockRepo)
	})

	Describe("CreateUser", func() {
		var newUser models.CreateUserInput

		BeforeEach(func() {
			newUser = models.CreateUserInput{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			}
		})

		It("should create a new user successfully", func() {
			user, err := userService.CreateUser(ctx, newUser)
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Email).To(Equal(newUser.Email))
			Expect(user.Name).To(Equal(newUser.Name))
		})

		It("should return error if user already exists", func() {
			_, err := userService.CreateUser(ctx, newUser)
			Expect(err).NotTo(HaveOccurred())
			_, err = userService.CreateUser(ctx, newUser)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.AuthError{}))
			Expect(err.(*errors.AuthError).ErrorType).To(Equal("UserAlreadyExists"))
		})
	})

	Describe("GetUserByEmail", func() {
		var newUser models.CreateUserInput

		BeforeEach(func() {
			newUser = models.CreateUserInput{
				Email:    "test2@example.com",
				Name:     "Test User 2",
				Password: "password456",
			}
			_, err := userService.CreateUser(ctx, newUser)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should get user by email", func() {
			user, err := userService.GetUserByEmailWithPassword(ctx, newUser.Email)
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Email).To(Equal(newUser.Email))
			Expect(user.Name).To(Equal(newUser.Name))
		})

		It("should return error for non-existent email", func() {
			_, err := userService.GetUserByEmailWithPassword(ctx, "notfound@example.com")
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.AuthError{}))
			Expect(err.(*errors.AuthError).ErrorType).To(Equal("UserNotFound"))
		})
	})

	Describe("GetUserById", func() {
		var newUser models.CreateUserInput
		var createdUser models.UserResponse

		BeforeEach(func() {
			newUser = models.CreateUserInput{
				Email:    "test3@example.com",
				Name:     "Test User 3",
				Password: "password789",
			}
			var err error
			createdUser, err = userService.CreateUser(ctx, newUser)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should get user by id", func() {
			user, err := userService.GetUserById(ctx, createdUser.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Email).To(Equal(newUser.Email))
			Expect(user.Name).To(Equal(newUser.Name))
		})

		It("should return error for non-existent id", func() {
			_, err := userService.GetUserById(ctx, 9999)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.AuthError{}))
			Expect(err.(*errors.AuthError).ErrorType).To(Equal("UserNotFound"))
		})
	})

	Describe("DeleteUser", func() {
		var newUser models.CreateUserInput
		var createdUser models.UserResponse

		BeforeEach(func() {
			newUser = models.CreateUserInput{
				Email:    "test4@example.com",
				Name:     "Test User 4",
				Password: "password000",
			}
			var err error
			createdUser, err = userService.CreateUser(ctx, newUser)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should delete user by id", func() {
			err := userService.DeleteUser(ctx, createdUser.Id)
			Expect(err).NotTo(HaveOccurred())
			_, err = userService.GetUserById(ctx, createdUser.Id)
			Expect(err).To(HaveOccurred())
		})

		It("should return error for non-existent id", func() {
			err := userService.DeleteUser(ctx, 9999)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.AuthError{}))
			Expect(err.(*errors.AuthError).ErrorType).To(Equal("UserNotFound"))
		})
	})

	Describe("UpdateUser", func() {
		var newUser models.CreateUserInput
		var createdUser models.UserResponse

		BeforeEach(func() {
			newUser = models.CreateUserInput{
				Email:    "test5@example.com",
				Name:     "Test User 5",
				Password: "password111",
			}
			var err error
			createdUser, err = userService.CreateUser(ctx, newUser)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should update user name", func() {
			update := models.UpdateUserInput{Name: "Updated Name"}
			user, err := userService.UpdateUser(ctx, createdUser.Id, update)
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Name).To(Equal("Updated Name"))
		})

		It("should return error for non-existent id", func() {
			update := models.UpdateUserInput{Name: "Updated Name"}
			_, err := userService.UpdateUser(ctx, 9999, update)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.AuthError{}))
			Expect(err.(*errors.AuthError).ErrorType).To(Equal("UserNotFound"))
		})
	})

	Describe("GetUserByIdWithPassword", func() {
		var newUser models.CreateUserInput
		var createdUser models.UserResponse

		BeforeEach(func() {
			newUser = models.CreateUserInput{
				Email:    "test6@example.com",
				Name:     "Test User 6",
				Password: "password222",
			}
			var err error
			createdUser, err = userService.CreateUser(ctx, newUser)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should get user with password by id", func() {
			user, err := userService.GetUserByIdWithPassword(ctx, createdUser.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Email).To(Equal(newUser.Email))
			Expect(user.Name).To(Equal(newUser.Name))
			Expect(user.Password).NotTo(BeEmpty())
		})

		It("should return error for non-existent id", func() {
			_, err := userService.GetUserByIdWithPassword(ctx, 9999)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.AuthError{}))
			Expect(err.(*errors.AuthError).ErrorType).To(Equal("UserNotFound"))
		})
	})

	Describe("UpdateUserPassword", func() {
		var newUser models.CreateUserInput
		var createdUser models.UserResponse

		BeforeEach(func() {
			newUser = models.CreateUserInput{
				Email:    "test7@example.com",
				Name:     "Test User 7",
				Password: "password333",
			}
			var err error
			createdUser, err = userService.CreateUser(ctx, newUser)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should update user password", func() {
			newPassword := "newpassword123"
			user, err := userService.UpdateUserPassword(ctx, createdUser.Id, newPassword)
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Email).To(Equal(newUser.Email))
			Expect(user.Name).To(Equal(newUser.Name))

			// Verify the password was updated by trying to get user with password
			userWithPass, err := userService.GetUserByIdWithPassword(ctx, createdUser.Id)
			Expect(err).NotTo(HaveOccurred())
			Expect(userWithPass.Password).NotTo(Equal(newUser.Password))
		})

		It("should return error for non-existent id", func() {
			_, err := userService.UpdateUserPassword(ctx, 9999, "newpassword")
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.AuthError{}))
			Expect(err.(*errors.AuthError).ErrorType).To(Equal("UserNotFound"))
		})
	})
})
