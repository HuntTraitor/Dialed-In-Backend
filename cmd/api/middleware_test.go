package main

import (
	"bytes"
	"encoding/json"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/hunttraitor/dialed-in-backend/internal/mocks"
	"github.com/stretchr/testify/assert"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRecoverPanic(t *testing.T) {
	app := newTestApplication()
	ts := newTestServer(nil)
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
	app := newTestApplication()
	app.config.limiter.enabled = true
	app.config.limiter.rps = 1
	app.config.limiter.burst = 1
	app.config.limiter.expiration = expiration
	ts := newTestServer(nil)
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

func TestAuthenticate(t *testing.T) {
	app := newTestApplication()
	ts := newTestServer(nil)
	defer ts.Close()

	tests := []struct {
		name               string
		token              map[string]string
		expectedStatusCode int
		expectedUser       *data.User
		expectedErr        string
	}{
		{
			name:               "No auth header passes through with anonymous user",
			token:              map[string]string{"": ""},
			expectedStatusCode: http.StatusOK,
			expectedUser:       &data.User{},
			expectedErr:        "",
		},
		{
			name:               "Empty auth header return anonymous user",
			token:              map[string]string{"Authorization": ""},
			expectedStatusCode: http.StatusOK,
			expectedUser:       &data.User{},
			expectedErr:        "",
		},
		{
			name:               "Poorly formatted auth header returns error",
			token:              map[string]string{"Authorization": "Bearer "},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUser:       nil,
			expectedErr:        "invalid of missing authentication token",
		},
		{
			name:               "invalid auth header returns error",
			token:              map[string]string{"Authorization": "Bearer ASDJKLEPOIURERFJDKSLAIEJG2"},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUser:       nil,
			expectedErr:        "invalid of missing authentication token",
		},
		{
			name:               "Correct auth header returns user",
			token:              map[string]string{"Authorization": "Bearer ASDJKLEPOIURERFJDKSLAIEJG1"},
			expectedStatusCode: http.StatusOK,
			expectedUser: &data.User{
				ID:    1,
				Name:  "Test User",
				Email: "test@example.com",
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Handler to inspect the user context
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				user := app.contextGetUser(r)                // Get the user from the context
				assert.Equal(t, tt.expectedUser.ID, user.ID) // Validate user
				assert.Equal(t, tt.expectedUser.Name, user.Name)
				assert.Equal(t, tt.expectedUser.Email, user.Email)
				w.WriteHeader(http.StatusOK)
			})

			// Wrap the handler with the authenticate middleware
			authenticateMiddleware := app.authenticate(handler)

			// Create the request
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			for k, v := range tt.token {
				req.Header.Set(k, v)
			}

			rr := httptest.NewRecorder()

			// Serve the request
			authenticateMiddleware.ServeHTTP(rr, req)

			// Validate the response
			assert.Equal(t, tt.expectedStatusCode, rr.Code)
			if tt.expectedErr != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedErr)
			}
		})
	}
}

func TestRequireActivatedUserMiddleware(t *testing.T) {
	app := newTestApplication()
	ts := newTestServer(nil)
	defer ts.Close()

	tests := []struct {
		name        string
		user        *data.User
		expectedErr string
	}{
		{
			name: "Successfully passes through with activated user",
			user: &data.User{
				Activated: true,
			},
			expectedErr: "",
		},
		{
			name:        "Rejects Anonymous user",
			user:        data.AnonymousUser,
			expectedErr: "you must be authenticated to access this resource",
		},
		{
			name: "Rejects unactivated user",
			user: &data.User{
				Activated: false,
			},
			expectedErr: "your user account must be activated to access this feature",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			requireActivateUserMiddleware := app.requireActivatedUser(handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			// set req equal to the request with the user context set
			req = app.contextSetUser(req, tt.user)
			rr := httptest.NewRecorder()
			requireActivateUserMiddleware.ServeHTTP(rr, req)
			assert.Contains(t, rr.Body.String(), tt.expectedErr)
		})
	}
}

func TestRequireAuthenticatedUserMiddleware(t *testing.T) {
	app := newTestApplication()
	ts := newTestServer(nil)
	defer ts.Close()

	tests := []struct {
		name        string
		user        *data.User
		expectedErr string
	}{
		{
			name:        "Successfully passes through with authenticated empty user",
			user:        &data.User{},
			expectedErr: "",
		},
		{
			name:        "Rejects anonymous user",
			user:        data.AnonymousUser,
			expectedErr: "you must be authenticated to access this resource",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			requireActivateUserMiddleware := app.requireActivatedUser(handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)

			// set req equal to the request with the user context set
			req = app.contextSetUser(req, tt.user)
			rr := httptest.NewRecorder()
			requireActivateUserMiddleware.ServeHTTP(rr, req)
			assert.Contains(t, rr.Body.String(), tt.expectedErr)
		})
	}
}

func TestMetricsMiddleware(t *testing.T) {
	// Manually add a new application with metrics enabled
	var cfg config
	cfg.env = "test"
	cfg.metrics = true
	app := &application{
		config: cfg,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		models: mocks.NewMockModels(),
		mailer: mocks.NewMockMailer(),
	}

	router := app.routes()
	ts := newTestServer(router)
	defer ts.Close()

	t.Run("Successfully updates metrics on request", func(t *testing.T) {
		// Send a request to healthcheck
		req := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Send a request to /debug/vars to check output
		req = httptest.NewRequest(http.MethodGet, "/debug/vars", nil)
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		var responseBody map[string]any
		err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
		if err != nil {
			t.Fatal(err)
		}

		// assert the debug has a 200 response but not a 400 response
		assert.Equal(t, float64(2), responseBody["total_requests_received"])
		assert.Equal(t, float64(1), responseBody["total_responses_sent"])
		assert.Greater(t, responseBody["total_processing_time_microseconds"], float64(0))
		assert.Equal(t, float64(1), responseBody["total_responses_sent_by_status"].(map[string]any)["200"])
		assert.Equal(t, nil, responseBody["total_responses_sent_by_status"].(map[string]any)["400"])

		// Send a intentional 400 response request
		req = httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewReader([]byte("hi")))
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		// Check /debug/vars again
		req = httptest.NewRequest(http.MethodGet, "/debug/vars", nil)
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		err = json.Unmarshal(rr.Body.Bytes(), &responseBody)
		if err != nil {
			t.Fatal(err)
		}

		// Assert that the 400 request was not logged
		assert.Equal(t, float64(2), responseBody["total_responses_sent_by_status"].(map[string]any)["200"])
		assert.Equal(t, float64(1), responseBody["total_responses_sent_by_status"].(map[string]any)["400"])
	})
}
