package testutils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AuthenticationToken struct {
		Token  string `json:"token"`
		Expiry string `json:"expiry"`
	} `json:"authentication_token"`
}

type PasswordResetRequest struct {
	Email string `json:"email"`
}

func (f *FixtureFactory) Login(t *testing.T, email, password string) string {
	t.Helper()

	res := (&APIClient{BaseURL: f.BaseURL}).
		POSTJSON("/v1/tokens/authentication", map[string]any{
			"email":    email,
			"password": password,
		}).Expect(t)

	res.Status(http.StatusCreated)

	var body AuthResponse
	DecodeJSON(t, res, &body)

	require.NotEmpty(t, body.AuthenticationToken.Token)
	return body.AuthenticationToken.Token
}
