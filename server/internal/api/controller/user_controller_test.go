package controller_test

import (
	"expenses/internal/models"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserController", func() {
	Describe("GetUserById", func() {
		Context("with valid token", func() {
			It("should get user details successfully", func() {
				resp, response := testHelperUser1.MakeRequest(http.MethodGet, "/user", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("User retrieved successfully"))
				Expect(response["data"]).To(HaveKey("id"))
				Expect(response["data"]).To(HaveKey("email"))
				Expect(response["data"]).To(HaveKey("name"))
				Expect(response["data"]).NotTo(HaveKey("password"))
				Expect(response["data"].(map[string]any)["email"]).To(Equal("test1@example.com"))
				Expect(response["data"].(map[string]any)["name"]).To(Equal("Test user 1"))
			})
		})

		Context("with invalid token", func() {
			It("should return unauthorized", func() {
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodGet, "/user", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return unauthorized for missing authorization header", func() {
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodGet, "/user", "")
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("UpdateUser", func() {
		Context("with valid input", func() {
			It("should update user details successfully", func() {
				updateInput := models.UpdateUserInput{
					Name: "Updated Name",
				}
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, "/user", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("User updated successfully"))
				data := response["data"].(map[string]any)
				Expect(data["name"]).To(Equal("Updated Name"))
			})

			It("should trim whitespace from user name and update successfully", func() {
				updateInput := models.UpdateUserInput{
					Name: "  Trimmed Name  ", // Name with leading and trailing whitespace
				}
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, "/user", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("User updated successfully"))
				data := response["data"].(map[string]any)
				Expect(data["name"]).To(Equal("Trimmed Name")) // Should be trimmed
			})

			It("should trim complex whitespace characters from user name", func() {
				updateInput := models.UpdateUserInput{
					Name: "\t  Complex Whitespace Name  \n", // Name with tabs and newlines
				}
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, "/user", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("User updated successfully"))
				data := response["data"].(map[string]any)
				Expect(data["name"]).To(Equal("Complex Whitespace Name")) // Should be trimmed
			})
		})

		Context("with invalid input", func() {
			It("should return bad request for invalid JSON", func() {
				resp, _ := testHelperUser1.MakeRequest(http.MethodPatch, "/user", "invalid json")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for empty body", func() {
				resp, _ := testHelperUser1.MakeRequest(http.MethodPatch, "/user", "")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for whitespace-only name", func() {
				updateInput := models.UpdateUserInput{
					Name: "   ",
				}
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, "/user", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(Equal("no fields to update"))
			})

			It("should return bad request for tabs and newlines only in name", func() {
				updateInput := models.UpdateUserInput{
					Name: "\t\n  \r  ", // Only various whitespace characters
				}
				resp, response := testHelperUser1.MakeRequest(http.MethodPatch, "/user", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("no fields to update"))
			})

			It("should return bad request for invalid input", func() {
				updateInput := map[string]any{
					"somerandomparam": 123,
				}
				resp, _ := testHelperUser1.MakeRequest(http.MethodPatch, "/user", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return not found when user doesn't exist", func() {
				// Create a new user
				userInput := models.CreateUserInput{
					Email:    "userToUpdateNotExists@example.com",
					Name:     "Test user for update not exists",
					Password: "password123",
				}
				user := NewTestHelper(baseURL)
				resp, _ := user.MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				user.Login("userToUpdateNotExists@example.com", "password123")

				// Delete the user
				resp, _ = user.MakeRequest(http.MethodDelete, "/user", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

				// Now try to update the deleted user - should get 404
				updateInput := models.UpdateUserInput{
					Name: "Updated Name",
				}
				resp, response := user.MakeRequest(http.MethodPatch, "/user", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(Equal("user not found"))
			})

			It("should be unauthorized for invalid token", func() {
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodPatch, "/user", "{}")
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should be unauthorized for missing authorization header", func() {
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodPatch, "/user", "")
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("UpdateUserPassword", func() {
		Context("with valid input", func() {
			It("should update password successfully", func() {
				// Create a new user
				userInput := models.CreateUserInput{
					Email:    "passwordUpdate@example.com",
					Name:     "Test password update",
					Password: "password123",
				}
				user := NewTestHelper(baseURL)
				resp, _ := user.MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				user.Login("passwordUpdate@example.com", "password123")

				updateInput := models.UpdateUserPasswordInput{
					OldPassword: "password123",
					NewPassword: "newpassword123",
				}
				resp, response := user.MakeRequest(http.MethodPost, "/user/password", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("User password updated successfully"))
				Expect(response).NotTo(HaveKey("password"))

				// Verify new password works by trying to login
				loginInput := models.LoginInput{
					Email:    "passwordUpdate@example.com",
					Password: "newpassword123",
				}
				resp, _ = user.MakeRequest(http.MethodPost, "/login", loginInput)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})

			It("should trim whitespace from passwords and update successfully", func() {
				// Create a new user
				userInput := models.CreateUserInput{
					Email:    "passwordWhitespaceUpdate@example.com",
					Name:     "Test password whitespace update",
					Password: "password123",
				}
				user := NewTestHelper(baseURL)
				resp, _ := user.MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				user.Login("passwordWhitespaceUpdate@example.com", "password123")

				updateInput := models.UpdateUserPasswordInput{
					OldPassword: "  password123  ",    // Old password with whitespace
					NewPassword: "  newpassword123  ", // New password with whitespace
				}
				resp, response := user.MakeRequest(http.MethodPost, "/user/password", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(response["message"]).To(Equal("User password updated successfully"))

				// Verify new password works (trimmed version) by trying to login
				loginInput := models.LoginInput{
					Email:    "passwordWhitespaceUpdate@example.com",
					Password: "newpassword123", // Trimmed password should work
				}
				resp, _ = user.MakeRequest(http.MethodPost, "/login", loginInput)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("with invalid input", func() {
			It("should return unauthorized for incorrect old password", func() {
				updateInput := models.UpdateUserPasswordInput{
					OldPassword: "wrongpassword",
					NewPassword: "newpassword123",
				}
				resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/user/password", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return bad request for invalid JSON", func() {
				resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/user/password", "invalid json")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for empty body", func() {
				resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/user/password", "")
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for whitespace-only old password", func() {
				updateInput := models.UpdateUserPasswordInput{
					OldPassword: "   ", // Only whitespace - will become empty after trimming
					NewPassword: "newpassword123",
				}
				resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/user/password", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("required"))
			})

			It("should return bad request for whitespace-only new password", func() {
				updateInput := models.UpdateUserPasswordInput{
					OldPassword: "password123",
					NewPassword: "   ", // Only whitespace - will become empty after trimming
				}
				resp, response := testHelperUser1.MakeRequest(http.MethodPost, "/user/password", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				Expect(response["message"]).To(ContainSubstring("validation"))
			})

			It("should be unauthorized for invalid token", func() {
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodPost, "/user/password", "{}")
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should be unauthorized for missing authorization header", func() {
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodPost, "/user/password", "")
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return bad request for missing old password", func() {
				updateInput := models.UpdateUserPasswordInput{
					NewPassword: "newpassword123",
				}
				resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/user/password", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for missing new password", func() {
				updateInput := models.UpdateUserPasswordInput{
					OldPassword: "password",
				}
				resp, _ := testHelperUser1.MakeRequest(http.MethodPost, "/user/password", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return not found when user doesn't exist", func() {
				// Create a new user
				user := NewTestHelper(baseURL)
				userInput := models.CreateUserInput{
					Email:    "userToPasswordUpdateNotExists@example.com",
					Name:     "Test user for password update not exists",
					Password: "password123",
				}
				resp, _ := user.MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				user.Login("userToPasswordUpdateNotExists@example.com", "password123")

				// Delete the user
				resp, _ = user.MakeRequest(http.MethodDelete, "/user", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

				// Now try to update password for the deleted user - should get 404
				updateInput := models.UpdateUserPasswordInput{
					OldPassword: "password123",
					NewPassword: "newpassword123",
				}
				resp, response := user.MakeRequest(http.MethodPost, "/user/password", updateInput)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				Expect(response["message"]).To(Equal("user not found"))
			})
		})
	})

	Describe("DeleteUser", func() {
		Context("with valid token", func() {
			It("should delete user successfully", func() {
				user := NewTestHelper(baseURL)
				// Create a new user
				userInput := models.CreateUserInput{
					Email:    "deleteUser@example.com",
					Name:     "Test user to delete",
					Password: "password123",
				}
				resp, _ := user.MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				user.Login("deleteUser@example.com", "password123")

				// Delete the user
				resp, _ = user.MakeRequest(http.MethodDelete, "/user", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

				// Verify user is deleted by trying to get user details
				resp, _ = user.MakeRequest(http.MethodGet, "/user", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		Context("with invalid token", func() {
			It("should return unauthorized", func() {
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodDelete, "/user", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return unauthorized for missing authorization header", func() {
				resp, _ := testHelperUnauthenticated.MakeRequest(http.MethodDelete, "/user", "")
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when user doesn't exist", func() {
			It("should return error when user doesn't exist", func() {
				user := NewTestHelper(baseURL)
				// Create a new user
				userInput := models.CreateUserInput{
					Email:    "userToDeleteNotExists@example.com",
					Name:     "Test user for delete not exists",
					Password: "password123",
				}
				resp, _ := user.MakeRequest(http.MethodPost, "/signup", userInput)
				Expect(resp.StatusCode).To(Equal(http.StatusCreated))

				user.Login("userToDeleteNotExists@example.com", "password123")

				// Delete the user first time
				resp, _ = user.MakeRequest(http.MethodDelete, "/user", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

				// Now try to delete the already deleted user - should get 404
				resp, _ = user.MakeRequest(http.MethodDelete, "/user", nil)
				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			})
		})
	})
})
