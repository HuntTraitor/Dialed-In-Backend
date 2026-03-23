package testutils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type CreateGrinderResponse struct {
	Grinder struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"grinder"`
}

func (f *FixtureFactory) CreateGrinder(t *testing.T, token string, name string) CreateGrinderResponse {
	t.Helper()

	if name == "" {
		name = "Test Grinder"
	}

	res := (&APIClient{BaseURL: f.BaseURL, Token: token}).
		POSTJSON("/v1/grinders", map[string]any{
			"name": name,
		}).Expect(t)

	res.Status(http.StatusCreated)
	var body CreateGrinderResponse
	DecodeJSON(t, res, &body)

	require.NotZero(t, body.Grinder.ID)
	return body
}
