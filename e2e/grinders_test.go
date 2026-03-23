package e2e

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/hunttraitor/dialed-in-backend/e2e/testutils"
)

// POST grinder successful
// POST grinder missing fields fails
// POST grinder name too long fails

// POST grinder not authenticated fails
func TestPostGrinder(t *testing.T) {
	app := testutils.NewTestApp(t)
	user := app.Factory.CreateUser(t)
	token := app.Factory.Login(t, user.Email, user.Password)

	tests := []struct {
		name   string
		mutate func(request *testutils.CreateGrinderRequest)
		assert func(*testing.T, *httpexpect.Response, testutils.CreateGrinderRequest)
	}{
		{
			name:   "Successfully creates a grinder",
			mutate: func(request *testutils.CreateGrinderRequest) {},
			assert: func(t *testing.T, res *httpexpect.Response, input testutils.CreateGrinderRequest) {
				grinder := res.Status(http.StatusCreated).JSON().Object().Value("grinder").Object()
				grinder.Value("id").Number().Gt(0)
				grinder.Value("user_id").Number().Equal(user.ID)
				grinder.Value("name").String().Equal(input.Name)
				grinder.Value("created_at").String().NotEmpty()
				grinder.Value("version").Number().Equal(1)
			},
		},
		{
			name: "Missing fields returns an error",
			mutate: func(request *testutils.CreateGrinderRequest) {
				request.Name = ""
			},
			assert: func(t *testing.T, res *httpexpect.Response, input testutils.CreateGrinderRequest) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("name").String().NotEmpty()
			},
		},
		{
			name: "Name too long returns an error",
			mutate: func(request *testutils.CreateGrinderRequest) {
				request.Name = strings.Repeat("a", 101)
			},
			assert: func(t *testing.T, res *httpexpect.Response, input testutils.CreateGrinderRequest) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("name").String().Contains("100")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grinder := testutils.ValidGrinder()
			tt.mutate(&grinder)
			res := app.Client(token).POSTJSON("/v1/grinders", grinder).Expect(t)
			tt.assert(t, res, grinder)
		})
	}

	t.Run("Not authenticated returns an error", func(t *testing.T) {
		app.Client("").POSTJSON("/v1/grinders", testutils.ValidGrinder()).Expect(t).Status(http.StatusUnauthorized)
	})
}
