package main

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/v1/healthcheck")

	var respBody map[string]any
	err := json.Unmarshal([]byte(body), &respBody)
	if err != nil {
		t.Errorf("client: could not marshal response: %s\n", err)
		os.Exit(1)
	}

	assert.Equal(t, http.StatusOK, code)

	systemInfo := respBody["system_info"].(map[string]any)
	assert.Equal(t, systemInfo["environment"], "test")
	assert.NotEmpty(t, systemInfo["version"])
}
