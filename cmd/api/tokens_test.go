package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestCreateAuthenticationHandler(t *testing.T) {
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
			name:               "Successfully authenticates user",
			payload:            `{"email": "test@example.com", "password": "password"}`,
			expectedStatusCode: http.StatusCreated,
			expectedWrapper:    "authentication_token",
		},
		//{
		//	name:               "Fails to authenticate unactivated user",
		//	payload:            `{"email": "notactivated@example.com", "password": "password"}`,
		//	expectedStatusCode: http.StatusUnauthorized,
		//	expectedWrapper:    "",
		//	expectedResponse: map[string]any{
		//		"error": "your user account must be verified to login, please verify your account by checking your email",
		//	},
		//},
		{
			name:               "Incorrect email returns error",
			payload:            `{"email": "notfound@example.com", "password": "password"}`,
			expectedStatusCode: http.StatusNotFound,
			expectedWrapper:    "",
			expectedResponse: map[string]any{
				"error": "The requested resource could not be found.",
			},
		},
		{
			name:               "Incorrect password returns error",
			payload:            `{"email": "test@example.com", "password": "incorrect"}`,
			expectedStatusCode: http.StatusUnauthorized,
			expectedWrapper:    "",
			expectedResponse: map[string]any{
				"error": "invalid authentication credentials",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, _, returnedBody := ts.post(t, "/v1/tokens/authentication", strings.NewReader(tt.payload))
			assert.Equal(t, tt.expectedStatusCode, statusCode)

			var body map[string]any
			err := json.Unmarshal([]byte(returnedBody), &body)
			if err != nil {
				t.Fatal(err)
			}

			if tt.expectedWrapper == "authentication_token" {
				actualContent := body[tt.expectedWrapper].(map[string]any)
				assert.NotEmpty(t, actualContent["token"])
				assert.NotEmpty(t, actualContent["expiry"])
			} else {
				assert.Equal(t, tt.expectedResponse["error"], body["error"])
			}
		})
	}

}
