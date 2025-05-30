package controller_test

import (
	"bytes"
	"encoding/json"
	"expenses/internal/models"
	"expenses/internal/server"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

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
	accessToken2  string
	refreshToken2 string
)

var _ = BeforeSuite(func() {
	server.StartAsync()
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	client = &http.Client{}
	baseURL = "http://localhost:" + port
	healthCheckSuccess := false

	for i := 0; i < 10; i++ {
		resp, err := http.Get(baseURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			healthCheckSuccess = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	if !healthCheckSuccess {
		panic("could not connect to server")
	}

	baseURL += "/api/v1"

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

	loginInput2 := models.LoginInput{
		Email:    "test3@example.com",
		Password: "password",
	}
	body2, _ := json.Marshal(loginInput2)
	req2, err := http.NewRequest(http.MethodPost, baseURL+"/login", bytes.NewBuffer(body2))
	Expect(err).NotTo(HaveOccurred())
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := client.Do(req2)
	Expect(err).NotTo(HaveOccurred())
	defer resp2.Body.Close()
	Expect(resp2.StatusCode).To(Equal(http.StatusOK))

	var response2 map[string]interface{}
	err = json.NewDecoder(resp2.Body).Decode(&response2)
	Expect(err).NotTo(HaveOccurred())
	Expect(response2["message"]).To(Equal("User logged in successfully"))

	accessToken2 = response2["data"].(map[string]interface{})["access_token"].(string)
	refreshToken2 = response2["data"].(map[string]interface{})["refresh_token"].(string)
})

// decodeJSON is a helper function to decode JSON from any io.Reader
func decodeJSON(reader io.Reader) (map[string]interface{}, error) {
	var response map[string]interface{}
	err := json.NewDecoder(reader).Decode(&response)
	return response, err
}
