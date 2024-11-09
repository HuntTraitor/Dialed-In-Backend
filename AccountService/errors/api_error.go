package errors

import (
	"fmt"
	"net/http"
)

type ApiError struct {
	StatusCode int `json:"status_code"`
	Message    any `json:"msg"`
}

func (a ApiError) Error() string {
	return fmt.Sprintf("api errors: %d", a.StatusCode)
}

func NewApiError(statusCode int, err error) ApiError {
	return ApiError{
		StatusCode: statusCode,
		Message:    err.Error(),
	}
}

func InvalidRequestData(errors map[string]string) ApiError {
	return ApiError{
		StatusCode: http.StatusUnprocessableEntity,
		Message:    errors,
	}
}

func InvalidJSON() ApiError {
	return NewApiError(http.StatusBadRequest, fmt.Errorf("invalid JSON request data"))
}

func DuplicateEmail() ApiError {
	return NewApiError(http.StatusConflict, fmt.Errorf("there is already an account for this email address"))
}
