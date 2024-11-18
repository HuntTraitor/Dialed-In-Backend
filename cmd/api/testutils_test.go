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

func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)
	return rs.StatusCode, rs.Header, string(body)
}

func (ts *testServer) post(t *testing.T, urlPath string, body io.Reader) (int, http.Header, string) {
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
