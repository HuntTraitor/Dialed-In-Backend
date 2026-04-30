package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

type wrappedResponseWriter struct {
	wrapped       http.ResponseWriter
	statusCode    int
	headerWritten bool
	body          *bytes.Buffer
}

// newWrappedResponseWriter creates a new response writer that is just a wrapper for a regular responseWriter
// that records the status codes
func newWrappedResponseWriter(w http.ResponseWriter) *wrappedResponseWriter {
	return &wrappedResponseWriter{
		wrapped:    w,
		statusCode: http.StatusOK,
		body:       &bytes.Buffer{},
	}
}

// Header to implement responseWriter interface
func (ww *wrappedResponseWriter) Header() http.Header {
	return ww.wrapped.Header()
}

// WriteHeader to implement responseWriter interface
func (ww *wrappedResponseWriter) WriteHeader(statusCode int) {
	ww.wrapped.WriteHeader(statusCode)

	if !ww.headerWritten {
		ww.statusCode = statusCode
		ww.headerWritten = true
	}
}

// Write to implement responseWriter interface
func (ww *wrappedResponseWriter) Write(b []byte) (int, error) {
	ww.headerWritten = true
	ww.body.Write(b)
	return ww.wrapped.Write(b)
}

// Unwrap returns the original wrapped requestWriter
func (ww *wrappedResponseWriter) Unwrap() http.ResponseWriter {
	return ww.wrapped
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// logRequest logs the incoming request in the logger
func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			next.ServeHTTP(w, r)
			return
		}

		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			app.logger.Error("failed to read request body",
				"error", err,
				"method", r.Method,
				"uri", r.URL.RequestURI(),
			)

			http.Error(w, "failed to read request body", http.StatusBadRequest)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(rawBody))

		app.logger.Debug("received request",
			"ip", r.RemoteAddr,
			"proto", r.Proto,
			"method", r.Method,
			"uri", r.URL.RequestURI(),
			"query_params", r.URL.Query(),
			"body", bodyForLog(rawBody, r.Header.Get("Content-Type")),
		)

		ww := newWrappedResponseWriter(w)

		next.ServeHTTP(ww, r)

		status := ww.statusCode
		if status == 0 {
			status = http.StatusOK
		}

		attrs := []any{
			"ip", r.RemoteAddr,
			"proto", r.Proto,
			"method", r.Method,
			"path", r.URL.Path,
			"uri", r.URL.RequestURI(),
			"query_params", r.URL.Query(),
			"status", status,
			"response_body", bodyForLog(ww.body.Bytes(), ww.Header().Get("Content-Type")),
		}

		switch {
		case status >= 500:
			app.logger.Error("request failed", attrs...)

		case status >= 400:
			app.logger.Warn("request warning", attrs...)

		default:
			app.logger.Info("request completed", attrs...)
		}
	})
}

// rateLimit limits the rate of requests per client in a map of ips to rates
func (app *application) rateLimit(next http.Handler) http.Handler {

	// Each client has their own limiter and lastSeen time
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a new goroutine that deletes client ips in the background after a certain amount of time
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.Lock()
			for ip, client := range clients {
				if time.Since(client.lastSeen) > app.config.limiter.expiration {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Check the rate limiting configuration in main.go
		if app.config.limiter.enabled {
			// Retrieve ip from host
			ip := realip.FromRequest(r)

			// Assign a new rate limiter to a client if they were not found
			mu.Lock()
			if _, found := clients[ip]; !found {
				clients[ip] = &client{
					limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst),
				}
			}

			// Set there last seen time to now
			clients[ip].lastSeen = time.Now()

			// If the client has exceeded rate limits, send an error and unlock the mutex
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}
			// Unlock mutex and continue
			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

// authenticate checks if a user token was placed in the header
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authorizationHeader := r.Header.Get("Authorization")

		// if the authorization header is not set, set the user context to an anonymous blank user
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// if authorization header is poorly formatted, send error
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		v := validator.New()

		// if the authorization token is not valid
		if data.ValidateTokenPlainText(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Get the user associated with the token
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

// requireAuthenticatedUser checks that the user is not anonymous
func (app *application) requireAuthenticatedUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// requireActivatedUser checks that user.Activated is true
func (app *application) requireActivatedUser(next http.Handler) http.Handler {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	return app.requireAuthenticatedUser(fn)
}

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "The total number of HTTP requests received.",
		},
		[]string{"method", "route"},
	)

	httpResponsesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_responses_total",
			Help: "The total number of HTTP responses sent.",
		},
		[]string{"method", "route", "status"},
	)

	httpRequestDurationSeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "The duration of HTTP requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route", "status"},
	)
)

// metrics records request counts, response status counts, and request durations.
func (app *application) metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := newWrappedResponseWriter(w)

		next.ServeHTTP(ww, r)

		route := "unknown"
		if rctx := chi.RouteContext(r.Context()); rctx != nil {
			if pattern := rctx.RoutePattern(); pattern != "" {
				route = pattern
			}
		}

		status := strconv.Itoa(ww.statusCode)

		httpRequestsTotal.WithLabelValues(
			r.Method,
			route,
		).Inc()

		httpResponsesTotal.WithLabelValues(
			r.Method,
			route,
			status,
		).Inc()

		httpRequestDurationSeconds.WithLabelValues(
			r.Method,
			route,
			status,
		).Observe(time.Since(start).Seconds())
	})
}
