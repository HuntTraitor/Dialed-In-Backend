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
			statusCode, _, body := post(t, requestURL, strings.NewReader(tt.payload))

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
				}
			}
		})
	}
}
