package validator

import (
	"regexp"
	"slices"
)

// EmailRX https://html.spec.whatwg.org/#valid-e-mail-address
var (
	EmailRX = "^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$"
)

type Validator struct {
	Errors map[string]string
}

// New returns a new Validator
func New() *Validator {
	return &Validator{Errors: make(map[string]string)}
}

// Valid checks if there are any errors
func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

// AddError adds an error to errors array if it does not exist
func (v *Validator) AddError(key, message string) {
	if _, exists := v.Errors[key]; !exists {
		v.Errors[key] = message
	}
}

// Check adds an error message if the validation is not 'ok'
func (v *Validator) Check(ok bool, key, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// PermittedValue checks that a value is within a slice
func PermittedValue[T comparable](value T, permittedValues ...T) bool {
	return slices.Contains(permittedValues, value)
}

// Matches checks a string to a regex and returns true if it matches
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// Unique checks that all values in a given slice are unique
func Unique[T comparable](values []T) bool {
	uniqueValues := make(map[T]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(values) == len(uniqueValues)
}
