package e2e

import (
	"encoding/json"
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
					Name:      "Mock Method 1",
					Img:       "https://example.com/img1.png",
					CreatedAt: "2025-01-25 00:28:23 +00:00",
				},
				{
					ID:        2,
					Name:      "Mock Method 2",
					Img:       "https://example.com/img2.png",
					CreatedAt: "2025-01-25 00:28:23 +00:00",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestURL := fmt.Sprintf("http://localhost:%d/v1/methods", 3001)
			statusCode, _, returnedBody := get(t, requestURL, nil)
			var body map[string]any
			err = json.Unmarshal([]byte(returnedBody), &body)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, tt.expectedStatusCode, statusCode)
			methods := body["methods"].([]any)
			// TODO add seeds to test database script for e2e tests.
			for i, item := range methods {
				method := item.(map[string]any)
				assert.Equal(t, tt.expectedMethods[i].Name, method["name"])
				assert.Equal(t, tt.expectedMethods[i].Img, method["img"])
			}
		})
	}

}
