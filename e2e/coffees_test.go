package e2e

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
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
