package main

import (
	"fmt"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/v1/healthcheck")
	fmt.Println(code)
	fmt.Println(body)
}
