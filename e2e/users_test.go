package e2e

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestCreateUser(t *testing.T) {
	cleanup, _, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatalf("failed to launch test program: %v", err)
	}
	t.Cleanup(cleanup)

	tests := []struct {
		name               string
		payload            string
		expectedStatusCode int
		expectedWrapper    string
		expectedResponse   map[string]any
	}{
		{
			name: "Successfully creates new user",
			payload: `{
					"name":     "Test User",
					"email":    "test@example.com",
					"password": "password"
				}`,
			expectedStatusCode: http.StatusCreated,
			expectedWrapper:    "user",
			expectedResponse: map[string]any{
				"user": map[string]any{
					"id":         1,
					"created_at": "2024-11-14T21:46:09Z",
					"name":       "Test User",
					"email":      "test@example.com",
					"activated":  false,
				},
			},
		},
		{
			name: "Creates a duplicate user",
			payload: `{
					"name":     "Test User",
					"email":    "test@example.com",
					"password": "password"
				}`,
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedWrapper:    "error",
			expectedResponse: map[string]any{
				"error": map[string]any{
					"email": "a user with this email address already exists",
				},
			},
		},
		{
			name:               "No body provided",
			payload:            `{}`,
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedWrapper:    "error",
			expectedResponse: map[string]any{
				"error": map[string]any{
					"email":    "must be provided",
					"name":     "must be provided",
					"password": "must be provided",
				},
			},
		},
		{
			name: "Inputting a bad email",
			payload: `{
					"name":     "Test User",
					"email":    "testexample.com",
					"password": "password"
				}`,
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedWrapper:    "error",
			expectedResponse: map[string]any{
				"error": map[string]any{
					"email": "must be a valid email address",
				},
			},
		},
		{
			name: "Too short of a password",
			payload: `{
					"name":     "Test User",
					"email":    "test@example.com",
					"password": "1234"
			}`,
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedWrapper:    "error",
			expectedResponse: map[string]any{
				"error": map[string]any{
					"password": "must be at least 8 bytes long",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestURL := fmt.Sprintf("http://localhost:%d/v1/users", 3001)
			statusCode, _, body := post(t, requestURL, strings.NewReader(tt.payload), nil)

			// Assertions
			assert.Equal(t, tt.expectedStatusCode, statusCode)
			assert.NotEmpty(t, body[tt.expectedWrapper])
			actualContent := body[tt.expectedWrapper].(map[string]any)
			expectedContent := tt.expectedResponse[tt.expectedWrapper].(map[string]any)
			for k, v := range actualContent {
				switch k {
				case "id":
					assert.NotEmpty(t, v)
				case "created_at":
					assert.NotEmpty(t, v)
				default:
					assert.Equal(t, expectedContent[k], v)

					// wait for emails to be sent
					var receivedCount int
					waitFor(t, func() bool {
						_, count := getEmail(t, "containing", "Welcome")
						if count >= 1 {
							receivedCount = count
							return true
						}
						return false
					})

					assert.GreaterOrEqual(t, receivedCount, 1)
				}
			}
		})
	}
}

// TODO Disabled these tests for now as activate user isnt a function of the app at the moment
//func TestActivateUser(t *testing.T) {
//	cleanup, _, err := LaunchTestProgram(port)
//	if err != nil {
//		t.Fatalf("failed to launch test program: %v", err)
//	}
//	t.Cleanup(cleanup)
//
//	_ = createUser(t)
//
//	tests := []struct {
//		name               string
//		setupPayload       func(token string) string
//		expectedStatusCode int
//		expectedWrapper    string
//		expectedResponse   map[string]any
//	}{
//		{
//			name: "Successfully activate user",
//			setupPayload: func(token string) string {
//				return fmt.Sprintf(`{"token":"%s"}`, token)
//			},
//			expectedStatusCode: http.StatusOK,
//			expectedWrapper:    "user",
//			expectedResponse: map[string]any{
//				"user": map[string]any{
//					"activated": true,
//				},
//			},
//		},
//		{
//			name: "User did not input a token",
//			setupPayload: func(token string) string {
//				return `{"token":""}`
//			},
//			expectedStatusCode: http.StatusUnprocessableEntity,
//			expectedWrapper:    "error",
//			expectedResponse: map[string]any{
//				"error": map[string]any{
//					"token": "must be provided",
//				},
//			},
//		},
//		{
//			name: "Token is incorrect / expired",
//			setupPayload: func(token string) string {
//				return `{"token":"ASDJKLEPOIURERFJDKSLAIEJG1"}`
//			},
//			expectedStatusCode: http.StatusUnprocessableEntity,
//			expectedWrapper:    "error",
//			expectedResponse: map[string]any{
//				"error": map[string]any{
//					"token": "token not found",
//				},
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			// Fetch activation token from email
//			var token string
//			waitFor(t, func() bool {
//				body, _ := getEmail(t, "containing", "token")
//				token = extractToken(t, body)
//				return token != ""
//			})
//
//			// Activate the user
//			requestURL := fmt.Sprintf("http://localhost:%d/v1/users/activated", 3001)
//			payload := tt.setupPayload(token)
//			statusCode, _, body := put(t, requestURL, strings.NewReader(payload))
//
//			// Assertions
//			assert.Equal(t, tt.expectedStatusCode, statusCode)
//			assert.NotEmpty(t, body[tt.expectedWrapper])
//
//			actualContent := body[tt.expectedWrapper].(map[string]any)
//			expectedContent := tt.expectedResponse[tt.expectedWrapper].(map[string]any)
//
//			for key, value := range expectedContent {
//				assert.Equal(t, value, actualContent[key], "Mismatch for key: %s", key)
//			}
//		})
//	}
//}

func TestVerifyUser(t *testing.T) {
	cleanup, _, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatalf("failed to launch test program: %v", err)
	}
	t.Cleanup(cleanup)

	_ = createUser(t)

	resp := authenticateUser(t, "test@example.com", "password")

	token := resp["authentication_token"].(map[string]any)["token"].(string)

	tests := []struct {
		name               string
		token              string
		expectedStatusCode int
		expectedWrapper    string
		expectedResponse   map[string]any
	}{
		{
			name:               "Succesfully verifies user",
			token:              token,
			expectedStatusCode: http.StatusOK,
			expectedWrapper:    "user",
			expectedResponse: map[string]any{
				"user": map[string]any{
					"id":        1,
					"name":      "Test User",
					"email":     "test@example.com",
					"activated": false,
				},
			},
		},
		{
			name:               "No Token provided",
			token:              "",
			expectedStatusCode: http.StatusUnauthorized,
			expectedWrapper:    "",
			expectedResponse: map[string]any{
				"error": "invalid or missing authentication token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", tt.token),
			}
			requestURL := fmt.Sprintf("http://localhost:%d/v1/users/verify", 3001)
			statusCode, _, returnedBody := get(t, requestURL, headers)
			assert.Equal(t, tt.expectedStatusCode, statusCode)

			if tt.expectedWrapper == "" {
				assert.Contains(t, returnedBody["error"], tt.expectedResponse["error"])
				return
			}

			assert.NotEmpty(t, returnedBody[tt.expectedWrapper])
			actualContent := returnedBody[tt.expectedWrapper].(map[string]any)
			expectedContent := tt.expectedResponse[tt.expectedWrapper].(map[string]any)

			for k, v := range actualContent {
				switch k {
				case "id":
					assert.NotEmpty(t, v)
				case "created_at":
					assert.NotEmpty(t, v)
				default:
					assert.Equal(t, expectedContent[k], v)
				}
			}
		})
	}
}

func TestResetPassword(t *testing.T) {
	cleanup, _, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatalf("failed to launch test program: %v", err)
	}
	t.Cleanup(cleanup)

	_ = createUser(t)

	t.Run("Succesfully Resets Password", func(t *testing.T) {

		// Check that you can log in with the old password
		requestURL := fmt.Sprintf("http://localhost:%d/v1/tokens/authentication", 3001)
		requestBody := `{"email": "test@example.com", "password": "password"}`
		statusCode, _, _ := post(t, requestURL, strings.NewReader(requestBody), nil)
		assert.Equal(t, http.StatusCreated, statusCode)

		// Send the request to reset
		requestURL = fmt.Sprintf("http://localhost:%d/v1/tokens/password-reset", 3001)
		requestBody = `{"email": "test@example.com"}`

		statusCode, _, returnedBody := post(t, requestURL, strings.NewReader(requestBody), nil)
		expectedResponse := map[string]any{
			"message": "an email will be sent to you containing password reset instructions",
		}

		assert.Equal(t, http.StatusCreated, statusCode)
		assert.Equal(t, expectedResponse, returnedBody)

		// Get the reset token from the email
		var token string
		waitFor(t, func() bool {
			body, _ := getEmail(t, "containing", "password%20reset%20token")
			token = extractToken(t, body)
			return token != ""
		})

		// Send a request to the reset password
		requestURL = fmt.Sprintf("http://localhost:%d/v1/users/password", 3001)
		requestBody = fmt.Sprintf(`{"password": "password2", "token": "%s"}`, token)

		statusCode, _, returnedBody = put(t, requestURL, strings.NewReader(requestBody))
		assert.Equal(t, http.StatusOK, statusCode)
		expectedResponse = map[string]any{
			"message": "your password was successfully reset",
		}
		assert.Equal(t, expectedResponse, returnedBody)

		// Check that you cannot log in with the old password
		requestURL = fmt.Sprintf("http://localhost:%d/v1/tokens/authentication", 3001)
		requestBody = `{"email": "test@example.com", "password": "password"}`
		statusCode, _, _ = post(t, requestURL, strings.NewReader(requestBody), nil)
		assert.Equal(t, http.StatusUnauthorized, statusCode)

		//Check that you can log in with the new password
		requestURL = fmt.Sprintf("http://localhost:%d/v1/tokens/authentication", 3001)
		requestBody = `{"email": "test@example.com", "password": "password2"}`
		statusCode, _, _ = post(t, requestURL, strings.NewReader(requestBody), nil)
		assert.Equal(t, http.StatusCreated, statusCode)
	})
}
