package main

import (
	"bytes"
	"github.com/hunttraitor/dialed-in-backend/internal/mocks"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testServer struct {
	*httptest.Server
}

func newTestApplication() *application {
	var cfg config
	cfg.env = "test"

	return &application{
		config: cfg,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		models: mocks.NewMockModels(),
		mailer: mocks.NewMockMailer(),
	}
}

func newTestServer(h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)
	return &testServer{ts}
}

func (ts *testServer) get(t *testing.T, urlPath string, headers map[string]string) (int, http.Header, string) {
	t.Helper()

	// Create a new GET request
	req, err := http.NewRequest(http.MethodGet, ts.URL+urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Add custom headers to the request
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Send the request
	rs, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()

	// Read and return the response
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)
	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) post(t *testing.T, urlPath string, body io.Reader) (int, http.Header, string) {
	t.Helper()
	rs, err := ts.Client().Post(ts.URL+urlPath, "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()

	returnedBody, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	returnedBody = bytes.TrimSpace(returnedBody)
	return rs.StatusCode, rs.Header, string(returnedBody)
}

func (ts *testServer) put(t *testing.T, urlPath string, body io.Reader) (int, http.Header, string) {
	t.Helper()

	// Create a PUT request
	req, err := http.NewRequest(http.MethodPut, ts.URL+urlPath, body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	rs, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()

	// Read the response body
	returnedBody, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	returnedBody = bytes.TrimSpace(returnedBody)

	// Return the response details
	return rs.StatusCode, rs.Header, string(returnedBody)
}
