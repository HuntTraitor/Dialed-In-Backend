package data

import (
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	t.Run("should generate a token", func(t *testing.T) {
		token, err := generateToken(1, 3*time.Second, "scope")
		assert.NoError(t, err)
		assert.Equal(t, int64(1), token.UserID)
		assert.NotEmpty(t, token.Plaintext)
		assert.NotEmpty(t, token.Hash)
		assert.NotEmpty(t, token.Expiry)
		assert.Equal(t, "scope", token.Scope)
	})
}

func TestValidatePlainText(t *testing.T) {
	test := []struct {
		name           string
		tokenPlainText string
		expectedErrors map[string]string
	}{
		{
			name:           "Inserts token successfully",
			tokenPlainText: "ASDJKLEPOIURERFJDKSLAIEJGH",
			expectedErrors: map[string]string{},
		},
		{
			name:           "No token provided",
			tokenPlainText: "",
			expectedErrors: map[string]string{
				"token": "must be provided",
			},
		},
		{
			name:           "Token too short",
			tokenPlainText: strings.Repeat("a", 2),
			expectedErrors: map[string]string{
				"token": "must be 26 bytes long",
			},
		},
		{
			name:           "Token too long",
			tokenPlainText: strings.Repeat("a", 6),
			expectedErrors: map[string]string{
				"token": "must be 26 bytes long",
			},
		},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()

			ValidatePlainText(v, tt.tokenPlainText)
			assert.Equal(t, tt.expectedErrors, v.Errors)
		})
	}
}
