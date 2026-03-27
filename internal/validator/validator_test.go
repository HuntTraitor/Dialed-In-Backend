package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	tests := []struct {
		name           string
		checkResult    bool
		inputError     map[string]string
		expectedErrors map[string]string
	}{
		{
			name:        "Adds errors when check is false",
			checkResult: false,
			inputError: map[string]string{
				"error": "error message",
			},
			expectedErrors: map[string]string{
				"error": "error message",
			},
		},
		{
			name:        "Does not add errors when check is true",
			checkResult: true,
			inputError: map[string]string{
				"error": "error message",
			},
			expectedErrors: map[string]string{
				"": "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.Check(tt.checkResult, "error", tt.inputError["error"])
			assert.Equal(t, tt.expectedErrors["error"], v.Errors["error"])
		})
	}
}

func TestEmailRegex(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{
			name:  "test@gmail.com - Valid",
			email: "test@gmail.com",
			valid: true,
		},
		{
			name:  "testgmail.com - Valid",
			email: "testgmail.com",
			valid: false,
		},
		{
			name:  "test@gmail - Invalid",
			email: "test@gmail",
			valid: true,
		},
		{
			name:  "someemail - Invalid",
			email: "someemail",
			valid: false,
		},
		{
			name:  "@gmail.com - Invalid",
			email: "@gmail.com",
			valid: false,
		},
		{
			name:  "hello.World@gmail. - Invalid",
			email: "hello.World@gmail.",
			valid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, Matches(tt.email, EmailRX))
		})
	}
}

func TestUrlRegex(t *testing.T) {
	tests := []struct {
		name  string
		url   string
		valid bool
	}{
		{
			name:  "https://example.com - Valid",
			url:   "https://example.com",
			valid: true,
		},
		{
			name:  "http://example.com - Valid",
			url:   "http://example.com",
			valid: true,
		},
		{
			name:  "https://examplecom - Invalid",
			url:   "https://examplecom",
			valid: false,
		},
		{
			name:  "https:/example.com - Invalid",
			url:   "https:/example.com",
			valid: false,
		},
		{
			name:  "Empty - Invalid",
			url:   "",
			valid: false,
		},
		{
			name:  "example - Invalid",
			url:   "example",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, Matches(tt.url, UrlRX))
		})
	}
}

func TestTemperatureRegex(t *testing.T) {
	tests := []struct {
		name        string
		temperature string
		valid       bool
	}{
		{
			name:        "100 - Invalid",
			temperature: "100",
			valid:       false,
		},
		{
			name:        "100°C - Valid",
			temperature: "100°C",
			valid:       true,
		},
		{
			name:        "100°F - Valid",
			temperature: "100°F",
			valid:       true,
		},
		{
			name:        "Empty - Invalid",
			temperature: "",
			valid:       false,
		},
		{
			name:        "°F - Invalid",
			temperature: "°F",
			valid:       false,
		},
		{
			name:        "°C - Invalid",
			temperature: "°C",
			valid:       false,
		},
		{
			name:        "hi°C - invalid",
			temperature: "hi°C",
			valid:       false,
		},
		{
			name:        "hi°F - invalid",
			temperature: "hi°F",
			valid:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, Matches(tt.temperature, TempRX))
		})
	}

}
