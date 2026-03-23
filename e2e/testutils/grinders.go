package testutils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

type CreateGrinderResponse struct {
	Grinder struct {
		ID        int64  `json:"id"`
		UserId    int64  `json:"user_id"`
		CreatedAt string `json:"created_at"`
		Name      string `json:"name"`
		Version   int64  `json:"version"`
	} `json:"grinder"`
}

type CreateGrinderRequest struct {
	Name string `json:"name"`
}

type PatchGrinderRequest struct {
	Name string `json:"name"`
}

type PatchGrinderResponse struct {
	Grinder struct {
		ID        int64  `json:"id"`
		UserId    int64  `json:"user_id"`
		CreatedAt string `json:"created_at"`
		Name      string `json:"name"`
		Version   int64  `json:"version"`
	} `json:"grinder"`
}

func ValidGrinder() CreateGrinderRequest {
	return CreateGrinderRequest{
		Name: "Test Grinder",
	}
}

func ValidPatchGrinder() PatchGrinderRequest {
	return PatchGrinderRequest{
		Name: "Updated Grinder",
	}
}

func (f *FixtureFactory) CreateGrinder(t *testing.T, token string, request CreateGrinderRequest) FixtureGrinder {
	t.Helper()

	grinder := FixtureGrinder{
		Name: request.Name,
	}

	res := (&APIClient{BaseURL: f.BaseURL, Token: token}).
		POSTJSON("/v1/grinders", map[string]any{
			"name": grinder.Name,
		}).Expect(t)

	res.Status(http.StatusCreated)
	var body CreateGrinderResponse
	DecodeJSON(t, res, &body)

	require.NotZero(t, body.Grinder.ID)
	grinder.ID = body.Grinder.ID
	grinder.Name = body.Grinder.Name
	grinder.CreatedAt = body.Grinder.CreatedAt
	grinder.Version = body.Grinder.Version
	grinder.UserID = body.Grinder.UserId
	return grinder
}
