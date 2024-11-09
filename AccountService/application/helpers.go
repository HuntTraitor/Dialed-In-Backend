package application

import (
	"encoding/json"
	"errors"
	customError "github.com/hunttraitor/dialed-in-backend/errors"
	"log/slog"
	"net/http"
)

type ApiFunc func(w http.ResponseWriter, r *http.Request) error

func Make(h ApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			var apiErr customError.ApiError
			if errors.As(err, &apiErr) {
				writeJson(w, apiErr.StatusCode, apiErr)
			} else {
				errResp := map[string]any{
					"statusCode": http.StatusInternalServerError,
					"msg":        "internal server errors",
				}
				writeJson(w, http.StatusInternalServerError, errResp)
			}
			slog.Error("HTTP API Error", "err", err.Error(), "path", r.URL.Path)
		}
	}
}

func writeJson(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
