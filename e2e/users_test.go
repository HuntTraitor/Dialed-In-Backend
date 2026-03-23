package e2e

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/hunttraitor/dialed-in-backend/e2e/testutils"
)

func TestCreateUser(t *testing.T) {
	app := testutils.NewTestApp(t)

	tests := []struct {
		name   string
		mutate func(*testutils.CreateUserInput)
		assert func(*testing.T, *httpexpect.Response, testutils.CreateUserInput)
	}{
		{
			name:   "Creating a new user will send a 201 and send an email",
			mutate: func(input *testutils.CreateUserInput) {},
			assert: func(t *testing.T, res *httpexpect.Response, input testutils.CreateUserInput) {
				user := res.Status(http.StatusCreated).JSON().Object().Value("user").Object()

				user.Value("id").Number().Gt(0)
				user.Value("created_at").String().NotEmpty()
				user.Value("name").String().Equal(input.Name)
				user.Value("email").String().Equal(input.Email)
				user.Value("activated").Boolean().False()

				testutils.AssertEmailSent(t, "to", user.Value("email").String().Raw())
			},
		},
		{
			name: "Creating a new user without any fields returns a bad input",
			mutate: func(input *testutils.CreateUserInput) {
				input.Name = ""
				input.Email = ""
				input.Password = ""
			},
			assert: func(t *testing.T, res *httpexpect.Response, input testutils.CreateUserInput) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("email").String().NotEmpty()
				err.Value("name").String().NotEmpty()
				err.Value("password").String().NotEmpty()
			},
		},
		{
			name: "Creating a new user with a bad email returns a bad input",
			mutate: func(input *testutils.CreateUserInput) {
				input.Email = "testexample.com"
			},
			assert: func(t *testing.T, res *httpexpect.Response, input testutils.CreateUserInput) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("email").String().NotEmpty()
			},
		},
		{
			name: "Creating a new user with a short password returns a bad input",
			mutate: func(input *testutils.CreateUserInput) {
				input.Password = "1234567"
			},
			assert: func(t *testing.T, res *httpexpect.Response, input testutils.CreateUserInput) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("password").String().Contains("8")
			},
		},
		{
			name: "Creating a new user with a too long password returns a bad input",
			mutate: func(input *testutils.CreateUserInput) {
				input.Password = strings.Repeat("a", 73)
			},
			assert: func(t *testing.T, res *httpexpect.Response, input testutils.CreateUserInput) {
				t.Log(input)
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("password").String().Contains("72")
			},
		},
		{
			name: "Creating a new user with a too long name returns a bad input",
			mutate: func(input *testutils.CreateUserInput) {
				input.Name = strings.Repeat("a", 501)
			},
			assert: func(t *testing.T, res *httpexpect.Response, input testutils.CreateUserInput) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("name").String().Contains("500")
			},
		},
		{
			name:   "Creating a new user with a duplicate email address returns a status conflict",
			mutate: func(input *testutils.CreateUserInput) {},
			assert: func(t *testing.T, res *httpexpect.Response, input testutils.CreateUserInput) {
				user := res.Status(http.StatusCreated).JSON().Object().Value("user").Object()
				newUser := testutils.CreateUserInput{
					Name:     input.Name,
					Email:    user.Value("email").String().Raw(),
					Password: input.Password,
				}
				err := app.Client("").POSTJSON("/v1/users", newUser).Expect(t).Status(http.StatusUnprocessableEntity).
					JSON().Object().Value("error").Object()

				err.Value("email").String().Contains("exists")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := testutils.ValidUser()
			tt.mutate(&input)
			res := app.Client("").POSTJSON("/v1/users", input).Expect(t)
			tt.assert(t, res, input)
		})
	}
}

func TestVerifyUser(t *testing.T) {
	app := testutils.NewTestApp(t)

	tests := []struct {
		name   string
		mutate func(token *string)
		assert func(*testing.T, *httpexpect.Response, testutils.FixtureUser)
	}{
		{
			name:   "Successfully verifies user",
			mutate: func(token *string) {},
			assert: func(t *testing.T, res *httpexpect.Response, user testutils.FixtureUser) {
				resp := res.Status(http.StatusOK).JSON().Object().Value("user").Object()

				resp.Value("id").Number().Equal(user.ID)
				resp.Value("name").String().Equal(user.Name)
				resp.Value("email").String().Equal(user.Email)
				resp.Value("created_at").String().NotEmpty()
				resp.Value("activated").Boolean().False()
			},
		},
		{
			name: "No Token provided returns unauthorized error",
			mutate: func(token *string) {
				*token = ""
			},
			assert: func(t *testing.T, res *httpexpect.Response, user testutils.FixtureUser) {
				res.Status(http.StatusUnauthorized).JSON().Object().Value("error").String().NotEmpty()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := app.Factory.CreateUser(t)
			token := app.Factory.Login(t, user.Email, user.Password)

			tt.mutate(&token)
			res := app.Client(token).GET("/v1/users/verify").Expect(t)
			tt.assert(t, res, user)
		})
	}

}

func TestResetPassword(t *testing.T) {
	app := testutils.NewTestApp(t)

	tests := []struct {
		name   string
		mutate func(*testutils.PasswordResetRequest)
		assert func(*testing.T, *httpexpect.Response, testutils.FixtureUser, string)
	}{
		{
			name:   "Successfully resets password",
			mutate: func(request *testutils.PasswordResetRequest) {},
			assert: func(t *testing.T, res *httpexpect.Response, user testutils.FixtureUser, newPassword string) {
				res.Status(http.StatusOK).JSON().Object().Value("message").String().Contains("reset")

				// expect the old password to be rejected
				loginWithOldPassword := app.Client("").POSTJSON("/v1/tokens/authentication", testutils.AuthRequest{
					Email:    user.Email,
					Password: user.Password,
				})
				loginWithOldPassword.Expect(t).Status(http.StatusUnauthorized)

				// expect the new password to be accepted
				loginWithNewPassword := app.Client("").POSTJSON("/v1/tokens/authentication", testutils.AuthRequest{
					Email:    user.Email,
					Password: newPassword,
				})
				loginWithNewPassword.Expect(t).Status(http.StatusCreated)
			},
		},
		{
			name: "Empty body returns errors",
			mutate: func(request *testutils.PasswordResetRequest) {
				request.Token = ""
				request.Password = ""
			},
			assert: func(t *testing.T, res *httpexpect.Response, user testutils.FixtureUser, newPassword string) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("token").String().NotEmpty()
				err.Value("password").String().NotEmpty()
			},
		},
		{
			name: "Too long password gets rejected",
			mutate: func(request *testutils.PasswordResetRequest) {
				request.Password = strings.Repeat("a", 73)
			},
			assert: func(t *testing.T, res *httpexpect.Response, user testutils.FixtureUser, newPassword string) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("password").String().Contains("72")
			},
		},
		{
			name: "Too short password gets rejected",
			mutate: func(request *testutils.PasswordResetRequest) {
				request.Password = "1234567"
			},
			assert: func(t *testing.T, res *httpexpect.Response, user testutils.FixtureUser, newPassword string) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("password").String().Contains("8")
			},
		},
		{
			name: "Incorrect token gets rejected",
			mutate: func(request *testutils.PasswordResetRequest) {
				request.Token = "000000"
			},
			assert: func(t *testing.T, res *httpexpect.Response, user testutils.FixtureUser, newPassword string) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("token").String().Contains("invalid")
			},
		},
		{
			name: "Invalid token format gets rejected",
			mutate: func(request *testutils.PasswordResetRequest) {
				request.Token = "0"
			},
			assert: func(t *testing.T, res *httpexpect.Response, user testutils.FixtureUser, newPassword string) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("token").String().Contains("6")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := app.Factory.CreateUser(t)

			app.Factory.SendResetPasswordEmail(t, user.Email)
			token := testutils.AssertPasswordResetToken(t, user.Email)

			req := testutils.PasswordResetRequest{
				Token:    token,
				Password: testutils.ValidPassword(),
			}

			tt.mutate(&req)
			res := app.Client("").PUTJSON("/v1/users/password", req).Expect(t)
			tt.assert(t, res, user, req.Password)
		})
	}
}
