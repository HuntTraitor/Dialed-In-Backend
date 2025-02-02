package e2e

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"strings"
	"testing"
)

func TestGetAllCoffees(t *testing.T) {
	cleanup, _, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatalf("failed to launch test program: %v", err)
	}
	t.Cleanup(cleanup)

	t.Run("Successfully gets list of coffees", func(t *testing.T) {

		// Log in user
		authenticateBody := authenticateUser(t, "hunter@gmail.com", "password")
		token := authenticateBody["authentication_token"].(map[string]any)["token"].(string)

		// Send request to coffee endpoint and extract body
		requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)
		requestHeaders := map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", token),
		}
		statusCode, _, respBody := get(t, requestURL, requestHeaders)
		var body map[string]any
		err = json.Unmarshal([]byte(respBody), &body)
		if err != nil {
			t.Fatal(err)
		}

		// assert status codes are equal
		assert.Equal(t, http.StatusOK, statusCode)

		// assert body is correct
		coffees := body["coffees"].([]any)
		for _, coffee := range coffees {
			c := coffee.(map[string]any)
			assert.NotEmpty(t, c["id"].(float64))
			assert.NotEmpty(t, c["user_id"].(float64))
			assert.NotEmpty(t, c["created_at"].(string))
			assert.Equal(t, "Milky Cake", c["name"].(string))
			assert.Equal(t, "Columbia", c["region"].(string))
			assert.Equal(t, float64(1), c["version"].(float64))
			assert.NotEmpty(t, c["img"].(string))
			assert.NotEmpty(t, c["description"].(string))
		}
	})

	t.Run("Fails to get coffees when not logged in", func(t *testing.T) {

		// Send request with no auth header
		requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)
		statusCode, _, respBody := get(t, requestURL, nil)
		var body map[string]any
		err = json.Unmarshal([]byte(respBody), &body)
		if err != nil {
			t.Fatal(err)
		}

		// Assert there is a failed response
		assert.Equal(t, http.StatusUnauthorized, statusCode)
		assert.Equal(t, "you must be authenticated to access this resource", body["error"].(string))
	})
}

func TestPostCoffee(t *testing.T) {
	cleanup, _, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatalf("failed to launch test program: %v", err)
	}
	t.Cleanup(cleanup)

	resp := authenticateUser(t, "hunter@gmail.com", "password")
	token := resp["authentication_token"].(map[string]any)["token"].(string)

	longName := strings.Repeat("A", 510)
	longRegion := strings.Repeat("B", 110)
	longDescription := strings.Repeat("C", 1010)
	longImage := strings.Repeat("D", 9000)

	tests := []struct {
		name               string
		payload            string
		expectedStatusCode int
		expectedResponse   map[string]any
		expectedError      map[string]any
	}{
		{
			name: "Successfully posts a new coffee",
			payload: `{
				"name": "Blueberry Boom",
				"region": "Ethiopia",
				"img": "https://st.kofio.co/img_product/boeV9yxzHn2OwWv/9626/sq_350_DisfG6edTXbtaYponjRQ_102573.png",
				"description": "This is a delicious blueberry coffee :)"
			}`,
			expectedStatusCode: http.StatusCreated,
			expectedResponse: map[string]any{
				"id":          1,
				"user_id":     1,
				"name":        "Blueberry Boom",
				"region":      "Ethiopia",
				"img":         "https://st.kofio.co/img_product/boeV9yxzHn2OwWv/9626/sq_350_DisfG6edTXbtaYponjRQ_102573.png",
				"description": "This is a delicious blueberry coffee :)",
				"created_at":  "2025-02-01T03:59:07Z",
				"version":     1,
			},
			expectedError: nil,
		},
		{
			name:               "Missing parameters returns errors",
			payload:            `{}`,
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"description": "must be provided",
				"name":        "must be provided",
				"region":      "must be provided",
			},
		},
		{
			name: "Coffee name too long returns an error",
			payload: fmt.Sprintf(`{
				"name": "%s",
				"region": "Ethiopia",
				"img": "https://st.kofio.co/img_product/boeV9yxzHn2OwWv/9626/sq_350_DisfG6edTXbtaYponjRQ_102573.png",
				"description": "This is a delicious blueberry coffee :)"
			}`, longName),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"name": "must not be more than 500 bytes long",
			},
		},
		{
			name: "Coffee region too long returns an error",
			payload: fmt.Sprintf(`{
				"name": "Blueberry Boom",
				"region": "%s",
				"img": "https://st.kofio.co/img_product/boeV9yxzHn2OwWv/9626/sq_350_DisfG6edTXbtaYponjRQ_102573.png",
				"description": "This is a delicious blueberry coffee :)"
			}`, longRegion),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"region": "must not be more than 100 bytes long",
			},
		},
		{
			name: "Coffee description too long returns an error",
			payload: fmt.Sprintf(`{
				"name": "Blueberry Boom",
				"region": "Ethiopia",
				"img": "https://st.kofio.co/img_product/boeV9yxzHn2OwWv/9626/sq_350_DisfG6edTXbtaYponjRQ_102573.png",
				"description": "%s"
			}`, longDescription),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"description": "must not be more than 1000 bytes long",
			},
		},
		{
			name: "Coffee image url too long returns an error",
			payload: fmt.Sprintf(`{
				"name": "Blueberry Boom",
				"region": "Ethiopia",
				"img": "%s",
				"description": "This is a delicious blueberry coffee :)"
			}`, longImage),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"img": "must not be more than 8192 bytes long",
			},
		},
		{
			name: "Coffee image url is not a valid url",
			payload: `{
				"name": "Blueberry Boom",
				"region": "Ethiopia",
				"img": "htts://st.kofio.co/img_product/boeV9yxzHn2OwWv/9626/sq_350_DisfG6edTXbtaYponjRQ_102573.png",
				"description": "This is a delicious blueberry coffee :)"
			}`,
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedError: map[string]any{
				"img": "must be a valid image URL",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", token),
			}
			requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)
			statusCode, _, returnedBody := post(t, requestURL, strings.NewReader(tt.payload), headers)
			assert.Equal(t, tt.expectedStatusCode, statusCode)

			if tt.expectedError == nil {
				returnedCoffee := returnedBody["coffee"].(map[string]any)
				assert.Equal(t, tt.expectedResponse["name"], returnedCoffee["name"])
				assert.Equal(t, tt.expectedResponse["region"], returnedCoffee["region"])
				assert.Equal(t, tt.expectedResponse["img"], returnedCoffee["img"])
				assert.Equal(t, tt.expectedResponse["description"], returnedCoffee["description"])
				assert.NotEmpty(t, returnedCoffee["version"])
				assert.NotEmpty(t, returnedCoffee["id"])
				assert.NotEmpty(t, returnedCoffee["user_id"])
				assert.NotEmpty(t, returnedCoffee["created_at"])
			} else {
				returnedError := returnedBody["error"].(map[string]any)
				assert.Equal(t, tt.expectedError, returnedError)
			}

		})
	}

	t.Run("Unauthenticated call response when an error", func(t *testing.T) {
		requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)
		payload := `{
				"name": "Blueberry Boom",
				"region": "Ethiopia",
				"img": "https://st.kofio.co/img_product/boeV9yxzHn2OwWv/9626/sq_350_DisfG6edTXbtaYponjRQ_102573.png",
				"description": "This is a delicious blueberry coffee :)"
			}`
		statusCode, _, returnedBody := post(t, requestURL, strings.NewReader(payload), nil)
		assert.Equal(t, http.StatusUnauthorized, statusCode)
		assert.Equal(t, "you must be authenticated to access this resource", returnedBody["error"])
	})
}
