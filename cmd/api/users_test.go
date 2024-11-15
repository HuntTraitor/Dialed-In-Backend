package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestCreateUser(t *testing.T) {
	app := newTestApplication()
	ts := newTestServer(app.routes())
	defer ts.Close()

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
					"name":     "Duplicate User",
					"email":    "dupe@example.com",
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
			name: "Internal server error",
			payload: `{
					"name":     "Duplicate User",
					"email":    "error@example.com",
					"password": "password"
				}`,
			expectedStatusCode: http.StatusInternalServerError,
			expectedWrapper:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, _, returnedBody := ts.post(t, "/v1/users", strings.NewReader(tt.payload))

			var body map[string]any
			err := json.Unmarshal([]byte(returnedBody), &body)
			if err != nil {
				t.Fatal(err)
			}

			// Assertions
			assert.Equal(t, tt.expectedStatusCode, statusCode)

			if tt.expectedWrapper == "" {
				assert.Contains(t, body["error"], "internal server error")
				return
			}

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
