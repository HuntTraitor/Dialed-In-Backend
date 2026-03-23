package e2e

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/hunttraitor/dialed-in-backend/e2e/testutils"
)

//func TestResetPassword(t *testing.T) {
//	cleanup, _, err := testutils.LaunchTestProgram(port)
//	if err != nil {
//		t.Fatalf("failed to launch test program: %v", err)
//	}
//	t.Cleanup(cleanup)
//
//	_ = createUser(t)
//
//	t.Run("Succesfully Resets Password", func(t *testing.T) {
//
//		// Check that you can log in with the old password
//		requestURL := fmt.Sprintf("http://localhost:%d/v1/tokens/authentication", 3001)
//		requestBody := `{"email": "test@example.com", "password": "password"}`
//		statusCode, _, _ := post(t, requestURL, strings.NewReader(requestBody), nil)
//		assert.Equal(t, http.StatusCreated, statusCode)
//
//		// Send the request to reset
//		requestURL = fmt.Sprintf("http://localhost:%d/v1/tokens/password-reset", 3001)
//		requestBody = `{"email": "test@example.com"}`
//
//		statusCode, _, returnedBody := post(t, requestURL, strings.NewReader(requestBody), nil)
//		expectedResponse := map[string]any{
//			"message": "an email will be sent to you containing password reset instructions",
//		}
//
//		assert.Equal(t, http.StatusCreated, statusCode)
//		assert.Equal(t, expectedResponse, returnedBody)
//
//		// Get the reset token from the email
//		var token string
//		waitFor(t, func() bool {
//			body, _ := getEmail(t, "containing", "password%20reset%20token")
//			token = extractToken(t, body)
//			return token != ""
//		})
//
//		// Send a request to the reset password
//		requestURL = fmt.Sprintf("http://localhost:%d/v1/users/password", 3001)
//		requestBody = fmt.Sprintf(`{"password": "password2", "token": "%s"}`, token)
//
//		statusCode, _, returnedBody = put(t, requestURL, strings.NewReader(requestBody))
//		assert.Equal(t, http.StatusOK, statusCode)
//		expectedResponse = map[string]any{
//			"message": "your password was successfully reset",
//		}
//		assert.Equal(t, expectedResponse, returnedBody)
//
//		// Check that you cannot log in with the old password
//		requestURL = fmt.Sprintf("http://localhost:%d/v1/tokens/authentication", 3001)
//		requestBody = `{"email": "test@example.com", "password": "password"}`
//		statusCode, _, _ = post(t, requestURL, strings.NewReader(requestBody), nil)
//		assert.Equal(t, http.StatusUnauthorized, statusCode)
//
//		//Check that you can log in with the new password
//		requestURL = fmt.Sprintf("http://localhost:%d/v1/tokens/authentication", 3001)
//		requestBody = `{"email": "test@example.com", "password": "password2"}`
//		statusCode, _, _ = post(t, requestURL, strings.NewReader(requestBody), nil)
//		assert.Equal(t, http.StatusCreated, statusCode)
//	})
//}

// Creating a new user will send a 201 and send an email
// Creating a new user without any fields rejects
// Creating a new user with a bad email rejects
// Creating a new user with a short password rejects
// Creating a new user with a too long password rejects
// Creating a new user with a too long name rejects
// Creating a new user with a duplicate email rejects

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
