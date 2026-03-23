package e2e

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/hunttraitor/dialed-in-backend/e2e/testutils"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticateUser(t *testing.T) {
	app := testutils.NewTestApp(t)

	tests := []struct {
		name   string
		mutate func(request *testutils.AuthRequest)
		assert func(*httpexpect.Response)
	}{
		{
			name:   "Successfully authenticates user",
			mutate: func(request *testutils.AuthRequest) {},
			assert: func(res *httpexpect.Response) {
				auth := res.Status(http.StatusCreated).JSON().Object().Value("authentication_token").Object()
				auth.Value("token").String().NotEmpty()
				auth.Value("expiry").String().NotEmpty()
			},
		},
		{
			name: "Incorrect email returns an error",
			mutate: func(request *testutils.AuthRequest) {
				request.Email = "invalid@example.com"
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnauthorized).JSON().Object().Value("error").String().NotEmpty()
			},
		},
		{
			name: "Incorrect password returns an error",
			mutate: func(request *testutils.AuthRequest) {
				request.Password = "invalid123123"
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnauthorized).JSON().Object().Value("error").String().NotEmpty()
			},
		},
		{
			name: "Password is too short returns an error",
			mutate: func(request *testutils.AuthRequest) {
				request.Password = "123"
			},
			assert: func(res *httpexpect.Response) {
				pwd := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				pwd.Value("password").String().NotEmpty()
			},
		},
		{
			name: "Bad email returns an error",
			mutate: func(request *testutils.AuthRequest) {
				request.Email = "invalid"
			},
			assert: func(res *httpexpect.Response) {
				email := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				email.Value("email").String().NotEmpty()
			},
		},
		{
			name: "Missing email and password returns an error",
			mutate: func(request *testutils.AuthRequest) {
				request.Email = ""
				request.Password = ""
			},
			assert: func(res *httpexpect.Response) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("email").String().NotEmpty()
				err.Value("password").String().NotEmpty()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := app.Factory.CreateUser(t)
			authRequest := testutils.AuthRequest{
				Email:    user.Email,
				Password: user.Password,
			}
			tt.mutate(&authRequest)

			res := app.Client("").POSTJSON("/v1/tokens/authentication", authRequest).Expect(t)
			tt.assert(res)
		})
	}
}

// successfully recieves password reset email
// bad email doesnt recieve password reset email
// email not existing still gets same response  but doesnt send email

func TestResetPasswordEmailSent(t *testing.T) {
	app := testutils.NewTestApp(t)

	user := app.Factory.CreateUser(t)

	tests := []struct {
		name   string
		mutate func(request *testutils.PasswordResetEmailRequest)
		assert func(*testing.T, *httpexpect.Response)
	}{
		{
			name:   "Successfully sends password reset email",
			mutate: func(request *testutils.PasswordResetEmailRequest) {},
			assert: func(t *testing.T, res *httpexpect.Response) {
				res.Status(http.StatusCreated).JSON().Object().Value("message").String().Contains("sent")

				token := testutils.AssertPasswordResetToken(t, user.Email)
				assert.NotEmpty(t, token)
			},
		},
		{
			name: "Invalid email doesnt receive password reset email",
			mutate: func(request *testutils.PasswordResetEmailRequest) {
				request.Email = "invalidexample.com"
			},
			assert: func(t *testing.T, res *httpexpect.Response) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("email").String().NotEmpty()
			},
		},
		{
			name: "Email not existing receives json but doesnt send email",
			mutate: func(request *testutils.PasswordResetEmailRequest) {
				request.Email = "nonexistent@noexist.com"
			},
			assert: func(t *testing.T, res *httpexpect.Response) {
				res.Status(http.StatusCreated).JSON().Object().Value("message").String().Contains("sent")
				testutils.AssertNoPasswordResetToken(t, "nonexistent@noexist.com")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := testutils.PasswordResetEmailRequest{
				Email: user.Email,
			}
			tt.mutate(&request)
			res := app.Client("").POSTJSON("/v1/tokens/password-reset", request).Expect(t)
			tt.assert(t, res)
		})
	}
}
