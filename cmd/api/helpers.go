package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/form"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

type envelope map[string]any

type NullInt64 struct {
	Value   *int64
	Present bool
}

// readIDParam reads the id parameter and returns the id if its valid
func (app *application) readIDParam(r *http.Request) (int64, error) {
	param := chi.URLParam(r, "id")

	id, err := strconv.ParseInt(param, 10, 64)
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

// readMultipart reads and decodes multipart form data into dst
func (app *application) readMultipart(w http.ResponseWriter, r *http.Request, dst any) error {
	// Limit the request body size
	maxBytes := 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Parse multipart form
	err := r.ParseMultipartForm(int64(maxBytes))
	if err != nil {
		if errors.Is(err, http.ErrNotMultipart) {
			return errors.New("content type must be multipart/form-data")
		} else if err.Error() == "http: request body too large" {
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)
		}
		return fmt.Errorf("error parsing multipart form: %w", err)
	}

	// Dynamically extract allowed fields from the struct
	allowedFields := getAllowedFields(dst)

	// Validate fields
	for key := range r.MultipartForm.Value {
		if _, exists := allowedFields[key]; !exists {
			return fmt.Errorf("body contains unknown key %q", key)
		}
	}

	// Decode form fields into struct
	decoder := form.NewDecoder()
	if err := decoder.Decode(dst, r.MultipartForm.Value); err != nil {
		if strings.HasPrefix(err.Error(), "form: unknown field ") {
			fieldName := strings.TrimPrefix(err.Error(), "form: unknown field ")
			return fmt.Errorf("body contains unknown key %q", fieldName)
		}
		return fmt.Errorf("failed to decode form data: %w", err)
	}

	return nil
}

// getAllowedFields extracts struct field names based on `form` tags.
func getAllowedFields(dst any) map[string]struct{} {
	allowedFields := make(map[string]struct{})
	v := reflect.TypeOf(dst)

	// Ensure we handle a pointer to struct
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Iterate through struct fields
	for i := 0; i < v.NumField(); i++ {
		tag := v.Field(i).Tag.Get("form")
		if tag != "" {
			allowedFields[tag] = struct{}{}
		}
	}
	return allowedFields
}

// background spins up a job to in a new goroutine with proper panic recovery
func (app *application) background(fn func()) {
	go func() {
		defer app.wg.Add(1)
		defer func() {
			if err := recover(); err != nil {
				app.logger.Error(fmt.Sprintf("%v", err))
			}
		}()
		fn()
	}()
}

// GetOptionalString extracts a pointer to a string from a multipart form request.
// If the field is not present, it returns nil.
func GetOptionalString(r *http.Request, field string) *string {
	if r.MultipartForm == nil {
		return nil
	}
	if values, exists := r.MultipartForm.Value[field]; exists && len(values) > 0 {
		value := values[0]
		return &value
	}
	return nil
}

// sanitizeMultipartBody reads the body and removes huge image payloads from the multipart body
func sanitizeMultipartBody(body []byte, contentType string) string {
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return sanitizeNonMultipartBody(body)
	}

	ct, params, err := mime.ParseMediaType(contentType)
	if err != nil || ct != "multipart/form-data" {
		return "[could not parse multipart]"
	}

	boundary, ok := params["boundary"]
	if !ok {
		return "[multipart with no boundary]"
	}

	mr := multipart.NewReader(bytes.NewReader(body), boundary)
	var out bytes.Buffer

	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}

		name := part.FormName()
		filename := part.FileName()

		if filename != "" {
			// File upload → OMIT the file bytes entirely
			out.WriteString(fmt.Sprintf(
				"--%s\n[field=%s filename=%s omitted]\n",
				boundary, name, filename,
			))
		} else {
			// Normal form field → read & log safely
			data, _ := io.ReadAll(part)
			cleaned := strings.TrimSpace(string(data))
			out.WriteString(fmt.Sprintf(
				"--%s\n%s=%q\n",
				boundary, name, cleaned,
			))
		}
	}

	return out.String()
}

// sanitizeNonMultipartBody just returns the stringified body
func sanitizeNonMultipartBody(body []byte) string {
	return string(body)
}

func (n *NullInt64) UnmarshalJSON(data []byte) error {
	n.Present = true

	if string(data) == "null" {
		n.Value = nil
		return nil
	}

	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	n.Value = &v
	return nil
}
