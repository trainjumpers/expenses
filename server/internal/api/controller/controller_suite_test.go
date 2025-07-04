package controller_test

import (
	"encoding/json"
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
	testHelperUnauthenticated *TestHelper
	testHelperUser1           *TestHelper
	testHelperUser2           *TestHelper
	testHelperUser3           *TestHelper
	baseURL                   string
)

var _ = BeforeSuite(func() {
	server.StartAsync()
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

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
	testHelperUser1 = NewTestHelper(baseURL)
	testHelperUser2 = NewTestHelper(baseURL)
	testHelperUser3 = NewTestHelper(baseURL)
	testHelperUnauthenticated = NewTestHelper(baseURL)

	// Login test users (each helper manages its own cookies)
	testHelperUser1.Login("test1@example.com", "password")
	testHelperUser2.Login("test2@example.com", "password")
	testHelperUser3.Login("test3@example.com", "password")
})

// decodeJSON is a helper function to decode JSON from any io.Reader
func decodeJSON(reader io.Reader) (map[string]interface{}, error) {
	var response map[string]interface{}
	err := json.NewDecoder(reader).Decode(&response)
	return response, err
}
