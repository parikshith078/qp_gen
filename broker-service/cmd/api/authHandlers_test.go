package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
	"github.com/parikshith078/qp_gen/broker/internal/db"
	"github.com/parikshith078/qp_gen/broker/internal/db/sqlc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*postgres.PostgresContainer, *Config) {
	ctx := context.Background()
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("test"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		testcontainers.WithWaitStrategy(
			wait.
				ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	require.NoError(t, err)

	// Get connection string
	connString, err := postgresContainer.ConnectionString(ctx)
	require.NoError(t, err)
	connString = connString + "&sslmode=disable"

	// Connect to database
	conn, err := pgx.Connect(ctx, connString)
	require.NoError(t, err)

	// Run migrations
	err = db.RunMigrations(connString, "file://../../migrations")
	require.NoError(t, err)

	// Initialize validator
	validate = validator.New(validator.WithRequiredStructEnabled())

	// Create app config
	db := sqlc.New(conn)
	require.NotNil(t, db)
	app := &Config{
		Db: db,
	}

	return postgresContainer, app
}

func TestRegisterRoute(t *testing.T) {
	container, app := setupTestDB(t)
	defer func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	tests := []struct {
		name           string
		requestBody    map[string]any
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Valid registration",
			requestBody: map[string]interface{}{
				"name":     "Test User",
				"email":    "test@example.com",
				"username": "testuser",
				"password": "password123",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name: "Missing required field",
			requestBody: map[string]interface{}{
				"name":     "Test User",
				"email":    "test@example.com",
				"password": "password123",
				// username missing
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Invalid email",
			requestBody: map[string]interface{}{
				"name":     "Test User",
				"email":    "invalid-email",
				"username": "testuser",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Empty password",
			requestBody: map[string]interface{}{
				"name":     "Test User",
				"email":    "test@example.com",
				"username": "testuser",
				"password": "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert request body to JSON
			jsonBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			app.Register(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Parse response
			var response jsonReponse
			err = json.NewDecoder(rr.Body).Decode(&response)
			require.NoError(t, err)

			// Verify response structure
			assert.Equal(t, tt.expectedError, response.Error)

			if !tt.expectedError && rr.Code == http.StatusOK {
				// Verify user was actually created in database
				user, ok := response.Data.(map[string]interface{})
				require.True(t, ok)

				// Verify the returned user has the expected fields
				assert.NotEmpty(t, user["id"])
				assert.Equal(t, tt.requestBody["email"], user["email"])
				assert.Equal(t, tt.requestBody["username"], user["username"])
				assert.Equal(t, tt.requestBody["name"], user["name"])

				// Verify password is not returned
				_, hasPassword := user["password_hash"]
				assert.False(t, hasPassword)

				// Optional: Verify user exists in database
				ctx := context.Background()
				dbUser, err := app.Db.GetUserByEmail(ctx, tt.requestBody["email"].(string))
				assert.NoError(t, err)
				assert.Equal(t, tt.requestBody["email"], dbUser.Email)
			}
		})
	}
}
