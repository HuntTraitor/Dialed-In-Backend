package testutils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/require"
)

type FixtureFactory struct {
	BaseURL string
}

type FixtureUser struct {
	ID       int64
	Name     string
	Email    string
	Password string
}

type CreateUserResponse struct {
	User struct {
		ID        int64  `json:"id"`
		CreatedAt string `json:"created_at"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		Activated bool   `json:"activated"`
	} `json:"user"`
}

type AuthResponse struct {
	AuthenticationToken struct {
		Token string `json:"token"`
	} `json:"authentication_token"`
}

type CreateCoffeeResponse struct {
	Coffee struct {
		ID int64 `json:"id"`
	} `json:"coffee"`
}

type CreateGrinderResponse struct {
	Grinder struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"grinder"`
}

type Method struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type ListMethodsResponse struct {
	Methods []Method `json:"methods"`
}

type HealthCheckResponse struct {
	Status     string `json:"status"`
	SystemInfo struct {
		Environment string `json:"environment"`
		Version     string `json:"version"`
	} `json:"system_info"`
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

func (f *FixtureFactory) CreateCoffee(t *testing.T, token string, form CoffeeForm) CreateCoffeeResponse {
	t.Helper()

	res := (&APIClient{BaseURL: f.BaseURL, Token: token}).
		POSTMultipart("/v1/coffees", form).Expect(t)

	res.Status(http.StatusCreated)

	var body CreateCoffeeResponse
	DecodeJSON(t, res, &body)

	require.NotZero(t, body.Coffee.ID)
	return body
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

func uniqueEmail() string {
	return fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())
}

func DecodeJSON(t *testing.T, res *httpexpect.Response, v any) {
	t.Helper()

	raw := res.JSON().Raw()

	b, err := json.Marshal(raw)
	require.NoError(t, err)

	err = json.Unmarshal(b, v)
	require.NoError(t, err)
}
