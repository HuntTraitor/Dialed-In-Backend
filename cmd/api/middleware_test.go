package main

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/hunttraitor/dialed-in-backend/internal/mocks"
	"github.com/stretchr/testify/assert"
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

func TestLogRequest(t *testing.T) {
	// Create a new logger that outputs its output into a buffer
	buf := new(bytes.Buffer)
	logHandler := slog.NewTextHandler(buf, nil)
	logger := slog.New(logHandler)
	var cfg config
	app := &application{
		config: cfg,
		logger: logger,
		models: mocks.NewMockModels(),
		mailer: mocks.NewMockMailer(),
	}

	ts := newTestServer(nil)
	defer ts.Close()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		_, err := w.Write([]byte("Test Response Body"))
		if err != nil {
			t.Fatal(err)
		}
	})

	loggerMiddleware := app.logRequest(handler)

	t.Run("Successfully logs the request and response", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		loggerMiddleware.ServeHTTP(rr, req)
		expectedRequestLog := `method=GET uri=/ query_params=map[] body=""`
		expectedResponseLog := `status=200 body="Test Response Body"`

		// Read from the buffer and assert the logs are correct
		assert.Contains(t, buf.String(), expectedRequestLog)
		assert.Contains(t, buf.String(), expectedResponseLog)

	})
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
			expectedErr:        "invalid or missing authentication token",
		},
		{
			name:               "invalid auth header returns error",
			token:              map[string]string{"Authorization": "Bearer ASDJKLEPOIURERFJDKSLAIEJG2"},
			expectedStatusCode: http.StatusUnauthorized,
			expectedUser:       nil,
			expectedErr:        "invalid or missing authentication token",
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

	t.Run("Exposes Prometheus metrics endpoint", func(t *testing.T) {
		// Trigger at least one request through API middleware.
		req := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)

		// Confirm /metrics is exposed and includes our metric families.
		req = httptest.NewRequest(http.MethodGet, "/metrics", nil)
		rr = httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "http_requests_total")
		assert.Contains(t, rr.Body.String(), "http_responses_total")
		assert.Contains(t, rr.Body.String(), "http_request_duration_seconds")
	})
}
