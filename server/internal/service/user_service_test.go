package service

import (
	"expenses/internal/errors"
	mock "expenses/internal/mock/repository"
	"expenses/internal/models"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gin-gonic/gin"
)

var _ = Describe("UserService", func() {
	var (
		userService UserServiceInterface
		mockRepo    *mock.MockUserRepository
		ctx         *gin.Context
	)

	BeforeEach(func() {
		ctx = &gin.Context{}
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
			user, err := userService.GetUserByEmail(ctx, newUser.Email)
			Expect(err).NotTo(HaveOccurred())
			Expect(user.Email).To(Equal(newUser.Email))
			Expect(user.Name).To(Equal(newUser.Name))
		})

		It("should return error for non-existent email", func() {
			_, err := userService.GetUserByEmail(ctx, "notfound@example.com")
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
})
