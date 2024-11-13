package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hunttraitor/dialed-in-backend/internal/assert"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadIDParam(t *testing.T) {
	app := new(application)

	tests := []struct {
		name        string
		id          string
		expectedID  int64
		expectedErr string
	}{
		{
			name:        "Valid ID",
			id:          "1",
			expectedID:  1,
			expectedErr: "",
		},
		{
			name:        "0 ID",
			id:          "0",
			expectedID:  0,
			expectedErr: "invalid id parameter",
		},
		{
			name:        "-1 ID",
			id:          "-1",
			expectedID:  0,
			expectedErr: "invalid id parameter",
		},
		{
			name:        "String ID",
			id:          "string",
			expectedID:  0,
			expectedErr: "invalid id parameter",
		},
		{
			name:        "Empty ID",
			id:          "",
			expectedID:  0,
			expectedErr: "invalid id parameter",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create http parameters with key and value as the ids
			params := httprouter.Params{
				httprouter.Param{Key: "id", Value: test.id},
			}
			// Create a context that encapsulates these parameters
			ctx := context.WithValue(context.Background(), httprouter.ParamsKey, params)

			// Create a request using that context
			req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

			id, err := app.readIDParam(req)

			if err != nil {
				assert.Equal(t, err.Error(), test.expectedErr)
			} else {
				assert.Equal(t, id, test.expectedID)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {

	mockData := struct {
		Message string `json:"message"`
	}{
		Message: "Test Message",
	}

	mockHeader := http.Header{
		"Content-Header": []string{"Custom-Value"},
	}

	app := new(application)

	tests := []struct {
		name        string
		status      int
		data        any
		headers     http.Header
		expectedErr error
	}{
		{
			name:        "Successfully writes application/json headers",
			status:      http.StatusOK,
			data:        mockData,
			headers:     http.Header{},
			expectedErr: nil,
		},
		{
			name:        "Successfully writes custom headers",
			status:      http.StatusOK,
			data:        mockData,
			headers:     mockHeader,
			expectedErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			err := app.writeJSON(rr, test.status, test.data, test.headers)

			if err != nil {
				assert.Equal(t, err.Error(), test.expectedErr.Error())
			} else {
				assert.Equal(t, rr.Code, test.status)
				assert.Equal(t, rr.Header().Get("Content-Type"), "application/json")

				// Ensure custom header is there
				for key, values := range test.headers {
					for _, value := range values {
						assert.Equal(t, rr.Header().Get(key), value)
					}
				}
				assert.StringContains(t, rr.Body.String(), mockData.Message)
			}
		})
	}
}

func TestReadJSON(t *testing.T) {
	app := new(application)

	type mockDestination struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	mockName := "John Doe"
	mockAge := 30

	tests := []struct {
		name        string
		jsonBody    string
		expectedErr string
	}{
		{
			name:        "Successfully decodes json",
			jsonBody:    fmt.Sprintf(`{"name": "%s", "age": %d}`, mockName, mockAge),
			expectedErr: "",
		},
		{
			name:        "Badly Formed json at character",
			jsonBody:    `{"name": John Doe, "age": 30}`,
			expectedErr: "body contains badly-formed JSON (at character 10)",
		},
		{
			name:        "Badly formed json",
			jsonBody:    `{ "name": "John Doe"`,
			expectedErr: "body contains badly-formed JSON",
		},
		{
			name:        "Body contains incorrect JSON type for field",
			jsonBody:    `{ "name": "John Doe", "age": "thirty" }`,
			expectedErr: `body contains incorrect JSON type for field "age"`,
		},
		{
			name:        "Body contains incorrect JSON type for character",
			jsonBody:    `["unexpected_array"]`,
			expectedErr: `body contains incorrect JSON type (at character 1)`,
		},
		{
			name:        "Body must not be empty",
			jsonBody:    "",
			expectedErr: "body must not be empty",
		},
		{
			name:        "Unknown Field",
			jsonBody:    `{"name": "John", "unknown_field": "unexpected"}`,
			expectedErr: `body contains unknown key "unknown_field"`,
		},
		{
			name:        "Data too large",
			jsonBody:    `{"name": "` + strings.Repeat("A", 1_048_577) + `"}`,
			expectedErr: `body must not be larger than 1048576 bytes`,
		},
		{
			name:        "Multiple json values",
			jsonBody:    `{"name": "John Doe", "age": 30}{"name": "John Doe", "age": 30}`,
			expectedErr: "body must only contain a single JSON value",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(test.jsonBody))
			w := httptest.NewRecorder()

			var dst mockDestination
			err := app.readJSON(w, r, &dst)

			if err != nil {
				assert.Equal(t, test.expectedErr, err.Error())
			} else {
				assert.Equal(t, dst.Name, "John Doe")
				assert.Equal(t, dst.Age, 30)
			}
		})
	}
}
