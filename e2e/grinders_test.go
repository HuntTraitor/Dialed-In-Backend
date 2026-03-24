package e2e

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/hunttraitor/dialed-in-backend/e2e/testutils"
)

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

func TestGetAllGrinders(t *testing.T) {
	app := testutils.NewTestApp(t)

	t.Run("Successfully gets all grinders", func(t *testing.T) {
		user := app.Factory.CreateUser(t)
		token := app.Factory.Login(t, user.Email, user.Password)

		// post 2 grinders
		for i := 0; i < 3; i++ {
			app.Factory.CreateGrinder(t, token, testutils.ValidGrinder())
		}

		res := app.Client(token).GET("/v1/grinders").Expect(t)
		arr := res.Status(http.StatusOK).JSON().Object().Value("grinders").Array()
		arr.Length().Equal(3)

		for i := 0; i < 3; i++ {
			arr.Element(i).Object().Value("name").String().Equal(testutils.ValidGrinder().Name)
		}
	})

	t.Run("No grinders returns an empty array", func(t *testing.T) {
		user := app.Factory.CreateUser(t)
		token := app.Factory.Login(t, user.Email, user.Password)
		res := app.Client(token).GET("/v1/grinders").Expect(t)
		res.Status(http.StatusOK).JSON().Object().Value("grinders").Array().Empty()
	})

	t.Run("Not authenticated returns an error", func(t *testing.T) {
		app.Client("").GET("/v1/grinders").Expect(t).Status(http.StatusUnauthorized)
	})
}

func TestGetOneGrinder(t *testing.T) {
	app := testutils.NewTestApp(t)
	user := app.Factory.CreateUser(t)
	token := app.Factory.Login(t, user.Email, user.Password)

	tests := []struct {
		name   string
		mutate func(*int64)
		assert func(*httpexpect.Response, testutils.FixtureGrinder)
	}{
		{
			name:   "Successfully gets one grinder",
			mutate: func(id *int64) {},
			assert: func(res *httpexpect.Response, input testutils.FixtureGrinder) {
				grinder := res.Status(http.StatusOK).JSON().Object().Value("grinder").Object()
				grinder.Value("id").Number().Equal(input.ID)
				grinder.Value("name").String().Equal(input.Name)
				grinder.Value("created_at").String().Equal(input.CreatedAt)
				grinder.Value("version").Number().Equal(input.Version)
				grinder.Value("user_id").Number().Equal(input.UserID)
			},
		},
		{
			name: "Invalid ID returns not found",
			mutate: func(id *int64) {
				*id = 0
			},
			assert: func(res *httpexpect.Response, input testutils.FixtureGrinder) {
				res.Status(http.StatusNotFound).JSON().Object().Value("error").String().NotEmpty()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grinder := app.Factory.CreateGrinder(t, token, testutils.ValidGrinder())
			tt.mutate(&grinder.ID)
			res := app.Client(token).GET(fmt.Sprintf("/v1/grinders/%d", grinder.ID)).Expect(t)
			tt.assert(res, grinder)
		})
	}

	t.Run("Not authenticated returns an error", func(t *testing.T) {
		app.Client("").GET("/v1/grinders/1").Expect(t).Status(http.StatusUnauthorized)
	})

	t.Run("Grinder that is unowned returns a 404", func(t *testing.T) {
		grinder := app.Factory.CreateGrinder(t, token, testutils.ValidGrinder())
		newUser := app.Factory.CreateUser(t)
		newToken := app.Factory.Login(t, newUser.Email, newUser.Password)

		app.Client(newToken).GET(fmt.Sprintf("/v1/grinders/%d", grinder.ID)).Expect(t).Status(http.StatusNotFound)
	})

}

func TestPatchGrinder(t *testing.T) {
	app := testutils.NewTestApp(t)
	user := app.Factory.CreateUser(t)
	token := app.Factory.Login(t, user.Email, user.Password)

	tests := []struct {
		name   string
		mutate func(*testutils.PatchGrinderRequest)
		assert func(*httpexpect.Response, testutils.PatchGrinderRequest)
	}{
		{
			name:   "Patching a grinder is successful",
			mutate: func(req *testutils.PatchGrinderRequest) {},
			assert: func(res *httpexpect.Response, input testutils.PatchGrinderRequest) {
				grinder := res.Status(http.StatusOK).JSON().Object().Value("grinder").Object()
				grinder.Value("id").Number().Gt(0)
				grinder.Value("user_id").Number().Equal(user.ID)
				grinder.Value("name").String().Equal(input.Name)
				grinder.Value("created_at").String().NotEmpty()
				grinder.Value("version").Number().Equal(2)

				newGrinder := app.Client(token).GET(fmt.Sprintf("/v1/grinders/%d", int(grinder.Value("id").Number().Raw()))).
					Expect(t).Status(http.StatusOK).JSON().Object().Value("grinder").Object()

				newGrinder.Value("name").String().Equal(input.Name)
			},
		},
		{
			name: "Name too long returns an error",
			mutate: func(req *testutils.PatchGrinderRequest) {
				req.Name = strings.Repeat("a", 101)
			},
			assert: func(res *httpexpect.Response, input testutils.PatchGrinderRequest) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("name").String().Contains("100")
			},
		},
		{
			name: "No body returns an error",
			mutate: func(req *testutils.PatchGrinderRequest) {
				req.Name = ""
			},
			assert: func(res *httpexpect.Response, input testutils.PatchGrinderRequest) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("name").String().NotEmpty()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grinder := app.Factory.CreateGrinder(t, token, testutils.ValidGrinder())

			req := testutils.ValidPatchGrinder()
			tt.mutate(&req)
			res := app.Client(token).PATCHJSON(fmt.Sprintf("/v1/grinders/%d", grinder.ID), req).Expect(t)
			tt.assert(res, req)

		})
	}

	t.Run("Patching a grinder that does not exist returns an error", func(t *testing.T) {
		request := testutils.CreateGrinderRequest{
			Name: testutils.ValidGrinder().Name,
		}
		app.Client(token).PATCHJSON("/v1/grinders/0", request).Expect(t).Status(http.StatusNotFound)
	})

	t.Run("Patching a grinder that you dont own returns an error", func(t *testing.T) {

		request := testutils.CreateGrinderRequest{
			Name: testutils.ValidGrinder().Name,
		}

		grinder := app.Factory.CreateGrinder(t, token, request)
		newUser := app.Factory.CreateUser(t)
		newUserToken := app.Factory.Login(t, newUser.Email, newUser.Password)

		app.Client(newUserToken).PATCHJSON(fmt.Sprintf("/v1/grinders/%d", grinder.ID), request).Expect(t).Status(http.StatusNotFound)
	})

	t.Run("Patching a grinder when not logged in returns an error", func(t *testing.T) {
		request := testutils.CreateGrinderRequest{
			Name: testutils.ValidGrinder().Name,
		}
		grinder := app.Factory.CreateGrinder(t, token, request)
		app.Client("").PATCHJSON(fmt.Sprintf("/v1/grinders/%d", grinder.ID), request).Expect(t).Status(http.StatusUnauthorized)
	})
}

func TestDeleteGrinder(t *testing.T) {
	app := testutils.NewTestApp(t)
	user := app.Factory.CreateUser(t)
	token := app.Factory.Login(t, user.Email, user.Password)

	tests := []struct {
		name   string
		mutate func(*int64)
		assert func(*httpexpect.Response, testutils.FixtureGrinder)
	}{
		{
			name:   "Successfully deletes a grinder",
			mutate: func(id *int64) {},
			assert: func(res *httpexpect.Response, input testutils.FixtureGrinder) {
				res.Status(http.StatusOK).JSON().Object().Value("message").String().Contains("deleted")
				app.Client(token).GET(fmt.Sprintf("/v1/grinders/%d", input.ID)).Expect(t).Status(http.StatusNotFound)
			},
		},
		{
			name: "Deleting a grinder that does not exist returns not found",
			mutate: func(id *int64) {
				*id = 0
			},
			assert: func(res *httpexpect.Response, input testutils.FixtureGrinder) {
				res.Status(http.StatusNotFound).JSON().Object().Value("error").String().NotEmpty()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grinder := app.Factory.CreateGrinder(t, token, testutils.ValidGrinder())
			tt.mutate(&grinder.ID)
			res := app.Client(token).DELETE(fmt.Sprintf("/v1/grinders/%d", grinder.ID)).Expect(t)
			tt.assert(res, grinder)
		})
	}

	t.Run("Deleting a grinder that you dont own returns not found", func(t *testing.T) {
		grinder := app.Factory.CreateGrinder(t, token, testutils.ValidGrinder())
		newUser := app.Factory.CreateUser(t)
		newToken := app.Factory.Login(t, newUser.Email, newUser.Password)
		app.Client(newToken).DELETE(fmt.Sprintf("/v1/grinders/%d", grinder.ID)).Expect(t).Status(http.StatusNotFound)
	})

	t.Run("Deleting a grinder when not logged in returns unauthorized", func(t *testing.T) {
		grinder := app.Factory.CreateGrinder(t, token, testutils.ValidGrinder())
		app.Client("").DELETE(fmt.Sprintf("/v1/grinders/%d", grinder.ID)).Expect(t).Status(http.StatusUnauthorized)
	})
}
