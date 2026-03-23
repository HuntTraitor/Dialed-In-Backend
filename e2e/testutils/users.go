package testutils

import (
	"net/http"
	"testing"
)

type CreateUserResponse struct {
	User struct {
		ID        int64  `json:"id"`
		CreatedAt string `json:"created_at"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		Activated bool   `json:"activated"`
	} `json:"user"`
}

func (f *FixtureFactory) CreateUser(t *testing.T) FixtureUser {
	t.Helper()

	user := FixtureUser{
		Name:     "Test User",
		Email:    uniqueEmail(),
		Password: "password123",
	}

	res := (&APIClient{BaseURL: f.BaseURL}).
		POSTJSON("/v1/users", map[string]any{
			"name":     user.Name,
			"email":    user.Email,
			"password": user.Password,
		}).Expect(t)

	res.Status(http.StatusCreated)

	var body CreateUserResponse
	DecodeJSON(t, res, &body)
	user.ID = body.User.ID
	return user
}
