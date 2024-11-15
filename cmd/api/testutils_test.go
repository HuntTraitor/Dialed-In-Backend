package main

import (
	"bytes"
	"database/sql"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/joho/godotenv"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type testServer struct {
	*httptest.Server
}

func newTestApplication(t *testing.T) *application {

	testDB := newTestDB(t)
	testModels := data.NewModels(testDB)

	var cfg config
	cfg.env = "test"

	return &application{
		config: cfg,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		models: testModels,
	}
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
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

func newTestDB(t *testing.T) *sql.DB {
	err := godotenv.Load("../../.env")
	if err != nil {
		t.Fatal(err)
	}
	testDatabaseURL := os.Getenv("TEST_DATABASE_URL")
	db, err := sql.Open("postgres", testDatabaseURL)
	if err != nil {
		t.Fatal(err)
	}

	script, err := os.ReadFile("../../db/sql/test_setup.sql")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(string(script))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		defer db.Close()

		script, err := os.ReadFile("../../db/sql/test_teardown.sql")
		if err != nil {
			t.Fatal(err)
		}

		_, err = db.Exec(string(script))
		if err != nil {
			t.Fatal(err)
		}

	})
	return db
}
