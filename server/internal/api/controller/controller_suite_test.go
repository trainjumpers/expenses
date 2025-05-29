package controller

import (
	"bytes"
	"encoding/json"
	"expenses/internal/models"
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
	client        *http.Client
	baseURL       string
	accessToken   string
	refreshToken  string
	accessToken1  string
	refreshToken1 string
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

	Expect(resp.StatusCode).To(Equal(http.StatusOK))
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	Expect(err).NotTo(HaveOccurred())
	Expect(response["message"]).To(Equal("User logged in successfully"))

	accessToken = response["data"].(map[string]interface{})["access_token"].(string)
	refreshToken = response["data"].(map[string]interface{})["refresh_token"].(string)

	loginInput1 := models.LoginInput{
		Email:    "test2@example.com",
		Password: "password",
	}
	body1, _ := json.Marshal(loginInput1)
	req1, err := http.NewRequest(http.MethodPost, baseURL+"/login", bytes.NewBuffer(body1))
	Expect(err).NotTo(HaveOccurred())
	req1.Header.Set("Content-Type", "application/json")
	resp1, err := client.Do(req1)
	Expect(err).NotTo(HaveOccurred())
	defer resp1.Body.Close()
	Expect(resp1.StatusCode).To(Equal(http.StatusOK))

	var response1 map[string]interface{}
	err = json.NewDecoder(resp1.Body).Decode(&response1)
	Expect(err).NotTo(HaveOccurred())
	Expect(response1["message"]).To(Equal("User logged in successfully"))

	accessToken1 = response1["data"].(map[string]interface{})["access_token"].(string)
	refreshToken1 = response1["data"].(map[string]interface{})["refresh_token"].(string)
})
