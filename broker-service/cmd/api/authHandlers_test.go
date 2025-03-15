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

func TestLoginRoute(t *testing.T) {
	container, app := setupTestDB(t)
	defer func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	// Create a test user first
	ctx := context.Background()
	password := "testPassword123"
	passwordHash, err := HashPassword(password)
	require.NoError(t, err)

	testUser, err := app.Db.CreateUser(ctx, sqlc.CreateUserParams{
		Name:         "Test User",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: passwordHash,
	})
	require.NoError(t, err)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  bool
		checkCookies   bool
		setupFunc      func() // For any additional setup needed
	}{
		{
			name: "Valid login",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "testPassword123",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
			checkCookies:   true,
		},
		{
			name: "Invalid email",
			requestBody: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "testPassword123",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
			checkCookies:   false,
		},
		{
			name: "Wrong password",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
			checkCookies:   false,
		},
		{
			name: "Missing email",
			requestBody: map[string]interface{}{
				"password": "testPassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			checkCookies:   false,
		},
		{
			name: "Missing password",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			checkCookies:   false,
		},
		{
			name: "Invalid email format",
			requestBody: map[string]interface{}{
				"email":    "invalid-email",
				"password": "testPassword123",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			checkCookies:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}

			// Convert request body to JSON
			jsonBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			app.Login(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Parse response
			var response jsonReponse
			err = json.NewDecoder(rr.Body).Decode(&response)
			require.NoError(t, err)

			// Verify response structure
			assert.Equal(t, tt.expectedError, response.Error)

			if !tt.expectedError {
				// Check if user data is returned correctly
				user, ok := response.Data.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, testUser.Email, user["email"])
				assert.Equal(t, testUser.Username, user["username"])
				assert.Equal(t, testUser.Name, user["name"])

				// Verify password hash is not included
				_, hasPassword := user["password_hash"]
				assert.False(t, hasPassword)
			}

			if tt.checkCookies {
				// Check for session cookie
				cookies := rr.Result().Cookies()
				var sessionCookie, csrfCookie *http.Cookie
				for _, cookie := range cookies {
					switch cookie.Name {
					case "session_token":
						sessionCookie = cookie
					case "csrf_token":
						csrfCookie = cookie
					}
				}

				// Verify session cookie
				require.NotNil(t, sessionCookie)
				assert.True(t, sessionCookie.HttpOnly)
				assert.NotEmpty(t, sessionCookie.Value)
				assert.True(t, sessionCookie.Expires.After(time.Now()))

				// Verify CSRF cookie
				require.NotNil(t, csrfCookie)
				assert.False(t, csrfCookie.HttpOnly)
				assert.NotEmpty(t, csrfCookie.Value)
				assert.True(t, csrfCookie.Expires.After(time.Now()))

				// Verify session was stored in database
				session, err := app.Db.GetSessionTokenByToken(ctx, sessionCookie.Value)
				require.NoError(t, err)
				assert.Equal(t, testUser.ID, session.UserID)

				// Verify CSRF token was stored in database
				csrf, err := app.Db.GetCsrfTokenBySessionID(ctx, session.ID)
				require.NoError(t, err)
				assert.Equal(t, csrfCookie.Value, csrf.Token)
			}
		})
	}
}

func TestLogout(t *testing.T) {
	container, app := setupTestDB(t)
	defer func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	// Create a test user and session
	ctx := context.Background()
	password := "testPassword123"
	passwordHash, err := HashPassword(password)
	require.NoError(t, err)

	testUser, err := app.Db.CreateUser(ctx, sqlc.CreateUserParams{
		Name:         "Test User",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: passwordHash,
	})
	require.NoError(t, err)

	sessionToken := "test-session-token"
	csrfToken := "test-csrf-token"
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create session and CSRF tokens
	sessionDoc, err := app.Db.CreateSessionToken(ctx, sqlc.CreateSessionTokenParams{
		UserID:    testUser.ID,
		Token:     sessionToken,
		ExpiresAt: expiresAt,
	})
	require.NoError(t, err)

	_, err = app.Db.CreateCsrfToken(ctx, sqlc.CreateCsrfTokenParams{
		SessionID: sessionDoc.ID,
		Token:     csrfToken,
		ExpiresAt: expiresAt,
	})
	require.NoError(t, err)

	tests := []struct {
		name           string
		setupCookies   bool
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Successful logout",
			setupCookies:   true,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "Missing session token",
			setupCookies:   false,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/logout", nil)

			if tt.setupCookies {
				req.AddCookie(&http.Cookie{
					Name:  "session_token",
					Value: sessionToken,
				})
			}

			rr := httptest.NewRecorder()
			app.Logout(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response jsonReponse
			err = json.NewDecoder(rr.Body).Decode(&response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedError, response.Error)

			if !tt.expectedError {
				// Verify cookies are cleared
				cookies := rr.Result().Cookies()
				for _, cookie := range cookies {
					assert.True(t, cookie.Expires.Before(time.Now()))
					assert.Empty(t, cookie.Value)
				}

				// Verify session is removed from database
				_, err = app.Db.GetSessionTokenByToken(ctx, sessionToken)
				assert.Error(t, err)
			}
		})
	}
}

func testAuthMiddleware(t *testing.T) {
	container, app := setupTestDB(t)
	defer func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

	// Create a test user and session
	ctx := context.Background()
	testUser, err := app.Db.CreateUser(ctx, sqlc.CreateUserParams{
		Name:         "Test User",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hash",
	})
	require.NoError(t, err)

	validSessionToken := "valid-session-token"
	validCsrfToken := "valid-csrf-token"
	expiredSessionToken := "expired-session-token"
	expiredCsrfToken := "expired-csrf-token"

	validExpiresAt := time.Now().Add(24 * time.Hour)
	expiredExpiresAt := time.Now().Add(-24 * time.Hour)

	// Create valid session and CSRF tokens
	validSession, err := app.Db.CreateSessionToken(ctx, sqlc.CreateSessionTokenParams{
		UserID:    testUser.ID,
		Token:     validSessionToken,
		ExpiresAt: validExpiresAt,
	})
	require.NoError(t, err)

	_, err = app.Db.CreateCsrfToken(ctx, sqlc.CreateCsrfTokenParams{
		SessionID: validSession.ID,
		Token:     validCsrfToken,
		ExpiresAt: validExpiresAt,
	})
	require.NoError(t, err)

	// Create expired session and CSRF tokens
	expiredSession, err := app.Db.CreateSessionToken(ctx, sqlc.CreateSessionTokenParams{
		UserID:    testUser.ID,
		Token:     expiredSessionToken,
		ExpiresAt: expiredExpiresAt,
	})
	require.NoError(t, err)

	_, err = app.Db.CreateCsrfToken(ctx, sqlc.CreateCsrfTokenParams{
		SessionID: expiredSession.ID,
		Token:     expiredCsrfToken,
		ExpiresAt: expiredExpiresAt,
	})
	require.NoError(t, err)

	tests := []struct {
		name           string
		sessionToken   string
		csrfToken      string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "Valid tokens",
			sessionToken:   validSessionToken,
			csrfToken:      validCsrfToken,
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "Missing session token",
			sessionToken:   "",
			csrfToken:      validCsrfToken,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:           "Missing CSRF token",
			sessionToken:   validSessionToken,
			csrfToken:      "",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:           "Expired session",
			sessionToken:   expiredSessionToken,
			csrfToken:      expiredCsrfToken,
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:           "Invalid CSRF token",
			sessionToken:   validSessionToken,
			csrfToken:      "invalid-token",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test handler that will be wrapped by the middleware
			nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// If we reach here, middleware passed
				app.writeJSON(w, http.StatusOK, jsonReponse{
					Error:   false,
					Message: "success",
				})
			})

			// Create request with necessary cookies and headers
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.sessionToken != "" {
				req.AddCookie(&http.Cookie{
					Name:  "session_token",
					Value: tt.sessionToken,
				})
			}
			if tt.csrfToken != "" {
				req.Header.Set("X-CSRF-Token", tt.csrfToken)
			}

			rr := httptest.NewRecorder()

			// Apply middleware
			handler := app.AuthMiddleware(nextHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			var response jsonReponse
			err = json.NewDecoder(rr.Body).Decode(&response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedError, response.Error)

			if !tt.expectedError {
				// Verify user_id is added to context
				userId, err := GetUserIDFromContext(req.Context())
				assert.NoError(t, err)
				assert.NotNil(t, userId)
				// assert.Equal(t, testUser.ID.String(), userId.String())
			}
		})
	}
}
