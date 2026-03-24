package testutils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func GetEmail(t *testing.T, kind, query string) (string, int) {
	t.Helper()

	requestURL := fmt.Sprintf(
		"http://localhost:8025/api/v2/search?kind=%s&query=%s",
		url.QueryEscape(kind),
		url.QueryEscape(query),
	)

	resp, err := http.Get(requestURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	var data map[string]any
	err = json.Unmarshal(raw, &data)
	require.NoError(t, err)

	c, ok := data["count"].(float64)
	if !ok {
		require.Fail(t, "failed to parse 'count' as float64")
	}

	count := int(c)

	return string(raw), count
}

// ExtractToken takes an email body from mailhog and returns the token inside of that body
func ExtractToken(emailContent string) string {
	// Look for the text right before the token, then capture 6 digits after it
	re := regexp.MustCompile(`Your one-time password reset token is:[^0-9]*([0-9]{6})`)

	match := re.FindStringSubmatch(emailContent)

	if len(match) > 1 {
		return match[1] // the 6-digit token
	}

	return ""
}

func AssertPasswordResetToken(t *testing.T, email string) string {
	t.Helper()

	var token string

	assert.Eventually(t, func() bool {
		body, _ := GetEmail(t, "to", email)
		token = ExtractToken(body)
		return token != ""
	}, 2*time.Second, 100*time.Millisecond, "token was not found")

	return token
}

func AssertNoPasswordResetToken(t *testing.T, email string) {
	t.Helper()

	assert.Never(t, func() bool {
		body, _ := GetEmail(t, "to", email)
		token := ExtractToken(body)
		return token != ""
	}, 2*time.Second, 100*time.Millisecond, "token was found")
}

func AssertEmailSent(t *testing.T, field, value string) string {
	t.Helper()
	var body string

	assert.Eventually(t, func() bool {
		var count int
		body, count = GetEmail(t, field, value)
		return count > 0
	}, 2*time.Second, 100*time.Millisecond, "email was not found")

	return body
}

func AssertNoEmailSent(t *testing.T, field, value string) {
	t.Helper()

	assert.Never(t, func() bool {
		_, count := GetEmail(t, field, value)
		return count > 0
	}, 2*time.Second, 100*time.Millisecond, "email was found")
}
