package main

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRecoverPanic(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	expectedValue := "close"
	expectedError := map[string]any{
		"error": "An internal server error has occurred. Please try again later.",
	}

	rr := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	app.recoverPanic(panicHandler).ServeHTTP(rr, r)

	assert.Equal(t, rr.Code, http.StatusInternalServerError)
	assert.Equal(t, expectedValue, rr.Header().Get("Connection"))
	assert.Contains(t, rr.Body.String(), expectedError["error"])
}

func TestRateLimit(t *testing.T) {
	// Initializing expectations
	expiration := time.Second
	rateLimitExceededMessage := map[string]any{
		"error": "rate limit exceeded",
	}

	// Creating a new app with a configured limiter
	app := newTestApplication(t)
	app.config.limiter.enabled = true
	app.config.limiter.rps = 1
	app.config.limiter.burst = 1
	app.config.limiter.expiration = expiration
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	// Creating a handler on the app
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrapping that handler in rateLimit middleware
	rateLimitMiddleware := app.rateLimit(handler)

	// Check that it allows the first request
	t.Run("Allow first request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		rateLimitMiddleware.ServeHTTP(rr, req)
		assert.Equal(t, rr.Code, http.StatusOK)
	})

	// Check that it blocks the second request
	t.Run("Rate limit exceeded", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		rateLimitMiddleware.ServeHTTP(rr, req)
		assert.Equal(t, rr.Code, http.StatusTooManyRequests)
		assert.Contains(t, rr.Body.String(), rateLimitExceededMessage["error"])
	})

	// Send another request and wait until after the expiration and send another request
	t.Run("Client expiry after timeout", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		// Send first request
		rr := httptest.NewRecorder()
		rateLimitMiddleware.ServeHTTP(rr, req)
		assert.Equal(t, rr.Code, http.StatusTooManyRequests)

		// Sleep for a time over the expiration
		time.Sleep(expiration + time.Millisecond)

		// Send another request
		rr2 := httptest.NewRecorder()
		rateLimitMiddleware.ServeHTTP(rr2, req)
		assert.Equal(t, http.StatusOK, rr2.Code)
	})
}
