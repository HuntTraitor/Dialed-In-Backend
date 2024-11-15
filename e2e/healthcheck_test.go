package e2e

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const (
	port = "3001"
)

func TestHealthcheck(t *testing.T) {
	cleanup, _, err := LaunchTestProgram(port)
	if err != nil {
		t.Fatalf("failed to launch test program: %v", err)
	}
	t.Cleanup(cleanup)

	requestURL := fmt.Sprintf("http://localhost:%d/v1/healthcheck", 3001)
	req, err := http.NewRequest(http.MethodGet, requestURL, nil)
	if err != nil {
		t.Errorf("client: could not create request: %s\n", err)
		os.Exit(1)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf("client: error making http request: %s\n", err)
		os.Exit(1)
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("client: could not read response body: %s\n", err)
		os.Exit(1)
	}

	var test map[string]any
	err = json.Unmarshal(resBody, &test)
	if err != nil {
		t.Errorf("client: could not marshal response: %s\n", err)
		os.Exit(1)
	}

	assert.Equal(t, test["status"], "available")

	systemInfo := test["system_info"].(map[string]any)
	assert.Equal(t, systemInfo["environment"], "test")
	assert.NotEmpty(t, systemInfo["version"])
}
