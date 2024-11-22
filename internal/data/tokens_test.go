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

			ValidateTokenPlainText(v, tt.tokenPlainText)
			assert.Equal(t, tt.expectedErrors, v.Errors)
		})
	}
}

func TestNewToken(t *testing.T) {
	db := newTestDB(t)
	tm := TokenModel{db}
	m := UserModel{db}
	err := m.Insert(&User{
		Name:  "Test User",
		Email: "testuser@example.com",
		Password: password{
			hash: []byte("hash"),
		},
		Activated: false,
	})
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		userID        int64
		ttl           time.Duration
		scope         string
		expectedToken *Token
		isError       bool
	}{
		{
			name:   "Successfully inserts token",
			userID: 1,
			ttl:    5 * time.Minute,
			scope:  "scope",
			expectedToken: &Token{
				UserID: 1,
				Scope:  "scope",
			},
			isError: false,
		},
		{
			name:          "Wont insert token on no associated user id",
			userID:        2,
			ttl:           5 * time.Minute,
			scope:         "scope",
			expectedToken: nil,
			isError:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := tm.New(tt.userID, tt.ttl, tt.scope)
			if tt.isError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken.UserID, token.UserID)
				assert.Equal(t, tt.expectedToken.Scope, token.Scope)
				assert.NotEmpty(t, token.Hash)
				assert.NotEmpty(t, token.Expiry)
				assert.NotEmpty(t, token.Plaintext)
			}
		})
	}
}

func TestDeleteAllTokensForUser(t *testing.T) {
	db := newTestDB(t)
	tm := TokenModel{db}
	m := UserModel{db}
	err := m.Insert(&User{
		Name:  "Test User",
		Email: "testuser@example.com",
		Password: password{
			hash: []byte("hash"),
		},
		Activated: false,
	})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		userID int64
		scope  string
	}{
		{
			name:   "Deletes all tokens for user",
			userID: 1,
			scope:  "scope",
		},
		{
			name:   "Deletes Tokens even if user ID cannot be found",
			userID: 12,
			scope:  "scope",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Insert two tokens for a user
			for i := 0; i < 2; i++ {
				_, err := tm.New(1, 5*time.Minute, tt.scope)
				if err != nil {
					t.Fatal(err)
				}
			}

			// TODO get tokens for user and check that they exist
			err = tm.DeleteAllForUser(tt.scope, tt.userID)
			assert.NoError(t, err)
			// TODO get tokens for user and check that they dont exist
		})
	}
}
