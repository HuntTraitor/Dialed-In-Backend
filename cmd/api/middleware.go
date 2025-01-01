package main

import (
	"bytes"
	"errors"
	"expvar"
	"fmt"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
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
		var (
			ip     = r.RemoteAddr
			proto  = r.Proto
			method = r.Method
			uri    = r.URL.RequestURI()
		)

		var body []byte
		if r.Body != nil {
			body, _ = io.ReadAll(r.Body)
			r.Body = io.NopCloser(bytes.NewReader(body)) // Reset the body for further handlers
		}
		prettyBody := strings.ReplaceAll(string(body), " ", "")
		prettyBody = strings.ReplaceAll(prettyBody, "\n", "")

		app.logger.Info("received request",
			"ip", ip,
			"proto", proto,
			"method", method,
			"uri", uri,
			"body", prettyBody,
		)

		ww := newWrappedResponseWriter(w)

		next.ServeHTTP(ww, r)

		responseBody := strings.ReplaceAll(ww.body.String(), " ", "")
		responseBody = strings.ReplaceAll(responseBody, "\n", "")
		responseBody = strings.ReplaceAll(responseBody, "\t", "")

		app.logger.Info("received response", "status", ww.statusCode, "body", responseBody)
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

// metrics updates information about the requests, responses, and response times received by the server
func (app *application) metrics(next http.Handler) http.Handler {
	var (
		totalRequestsReceived           = expvar.NewInt("total_requests_received")
		totalResponsesSent              = expvar.NewInt("total_responses_sent")
		totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_microseconds")
		totalResponsesSentByStatus      = expvar.NewMap("total_responses_sent_by_status")
	)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		totalRequestsReceived.Add(1)

		// create new metricsResponseWriter
		ww := newWrappedResponseWriter(w)
		next.ServeHTTP(ww, r)

		totalResponsesSent.Add(1)
		totalResponsesSentByStatus.Add(strconv.Itoa(ww.statusCode), 1)
		duration := time.Since(start).Microseconds()

		totalProcessingTimeMicroseconds.Add(duration)
	})
}
