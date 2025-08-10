package e2e

import (
	"fmt"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetAllMethods(t *testing.T) {
	cleanup, _, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatalf("failed to launch test program: %v", err)
	}
	t.Cleanup(cleanup)

	tests := []struct {
		name               string
		expectedStatusCode int
		expectedMethods    []data.Method
	}{
		{
			name:               "Successfully gets methods",
			expectedStatusCode: http.StatusOK,
			expectedMethods: []data.Method{
				{
					ID:        1,
					Name:      "V60",
					CreatedAt: "2025-01-25 00:28:23 +00:00",
				},
				{
					ID:        2,
					Name:      "Hario Switch",
					CreatedAt: "2025-01-25 00:28:23 +00:00",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestURL := fmt.Sprintf("http://localhost:%d/v1/methods", 3001)
			statusCode, _, returnedBody := get(t, requestURL, nil)
			assert.Equal(t, tt.expectedStatusCode, statusCode)
			methods := returnedBody["methods"].([]any)
			for i, item := range methods {
				method := item.(map[string]any)
				assert.Equal(t, tt.expectedMethods[i].Name, method["name"])
			}
		})
	}
}

func TestGetOneMethod(t *testing.T) {
	cleanup, _, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatalf("failed to launch test program: %v", err)
	}
	t.Cleanup(cleanup)

	mockMethod := data.Method{
		ID:   1,
		Name: "V60",
	}

	tests := []struct {
		name               string
		expectedStatusCode int
		expectedMethod     data.Method
		expectedError      map[string]any
	}{
		{
			name:               "Successfully gets method",
			expectedStatusCode: http.StatusOK,
			expectedMethod:     mockMethod,
			expectedError:      nil,
		},
		{
			name:               "Method not found",
			expectedStatusCode: http.StatusNotFound,
			expectedMethod: data.Method{
				ID: 0,
			},
			expectedError: map[string]any{
				"error": "The requested resource could not be found.",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestURL := fmt.Sprintf("http://localhost:%d/v1/methods/%d", 3001, tt.expectedMethod.ID)
			statusCode, _, returnedBody := get(t, requestURL, nil)
			assert.Equal(t, tt.expectedStatusCode, statusCode)
			if tt.expectedError == nil {
				returnedMethod := returnedBody["method"].(map[string]any)
				assert.Equal(t, tt.expectedMethod.Name, returnedMethod["name"])
			} else {
				assert.Equal(t, tt.expectedError, returnedBody)
			}
		})
	}
}
