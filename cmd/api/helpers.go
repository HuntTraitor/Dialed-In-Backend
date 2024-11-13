package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"github.com/julienschmidt/httprouter"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type envelope map[string]any

// readIDParam reads the id parameter and returns the id if its valid
func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

// writeJSON writes data to the response so the client can see
func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers http.Header) error {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
	return nil
}

// readJSON reads and decodes data from the body and writes it to a dst
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	// Ensure max limit to how much json a user can send
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Force an error when an unknown field is passed in by the body
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		// Initialize the errors
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalFieldError *json.InvalidUnmarshalError
		var maxBytesError *http.MaxBytesError

		switch {
		// Example: {"name": John Doe, "age": 30}
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

			// Example: { "name": "John Doe"
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

			// Example: { "name": "John Doe", "age": "thirty" }
		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

			// Example: ""
		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

			// Example: {"name": "John", "unknown_field": "unexpected"}  (`unknown_field` is not in the struct)
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

			// Example: {"large_data": "..."}
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("body must not be larger than %d bytes", maxBytesError.Limit)

		case errors.As(err, &invalidUnmarshalFieldError):
			panic(err)

		default:
			return err
		}
	}

	// Check for situation where users inputted 2 json field
	// i.e. {"foo":"bar"}{"biz":"baz"}
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// readString returns a string from a query parameter, empty if it could not be found
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	return s
}

// readCSV returns an array of strings based on the query seperated by commas. i.e. ?value=1,2,3 returns [1,2,3]
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}
	return strings.Split(csv, ",")
}

// readInt returns an integer from a query parameter
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer")
		return defaultValue
	}
	return i
}
