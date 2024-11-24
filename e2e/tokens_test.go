package e2e

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestAuthenticateUser(t *testing.T) {
	cleanup, _, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatalf("failed to launch test program: %v", err)
	}
	t.Cleanup(cleanup)
	_ = createUser(t)

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
			requestURL := fmt.Sprintf("http://localhost:%d/v1/tokens/authentication", 3001)
			statusCode, _, body := post(t, requestURL, strings.NewReader(tt.payload))
			assert.Equal(t, tt.expectedStatusCode, statusCode)
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
