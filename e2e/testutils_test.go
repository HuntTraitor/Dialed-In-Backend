package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
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

// getEmail returns the string version and amount of emails with a kind and query linked to
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
