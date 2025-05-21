// tests/helpers.go
package tests

import (
	"context"
	controllers "expenses/internal/api/controller"
	"expenses/internal/models"
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TestEnv struct {
	DB        *pgxpool.Pool
	Container testcontainers.Container
	AuthCtrl  *controllers.AuthController
	TestUser  models.CreateUserInput
}

func SetupTestEnv(t *testing.T) *TestEnv {
	ctx := context.Background()

	// Define container configuration
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort(nat.Port("5432/tcp")),
	}

	// Start the container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get the container's host and port
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Set up environment variables for the test
	os.Setenv("DB_HOST", host)
	os.Setenv("DB_PORT", port.Port())
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_SCHEMA", "dev")
	os.Setenv("SERVER_PORT", "8080")
	os.Setenv("JWT_SECRET", "test-secret")

	dbURL := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())
	testDB, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err)
	_, err = testDB.Exec(ctx, fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", os.Getenv("DB_SCHEMA")))
	require.NoError(t, err)
	err = runMigrations()
	require.NoError(t, err)

	// Initialize the auth controller
	authCtrl := controllers.NewAuthController(testDB)

	// Set up test user data
	testUser := models.CreateUserInput{
		Email:    "test@example.com",
		Password: "testpass123",
		Name:     "Test User",
	}

	return &TestEnv{
		DB:        testDB,
		Container: container,
		AuthCtrl:  authCtrl,
		TestUser:  testUser,
	}
}

func runMigrations() error {
	fmt.Println("Running migrations")
	cmd := exec.Command("just", "install")
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run migrations during 'just install': %v", err)
	}
	cmd = exec.Command("just", "db-upgrade")
	cmd.Env = append(os.Environ(),
		"DB_SCHEMA=dev",
		"DB_HOST="+os.Getenv("DB_HOST"),
		"DB_PORT="+os.Getenv("DB_PORT"),
		"DB_USER="+os.Getenv("DB_USER"),
		"DB_PASSWORD="+os.Getenv("DB_PASSWORD"),
		"DB_NAME="+os.Getenv("DB_NAME"),
		"DB_SSL_MODE=disable",
	)
	output, err := cmd.CombinedOutput()
	fmt.Println("Migration output: ", string(output))
	if err != nil {
		return fmt.Errorf("failed to run migrations during 'just db-upgrade': %v", err)
	}
	return nil
}

func TeardownTestEnv(t *testing.T, env *TestEnv) {
	// Clean up environment variables
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_USER")
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("DB_NAME")
	os.Unsetenv("DB_SCHEMA")
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("JWT_SECRET")

	if env.DB != nil {
		env.DB.Close()
	}
	if env.Container != nil {
		err := env.Container.Terminate(context.Background())
		require.NoError(t, err)
	}
}
