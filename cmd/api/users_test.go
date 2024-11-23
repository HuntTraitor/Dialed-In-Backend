package main

import (
	"encoding/json"
	"github.com/hunttraitor/dialed-in-backend/internal/mocks"
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
					// Check that an email has been sent and a token has been created
					assert.Equal(t, 1, app.mailer.(*mocks.MockMailer).SendCalledCount)
					assert.Equal(t, 1, app.models.Tokens.(*mocks.MockTokenModel).TokenCreated)
				}
			}
		})
	}
}

func TestActivateUser(t *testing.T) {
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
			name:               "Successfully activates user",
			payload:            `{"token": "ASDJKLEPOIURERFJDKSLAIEJG1"}`,
			expectedStatusCode: http.StatusOK,
			expectedWrapper:    "user",
			expectedResponse: map[string]any{
				"user": map[string]any{
					"id":        "1",
					"name":      "Test User",
					"email":     "test@example.com",
					"activated": true,
				},
			},
		},
		{
			name:               "Nil token provided returns error",
			payload:            `{"token": ""}`,
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedWrapper:    "error",
			expectedResponse: map[string]any{
				"error": map[string]any{
					"token": "must be provided",
				},
			},
		},
		{
			name:               "Too short of a token provided returns error",
			payload:            `{"token": "shorttoken"}`,
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedWrapper:    "error",
			expectedResponse: map[string]any{
				"error": map[string]any{
					"token": "must be 26 bytes long",
				},
			},
		},
		{
			name:               "User associated with token not found",
			payload:            `{"token": "ASDJKLEPOIURERFJDKSLAIEJG2"}`,
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedWrapper:    "error",
			expectedResponse: map[string]any{
				"error": map[string]any{
					"token": "token not found",
				},
			},
		},
		{
			name:               "Edit conflict when updating user",
			payload:            `{"token": "ASDJKLEPOIURERFJDKSLAIEJG3"}`,
			expectedStatusCode: http.StatusConflict,
			expectedWrapper:    "",
			expectedResponse: map[string]any{
				"error": "unable to update the record due to an edit conflict, please try again",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, _, returnedBody := ts.put(t, "/v1/users/activated", strings.NewReader(tt.payload))
			assert.Equal(t, tt.expectedStatusCode, statusCode)

			var body map[string]any
			err := json.Unmarshal([]byte(returnedBody), &body)
			if err != nil {
				t.Fatal(err)
			}

			if tt.expectedWrapper == "" {
				assert.Contains(t, body["error"], tt.expectedResponse["error"])
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
