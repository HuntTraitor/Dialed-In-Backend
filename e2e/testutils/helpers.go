package testutils

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/require"
)

func uniqueEmail() string {
	return fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())
}

func DecodeJSON(t *testing.T, res *httpexpect.Response, v any) {
	t.Helper()

	raw := res.JSON().Raw()

	b, err := json.Marshal(raw)
	require.NoError(t, err)

	err = json.Unmarshal(b, v)
	require.NoError(t, err)
}

// WaitFor waits 5 seconds for a condition to be true
// Use this for asynchronous background tasks
func WaitFor(t *testing.T, condition func() bool) {
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

// Ptr takes a type and returns to pointer to that type
func Ptr[T any](v T) *T {
	return &v
}
