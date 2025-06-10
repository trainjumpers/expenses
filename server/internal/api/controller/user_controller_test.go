package controller_test

import (
	"bytes"
	"encoding/json"
	"expenses/internal/models"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserController", func() {
	Describe("GetUserById", func() {
		Context("with valid token", func() {
			It("should get user details successfully", func() {
				req, err := http.NewRequest(http.MethodGet, baseURL+"/user", nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("User retrieved successfully"))
				Expect(response["data"]).To(HaveKey("id"))
				Expect(response["data"]).To(HaveKey("email"))
				Expect(response["data"]).To(HaveKey("name"))
				Expect(response["data"]).NotTo(HaveKey("password"))
				Expect(response["data"].(map[string]interface{})["email"]).To(Equal("test1@example.com"))
				Expect(response["data"].(map[string]interface{})["name"]).To(Equal("Test user 1"))
			})
		})

		Context("with invalid token", func() {
			It("should return unauthorized", func() {
				req, err := http.NewRequest(http.MethodGet, baseURL+"/user", nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", "Bearer invalid-token")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return unauthorized for missing authorization header", func() {
				req, err := http.NewRequest(http.MethodGet, baseURL+"/user", nil)
				Expect(err).NotTo(HaveOccurred())

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

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

				body, _ := json.Marshal(updateInput)
				req, err := http.NewRequest(http.MethodPatch, baseURL+"/user", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("User updated successfully"))
				data := response["data"].(map[string]interface{})
				Expect(data["name"]).To(Equal("Updated Name"))
			})

			It("should trim whitespace from user name and update successfully", func() {
				updateInput := models.UpdateUserInput{
					Name: "  Trimmed Name  ", // Name with leading and trailing whitespace
				}

				body, _ := json.Marshal(updateInput)
				req, err := http.NewRequest(http.MethodPatch, baseURL+"/user", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("User updated successfully"))
				data := response["data"].(map[string]interface{})
				Expect(data["name"]).To(Equal("Trimmed Name")) // Should be trimmed
			})

			It("should trim complex whitespace characters from user name", func() {
				updateInput := models.UpdateUserInput{
					Name: "\t  Complex Whitespace Name  \n", // Name with tabs and newlines
				}

				body, _ := json.Marshal(updateInput)
				req, err := http.NewRequest(http.MethodPatch, baseURL+"/user", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("User updated successfully"))
				data := response["data"].(map[string]interface{})
				Expect(data["name"]).To(Equal("Complex Whitespace Name")) // Should be trimmed
			})
		})

		Context("with invalid input", func() {
			It("should return bad request for invalid JSON", func() {
				req, err := http.NewRequest(http.MethodPatch, baseURL+"/user", bytes.NewBuffer([]byte("invalid json")))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for empty body", func() {
				req, err := http.NewRequest(http.MethodPatch, baseURL+"/user", bytes.NewBuffer([]byte("")))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for whitespace-only name", func() {
				updateInput := models.UpdateUserInput{
					Name: "   ",
				}

				body, _ := json.Marshal(updateInput)
				req, err := http.NewRequest(http.MethodPatch, baseURL+"/user", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("no fields to update"))
			})

			It("should return bad request for tabs and newlines only in name", func() {
				updateInput := models.UpdateUserInput{
					Name: "\t\n  \r  ", // Only various whitespace characters
				}

				body, _ := json.Marshal(updateInput)
				req, err := http.NewRequest(http.MethodPatch, baseURL+"/user", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(ContainSubstring("no fields to update"))
			})

			It("should return bad request for invalid input", func() {
				updateInput := map[string]interface{}{
					"somerandomparam": 123,
				}

				body, _ := json.Marshal(updateInput)
				req, err := http.NewRequest(http.MethodPatch, baseURL+"/user", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return not found when user doesn't exist", func() {
				// Create a new user
				userInput := models.CreateUserInput{
					Email:    "userToUpdateNotExists@example.com",
					Name:     "Test user for update not exists",
					Password: "password123",
				}

				body, _ := json.Marshal(userInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/signup", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["data"]).To(HaveKey("access_token"))

				tokenForDeletedUser := response["data"].(map[string]interface{})["access_token"].(string)

				// Delete the user
				req, err = http.NewRequest(http.MethodDelete, baseURL+"/user", nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", "Bearer "+tokenForDeletedUser)

				resp, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

				// Now try to update the deleted user - should get 404
				updateInput := models.UpdateUserInput{
					Name: "Updated Name",
				}

				body, _ = json.Marshal(updateInput)
				req, err = http.NewRequest(http.MethodPatch, baseURL+"/user", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+tokenForDeletedUser)

				resp, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				response, err = decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("user not found"))
			})

			It("should be unauthorized for invalid token", func() {
				req, err := http.NewRequest(http.MethodPatch, baseURL+"/user", bytes.NewBuffer([]byte("{}")))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer invalid-token")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should be unauthorized for missing authorization header", func() {
				req, err := http.NewRequest(http.MethodPatch, baseURL+"/user", bytes.NewBuffer([]byte("{}")))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())

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

				body, _ := json.Marshal(userInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/signup", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["data"]).To(HaveKey("access_token"))

				newAccessToken := response["data"].(map[string]interface{})["access_token"].(string)

				updateInput := models.UpdateUserPasswordInput{
					OldPassword: "password123",
					NewPassword: "newpassword123",
				}

				body, _ = json.Marshal(updateInput)
				req, err = http.NewRequest(http.MethodPost, baseURL+"/user/password", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+newAccessToken)

				resp, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				response, err = decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("User password updated successfully"))
				Expect(response).NotTo(HaveKey("password"))

				// Verify new password works by trying to login
				loginInput := models.LoginInput{
					Email:    "passwordUpdate@example.com",
					Password: "newpassword123",
				}

				body, _ = json.Marshal(loginInput)
				req, err = http.NewRequest(http.MethodPost, baseURL+"/login", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})

			It("should trim whitespace from passwords and update successfully", func() {
				// Create a new user
				userInput := models.CreateUserInput{
					Email:    "passwordWhitespaceUpdate@example.com",
					Name:     "Test password whitespace update",
					Password: "password123",
				}

				body, _ := json.Marshal(userInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/signup", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["data"]).To(HaveKey("access_token"))

				newAccessToken := response["data"].(map[string]interface{})["access_token"].(string)

				updateInput := models.UpdateUserPasswordInput{
					OldPassword: "  password123  ",    // Old password with whitespace
					NewPassword: "  newpassword123  ", // New password with whitespace
				}

				body, _ = json.Marshal(updateInput)
				req, err = http.NewRequest(http.MethodPost, baseURL+"/user/password", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+newAccessToken)

				resp, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				response, err = decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("User password updated successfully"))

				// Verify new password works (trimmed version) by trying to login
				loginInput := models.LoginInput{
					Email:    "passwordWhitespaceUpdate@example.com",
					Password: "newpassword123", // Trimmed password should work
				}

				body, _ = json.Marshal(loginInput)
				req, err = http.NewRequest(http.MethodPost, baseURL+"/login", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("with invalid input", func() {
			It("should return unauthorized for incorrect old password", func() {
				updateInput := models.UpdateUserPasswordInput{
					OldPassword: "wrongpassword",
					NewPassword: "newpassword123",
				}

				body, _ := json.Marshal(updateInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/user/password", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return bad request for invalid JSON", func() {
				req, err := http.NewRequest(http.MethodPost, baseURL+"/user/password", bytes.NewBuffer([]byte("invalid json")))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for empty body", func() {
				req, err := http.NewRequest(http.MethodPost, baseURL+"/user/password", bytes.NewBuffer([]byte("")))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for whitespace-only old password", func() {
				updateInput := models.UpdateUserPasswordInput{
					OldPassword: "   ", // Only whitespace - will become empty after trimming
					NewPassword: "newpassword123",
				}

				body, _ := json.Marshal(updateInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/user/password", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(ContainSubstring("required"))
			})

			It("should return bad request for whitespace-only new password", func() {
				updateInput := models.UpdateUserPasswordInput{
					OldPassword: "password123",
					NewPassword: "   ", // Only whitespace - will become empty after trimming
				}

				body, _ := json.Marshal(updateInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/user/password", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(ContainSubstring("validation"))
			})

			It("should be unauthorized for invalid token", func() {
				req, err := http.NewRequest(http.MethodPost, baseURL+"/user/password", bytes.NewBuffer([]byte("{}")))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer invalid-token")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should be unauthorized for missing authorization header", func() {
				req, err := http.NewRequest(http.MethodPost, baseURL+"/user/password", bytes.NewBuffer([]byte("{}")))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())

				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return bad request for missing old password", func() {
				updateInput := models.UpdateUserPasswordInput{
					NewPassword: "newpassword123",
				}

				body, _ := json.Marshal(updateInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/user/password", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return bad request for missing new password", func() {
				updateInput := models.UpdateUserPasswordInput{
					OldPassword: "password",
				}

				body, _ := json.Marshal(updateInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/user/password", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+accessToken)

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))
			})

			It("should return not found when user doesn't exist", func() {
				// Create a new user
				userInput := models.CreateUserInput{
					Email:    "userToPasswordUpdateNotExists@example.com",
					Name:     "Test user for password update not exists",
					Password: "password123",
				}

				body, _ := json.Marshal(userInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/signup", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["data"]).To(HaveKey("access_token"))

				tokenForDeletedUser := response["data"].(map[string]interface{})["access_token"].(string)

				// Delete the user
				req, err = http.NewRequest(http.MethodDelete, baseURL+"/user", nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", "Bearer "+tokenForDeletedUser)

				resp, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

				// Now try to update password for the deleted user - should get 404
				updateInput := models.UpdateUserPasswordInput{
					OldPassword: "password123",
					NewPassword: "newpassword123",
				}

				body, _ = json.Marshal(updateInput)
				req, err = http.NewRequest(http.MethodPost, baseURL+"/user/password", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("Authorization", "Bearer "+tokenForDeletedUser)

				resp, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
				response, err = decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["message"]).To(Equal("user not found"))
			})
		})
	})

	Describe("DeleteUser", func() {
		Context("with valid token", func() {
			It("should delete user successfully", func() {
				// Create a new user
				userInput := models.CreateUserInput{
					Email:    "deleteUser@example.com",
					Name:     "Test user to delete",
					Password: "password123",
				}

				body, _ := json.Marshal(userInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/signup", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["data"]).To(HaveKey("access_token"))
				Expect(response["data"]).To(HaveKey("refresh_token"))
				Expect(response["data"]).To(HaveKey("user"))

				newAccessToken := response["data"].(map[string]interface{})["access_token"].(string)

				// Delete the user
				req, err = http.NewRequest(http.MethodDelete, baseURL+"/user", nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", "Bearer "+newAccessToken)

				resp, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

				// Verify user is deleted by trying to get user details
				req, err = http.NewRequest(http.MethodGet, baseURL+"/user", nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", "Bearer "+newAccessToken)

				resp, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusNotFound))
			})
		})

		Context("with invalid token", func() {
			It("should return unauthorized", func() {
				req, err := http.NewRequest(http.MethodDelete, baseURL+"/user", nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", "Bearer invalid-token")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})

			It("should return unauthorized for missing authorization header", func() {
				req, err := http.NewRequest(http.MethodDelete, baseURL+"/user", nil)
				Expect(err).NotTo(HaveOccurred())

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when user doesn't exist", func() {
			It("should return no content when user doesn't exist", func() {
				// Create a new user
				userInput := models.CreateUserInput{
					Email:    "userToDeleteNotExists@example.com",
					Name:     "Test user for delete not exists",
					Password: "password123",
				}

				body, _ := json.Marshal(userInput)
				req, err := http.NewRequest(http.MethodPost, baseURL+"/signup", bytes.NewBuffer(body))
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusCreated))
				response, err := decodeJSON(resp.Body)
				Expect(err).NotTo(HaveOccurred())
				Expect(response["data"]).To(HaveKey("access_token"))

				tokenForDeletedUser := response["data"].(map[string]interface{})["access_token"].(string)

				// Delete the user first time
				req, err = http.NewRequest(http.MethodDelete, baseURL+"/user", nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", "Bearer "+tokenForDeletedUser)

				resp, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))

				// Now try to delete the already deleted user - should get 404
				req, err = http.NewRequest(http.MethodDelete, baseURL+"/user", nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set("Authorization", "Bearer "+tokenForDeletedUser)

				resp, err = client.Do(req)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusNoContent))
			})
		})
	})
})
