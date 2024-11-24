package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"
)

// post takes an url and body and returns a status code, headers, and a json body
func post(t *testing.T, url string, body io.Reader) (int, http.Header, map[string]any) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	var responseBody map[string]any
	err = json.Unmarshal(resBody, &responseBody)
	if err != nil {
		t.Fatal(err)
	}
	return res.StatusCode, res.Header, responseBody
}

// put takes a URL and body and returns a status code, headers, and a JSON body
func put(t *testing.T, url string, body io.Reader) (int, http.Header, map[string]any) {
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		t.Fatal(err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close() // Ensure the response body is closed

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}

	var responseBody map[string]any
	err = json.Unmarshal(resBody, &responseBody)
	if err != nil {
		t.Fatal(err)
	}
	return res.StatusCode, res.Header, responseBody
}

// getEmail returns the string version and amount of emails with a kind and query linked to
// and returns the body and number of emails
// https://github.com/mailhog/MailHog/blob/master/docs/APIv2/swagger-2.0.yaml
func getEmail(t *testing.T, kind string, query string) (string, int) {
	t.Helper()
	requestURL := fmt.Sprintf("http://localhost:8025/api/v2/search?kind=%s&query=%s", kind, query)
	resp, err := http.Get(requestURL)
	if err != nil {
		t.Fatalf("failed to fetch messages from MailHog: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var data map[string]any
	err = json.Unmarshal(body, &data)
	if err != nil {
		t.Fatalf("failed to unmarshal JSON: %v\n", err)
	}

	count := int(data["count"].(float64))

	return string(body), count
}

// waitFor waits 5 seconds for a condition to be true
// Use this for asynchronous background tasks
func waitFor(t *testing.T, condition func() bool) {
	t.Helper()
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for values to equal each other")
		case <-ticker.C:
			if condition() {
				return
			}
		}
	}
}

// extractToken takes an email body from mailhog and returns the token inside of that body
func extractToken(t *testing.T, emailContent string) string {
	t.Helper()
	re := regexp.MustCompile(`\\?"token\\?":\\?\s*\\?"([A-Z0-9]+)\\?"`)
	match := re.FindStringSubmatch(emailContent)

	if len(match) > 1 {
		// The token is in the first capturing group
		return match[1]
	} else {
		t.Fatal("failed to extract token")
	}
	return ""
}

// createUser calls a post request to /users and returns the user
func createUser(t *testing.T) map[string]any {
	t.Helper()
	payload := `{
				"name":     "Test User",
				"email":    "test@example.com",
				"password": "password"
			}`
	requestURL := fmt.Sprintf("http://localhost:%d/v1/users", 3001)
	_, _, body := post(t, requestURL, strings.NewReader(payload))
	return body
}
