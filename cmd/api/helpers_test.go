package main

import (
	"context"
	"github.com/hunttraitor/dialed-in-backend/internal/assert"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"net/http/httptest"
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
