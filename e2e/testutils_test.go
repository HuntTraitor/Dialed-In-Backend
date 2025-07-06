package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	port = "3001"
)

// post takes an url and body and returns a status code, headers, and a json body
func post(t *testing.T, url string, body io.Reader, headers map[string]string) (int, http.Header, map[string]any) {
	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		t.Fatal(err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
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

// patch takes a URL and body and returns a status code, headers, and a JSON body
func patch(t *testing.T, url string, body io.Reader, headers map[string]string) (int, http.Header, map[string]any) {
	req, err := http.NewRequest(http.MethodPatch, url, body)
	if err != nil {
		t.Fatal(err)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
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

// get sends a get request to a certain urlPath with some headers
func get(t *testing.T, urlPath string, headers map[string]string) (int, http.Header, map[string]any) {
	t.Helper()

	// Create a new GET request
	req, err := http.NewRequest(http.MethodGet, urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add custom headers to the request
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Send the request
	rs, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()

	// Read and return the response
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	var responseBody map[string]any
	err = json.Unmarshal(body, &responseBody)
	if err != nil {
		t.Fatal(err)
	}
	return rs.StatusCode, rs.Header, responseBody
}

// delete sends a delete request to a certain urlPath with some headers
func delete(t *testing.T, urlPath string, headers map[string]string) (int, http.Header, map[string]any) {
	t.Helper()

	// Create a new DELETE request
	req, err := http.NewRequest(http.MethodDelete, urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add custom headers to the request
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Send the request
	rs, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()

	resBody, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	var responseBody map[string]any
	err = json.Unmarshal(resBody, &responseBody)
	if err != nil {
		t.Fatal(err)
	}
	return rs.StatusCode, rs.Header, responseBody
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

// boolPtr takes a boolean and returns to pointer to that boolean
// Mainly used for initialing data.Phase
func boolPtr(b bool) *bool {
	return &b
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
	_, _, body := post(t, requestURL, strings.NewReader(payload), nil)
	return body
}

// authenticateUser authenticates the user and returns the token
func authenticateUser(t *testing.T, email string, password string) map[string]any {
	t.Helper()
	payload := fmt.Sprintf(`{
		"email": "%s",
		"password": "%s"
	}`, email, password)
	requestURL := fmt.Sprintf("http://localhost:%d/v1/tokens/authentication", 3001)
	_, _, body := post(t, requestURL, strings.NewReader(payload), nil)
	return body
}

// activateUser calls a put request to /users/activated to activate a user
func activateUser(t *testing.T, token string) map[string]any {
	t.Helper()
	payload := fmt.Sprintf(`{
				"token": "%s"
			}`, token)

	requestURL := fmt.Sprintf("http://localhost:%d/v1/users/activated", 3001)
	_, _, body := put(t, requestURL, strings.NewReader(payload))
	return body
}

func createCoffee(t *testing.T, authToken string, coffee data.Coffee, image []byte) map[string]any {
	t.Helper()
	requestURL := fmt.Sprintf("http://localhost:%d/v1/coffees", 3001)

	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// Write form fields from coffee.Info
	info := coffee.Info

	writeField := func(key, value string) {
		if err := writer.WriteField(key, value); err != nil {
			t.Fatalf("failed to write field %s: %v", key, err)
		}
	}

	writeField("name", info.Name)
	writeField("region", info.Region)
	writeField("process", info.Process)
	writeField("description", info.Description)
	writeField("origin_type", info.OriginType)
	writeField("roast_level", info.RoastLevel)
	writeField("rating", strconv.Itoa(info.Rating))
	writeField("cost", fmt.Sprintf("%.2f", info.Cost))
	writeField("decaff", strconv.FormatBool(info.Decaff))

	// Add tasting_notes as repeated form fields
	for _, note := range info.TastingNotes {
		writeField("tasting_notes", note)
	}

	// Add image file
	fileWriter, err := writer.CreateFormFile("img", "test_image.jpg")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := fileWriter.Write(image); err != nil {
		t.Fatalf("failed to write image data: %v", err)
	}

	// Close the writer before sending
	writer.Close()

	// Headers
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", authToken),
		"Content-Type":  writer.FormDataContentType(),
	}

	// Send the request and return the parsed body
	_, _, returnedBody := post(t, requestURL, &b, headers)
	return returnedBody
}

// createRecipe sends a post request using an authToken and returns the body
func createRecipe(t *testing.T, authToken string, recipe data.Recipe) map[string]any {
	t.Helper()
	requestURL := fmt.Sprintf("http://localhost:%d/v1/recipes", 3001)
	payloadBytes, err := json.Marshal(recipe)
	if err != nil {
		t.Fatalf("failed to marshal payload: %v", err)
	}
	requestHeaders := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", authToken),
	}

	_, _, body := post(t, requestURL, bytes.NewReader(payloadBytes), requestHeaders)
	return body
}

// decodePhases decodes a phase into a comparable object for easy comprisons
func decodePhases(input []interface{}) []data.Phase {
	phases := make([]data.Phase, len(input))
	for i, item := range input {
		m := item.(map[string]interface{})
		open := m["open"].(bool)
		phaseTime := int(m["time"].(float64))
		amount := int(m["amount"].(float64))
		phases[i] = data.Phase{
			Open:   &open,
			Time:   phaseTime,
			Amount: amount,
		}
	}
	return phases
}
