package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestCreateRecipe(t *testing.T) {
	cleanup, _, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatalf("failed to launch test program: %v", err)
	}
	t.Cleanup(cleanup)

	authenticateBody := authenticateUser(t, "hunter@gmail.com", "password")
	token := authenticateBody["authentication_token"].(map[string]any)["token"].(string)

	// create a new coffee to attach the recipe to
	insertedCoffee := data.Coffee{
		Name:        "Test Coffee",
		Region:      "Test Region",
		Process:     "Test Process",
		Description: "Test Description",
		Img:         "Test Image",
	}
	createCoffee(t, token, insertedCoffee, []byte(insertedCoffee.Img))

	correctPayload := map[string]any{
		"coffee_id": 1,
		"method_id": 1,
		"info": map[string]any{
			"name":     "Test Name",
			"grams_in": 20,
			"ml_out":   320,
			"phases": []map[string]any{
				{
					"open":   true,
					"time":   45,
					"amount": 160,
				},
				{
					"open":   false,
					"time":   75,
					"amount": 160,
				},
				{
					"open":   true,
					"time":   60,
					"amount": 0,
				},
			},
		},
	}

	tests := []struct {
		name               string
		payload            map[string]any
		expectedStatusCode int
		expectedResponse   map[string]any
		expectedError      map[string]any
	}{
		{
			name:               "Successfully inserts a new recipe",
			payload:            correctPayload,
			expectedStatusCode: 201,
			expectedResponse: map[string]any{
				"method": map[string]any{
					"name": "Pour Over",
				},
				"coffee": map[string]any{
					"name": "Test Coffee",
				},
				"info": map[string]any{
					"name":     "Test Name",
					"grams_in": 20,
					"ml_out":   320,
					"phases": []map[string]any{
						{
							"open":   true,
							"time":   45,
							"amount": 160,
						},
						{
							"open":   false,
							"time":   75,
							"amount": 160,
						},
						{
							"open":   true,
							"time":   60,
							"amount": 0,
						},
					},
				},
			},
			expectedError: nil,
		},
		{
			name: "Returns 404 on insert to unknown coffee",
			payload: map[string]any{
				"coffee_id": 0,
				"method_id": 1,
				"info": map[string]any{
					"name":     "Test Name",
					"grams_in": 20,
					"ml_out":   320,
					"phases": []map[string]any{
						{
							"open":   true,
							"time":   45,
							"amount": 160,
						},
						{
							"open":   false,
							"time":   75,
							"amount": 160,
						},
						{
							"open":   true,
							"time":   60,
							"amount": 0,
						},
					},
				},
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   nil,
			expectedError: map[string]any{
				"error": "The requested resource could not be found.",
			},
		},
		{
			name: "Returns 404 on no method found",
			payload: map[string]any{
				"coffee_id": 1,
				"method_id": 0,
				"info": map[string]any{
					"name":     "Test Name",
					"grams_in": 20,
					"ml_out":   320,
					"phases": []map[string]any{
						{
							"open":   true,
							"time":   45,
							"amount": 160,
						},
						{
							"open":   false,
							"time":   75,
							"amount": 160,
						},
						{
							"open":   true,
							"time":   60,
							"amount": 0,
						},
					},
				},
			},
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   nil,
			expectedError: map[string]any{
				"error": "The requested resource could not be found.",
			},
		},
		{
			name: "Returns a 422 when there is a missing field in Recipe",
			payload: map[string]any{
				"coffee_id": 1,
				"method_id": 1,
				"info": map[string]any{
					"grams_in": 20,
					"ml_out":   320,
					"phases": []map[string]any{
						{
							"open":   true,
							"time":   45,
							"amount": 160,
						},
						{
							"open":   false,
							"time":   75,
							"amount": 160,
						},
						{
							"open":   true,
							"time":   60,
							"amount": 0,
						},
					},
				},
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   nil,
			expectedError: map[string]any{
				"error": map[string]any{
					"name": "must be provided",
				},
			},
		},
		{
			name: "Returns a 422 when there is a missing field in the recipes",
			payload: map[string]any{
				"coffee_id": 1,
				"method_id": 1,
				"info": map[string]any{
					"name":     "Test Name",
					"grams_in": 20,
					"ml_out":   320,
					"phases": []map[string]any{
						{
							"time":   45,
							"amount": 160,
						},
						{
							"open":   false,
							"time":   75,
							"amount": 160,
						},
						{
							"open":   true,
							"time":   60,
							"amount": 0,
						},
					},
				},
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   nil,
			expectedError: map[string]any{
				"error": map[string]any{
					"open": "must be provided",
				},
			},
		},
		{
			name: "Returns a 422 if the grams_in, ml_out, time, or amount is negative",
			payload: map[string]any{
				"coffee_id": 1,
				"method_id": 1,
				"info": map[string]any{
					"name":     "Test Name",
					"grams_in": -1,
					"ml_out":   -1,
					"phases": []map[string]any{
						{
							"open":   true,
							"time":   -1,
							"amount": -1,
						},
						{
							"open":   false,
							"time":   75,
							"amount": 160,
						},
						{
							"open":   true,
							"time":   60,
							"amount": 0,
						},
					},
				},
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   nil,
			expectedError: map[string]any{
				"error": map[string]any{
					"amount":   "must be greater than or equal to zero",
					"grams_in": "must be greater than zero",
					"ml_out":   "must be greater than zero",
					"time":     "must be greater than zero",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestURL := fmt.Sprintf("http://localhost:%d/v1/recipes", 3001)
			payloadBytes, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("failed to marshal payload: %v", err)
			}
			requestHeaders := map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", token),
			}
			statusCode, _, body := post(t, requestURL, bytes.NewReader(payloadBytes), requestHeaders)

			assert.Equal(t, tt.expectedStatusCode, statusCode)
			if tt.expectedResponse != nil {
				recipe := body["recipe"].(map[string]any)
				assert.Equal(t, tt.expectedResponse["name"], recipe["name"])
				assert.Equal(t, tt.expectedResponse["phases"], recipe["phases"])
				assert.Equal(t, tt.expectedResponse["ml_out"], recipe["ml_out"])
				assert.Equal(t, tt.expectedResponse["grams_in"], recipe["grams_in"])
				assert.Equal(t, tt.expectedResponse["coffee"].(map[string]any)["name"], recipe["coffee"].(map[string]any)["name"])
				assert.NotEmpty(t, recipe["coffee"].(map[string]any)["id"])
				assert.Equal(t, tt.expectedResponse["method"].(map[string]any)["name"], recipe["method"].(map[string]any)["name"])
				assert.NotEmpty(t, recipe["method"].(map[string]any)["id"])
			} else {
				assert.Equal(t, tt.expectedError, body)
			}
		})
	}

	t.Run("Unauthenticated call to post recipe returns an error", func(t *testing.T) {
		requestURL := fmt.Sprintf("http://localhost:%d/v1/recipes", 3001)
		payloadBytes, err := json.Marshal(correctPayload)
		if err != nil {
			t.Fatalf("failed to marshal payload: %v", err)
		}

		statusCode, _, _ := post(t, requestURL, bytes.NewReader(payloadBytes), nil)
		assert.Equal(t, http.StatusUnauthorized, statusCode)
	})
}

// TODO TestListRecipes
func TestListRecipes(t *testing.T) {

}
