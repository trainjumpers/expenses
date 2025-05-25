package controller

import (
	"bytes"
	"encoding/json"
	"expenses/internal/models"
	"fmt"
	"net/http"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestController(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Controller Suite")
}

var (
	client       *http.Client
	baseURL      string
	accessToken  string
	refreshToken string
)

var _ = BeforeSuite(func() {
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	client = &http.Client{}
	baseURL = "http://localhost:" + port + "/api/v1"

	// Login to a test user
	loginInput := models.LoginInput{
		Email:    "test1@example.com",
		Password: "password",
	}

	body, _ := json.Marshal(loginInput)
	req, err := http.NewRequest(http.MethodPost, baseURL+"/login", bytes.NewBuffer(body))
	Expect(err).NotTo(HaveOccurred())
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	Expect(err).NotTo(HaveOccurred())
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
	// Expect(resp.StatusCode).To(Equal(http.StatusOK))
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	fmt.Println(response)
	Expect(err).NotTo(HaveOccurred())
	Expect(response["message"]).To(Equal("User logged in successfully"))

	accessToken = response["data"].(map[string]interface{})["access_token"].(string)
	refreshToken = response["data"].(map[string]interface{})["refresh_token"].(string)
})
