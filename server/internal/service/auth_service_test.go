package service

import (
	"expenses/internal/config"
	"expenses/internal/errors"
	mock "expenses/internal/mock/repository"
	"expenses/internal/models"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gin-gonic/gin"
)

var _ = Describe("AuthService", func() {
	var (
		authService AuthServiceInterface
		userService UserServiceInterface
		mockRepo    *mock.MockUserRepository
		cfg         *config.Config
		ctx         *gin.Context
	)

	BeforeEach(func() {
		// Set environment variables before creating config
		os.Setenv("ENV", "test")
		os.Setenv("JWT_SECRET", "test-secret")
		os.Setenv("DB_SCHEMA", "test_schema")

		ctx = &gin.Context{}
		mockRepo = mock.NewMockUserRepository()
		userService = NewUserService(mockRepo)
		var err error
		cfg, err = config.NewConfig()
		Expect(err).NotTo(HaveOccurred())
		authService = NewAuthService(userService, cfg)
	})

	Describe("Signup", func() {
		var newUser models.CreateUserInput

		BeforeEach(func() {
			newUser = models.CreateUserInput{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			}
		})

		It("should successfully create a new user and return auth tokens", func() {
			response, err := authService.Signup(ctx, newUser)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.User.Email).To(Equal(newUser.Email))
			Expect(response.User.Name).To(Equal(newUser.Name))
			Expect(response.AccessToken).NotTo(BeEmpty())
			Expect(response.RefreshToken).NotTo(BeEmpty())
		})

		It("should return error when user already exists", func() {
			// Create user first time
			_, err := authService.Signup(ctx, newUser)
			Expect(err).NotTo(HaveOccurred())

			// Try to create same user again
			_, err = authService.Signup(ctx, newUser)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.AuthError{}))
			Expect(err.(*errors.AuthError).ErrorType).To(Equal("UserAlreadyExists"))
		})
	})

	Describe("Login", func() {
		var (
			loginInput models.LoginInput
			user       models.CreateUserInput
		)

		BeforeEach(func() {
			user = models.CreateUserInput{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			}
			loginInput = models.LoginInput{
				Email:    user.Email,
				Password: user.Password,
			}
		})

		It("should successfully login with correct credentials", func() {
			// Create user first
			_, err := authService.Signup(ctx, user)
			Expect(err).NotTo(HaveOccurred())

			// Try to login
			response, err := authService.Login(ctx, loginInput)

			Expect(err).NotTo(HaveOccurred())
			Expect(response.User.Email).To(Equal(user.Email))
			Expect(response.User.Name).To(Equal(user.Name))
			Expect(response.AccessToken).NotTo(BeEmpty())
			Expect(response.RefreshToken).NotTo(BeEmpty())
		})

		It("should return error with incorrect password", func() {
			// Create user first
			_, err := authService.Signup(ctx, user)
			Expect(err).NotTo(HaveOccurred())

			// Try to login with wrong password
			loginInput.Password = "wrongpassword"
			_, err = authService.Login(ctx, loginInput)

			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.AuthError{}))
			Expect(err.(*errors.AuthError).ErrorType).To(Equal("InvalidCredentials"))
		})

		It("should return error for non-existent user", func() {
			_, err := authService.Login(ctx, loginInput)

			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.AuthError{}))
			Expect(err.(*errors.AuthError).ErrorType).To(Equal("InvalidCredentials"))
		})
	})

	Describe("RefreshToken", func() {
		var (
			user         models.CreateUserInput
			authResponse models.AuthResponse
		)

		BeforeEach(func() {
			user = models.CreateUserInput{
				Email:    "test@example.com",
				Name:     "Test User",
				Password: "password123",
			}
		})

		It("should successfully refresh tokens with valid refresh token", func() {
			// Create user and get initial tokens
			var err error
			authResponse, err = authService.Signup(ctx, user)
			Expect(err).NotTo(HaveOccurred())

			// Try to refresh tokens
			newResponse, err := authService.RefreshToken(ctx, authResponse.RefreshToken)

			Expect(err).NotTo(HaveOccurred())
			Expect(newResponse.User.Email).To(Equal(user.Email))
			Expect(newResponse.User.Name).To(Equal(user.Name))
			Expect(newResponse.AccessToken).NotTo(BeEmpty())
			Expect(newResponse.RefreshToken).NotTo(BeEmpty())
			Expect(newResponse.RefreshToken).NotTo(Equal(authResponse.RefreshToken))
		})

		It("should return error with invalid refresh token", func() {
			_, err := authService.RefreshToken(ctx, "invalid-token")

			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.AuthError{}))
			Expect(err.(*errors.AuthError).ErrorType).To(Equal("InvalidToken"))
		})

		It("should return error with expired refresh token", func() {
			// Create user and get initial tokens
			var err error
			authResponse, err = authService.Signup(ctx, user)
			Expect(err).NotTo(HaveOccurred())

			// Manually expire the token using the test helper method
			err = authService.ExpireRefreshToken(authResponse.RefreshToken)
			Expect(err).NotTo(HaveOccurred())

			// Try to refresh tokens
			_, err = authService.RefreshToken(ctx, authResponse.RefreshToken)

			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.AuthError{}))
			Expect(err.(*errors.AuthError).ErrorType).To(Equal("InvalidToken"))
		})
	})

	Describe("UpdateUserPassword", func() {
		var (
			user         models.CreateUserInput
			authResponse models.AuthResponse
		)

		BeforeEach(func() {
			user = models.CreateUserInput{
				Email:    "test8@example.com",
				Name:     "Test User 8",
				Password: "password444",
			}
			var err error
			authResponse, err = authService.Signup(ctx, user)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should successfully update user password with correct old password", func() {
			updateInput := models.UpdateUserPasswordInput{
				OldPassword: user.Password,
				NewPassword: "newpassword123",
			}
			updatedUser, err := authService.UpdateUserPassword(ctx, authResponse.User.Id, updateInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedUser.Email).To(Equal(user.Email))
			Expect(updatedUser.Name).To(Equal(user.Name))

			// Verify new password works by trying to login
			loginInput := models.LoginInput{
				Email:    user.Email,
				Password: updateInput.NewPassword,
			}
			loginResponse, err := authService.Login(ctx, loginInput)
			Expect(err).NotTo(HaveOccurred())
			Expect(loginResponse.User.Email).To(Equal(user.Email))
		})

		It("should return error with incorrect old password", func() {
			updateInput := models.UpdateUserPasswordInput{
				OldPassword: "wrongpassword",
				NewPassword: "newpassword123",
			}
			_, err := authService.UpdateUserPassword(ctx, authResponse.User.Id, updateInput)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.AuthError{}))
			Expect(err.(*errors.AuthError).ErrorType).To(Equal("InvalidCredentials"))
		})

		It("should return error for non-existent user", func() {
			updateInput := models.UpdateUserPasswordInput{
				OldPassword: user.Password,
				NewPassword: "newpassword123",
			}
			_, err := authService.UpdateUserPassword(ctx, 9999, updateInput)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(&errors.AuthError{}))
			Expect(err.(*errors.AuthError).ErrorType).To(Equal("UserNotFound"))
		})
	})
})
