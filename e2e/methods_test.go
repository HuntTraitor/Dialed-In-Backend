package e2e

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/hunttraitor/dialed-in-backend/e2e/testutils"
)

func TestGetAllMethods(t *testing.T) {
	app := testutils.NewTestApp(t)

	t.Run("Successfully gets all methods", func(t *testing.T) {
		res := app.Client("").GET("/v1/methods").Expect(t)

		obj := res.JSON().Object().Value("methods").Array()

		obj.NotEmpty()
		for i := 0; i < len(obj.Iter()); i++ {
			m := obj.Element(i).Object()
			m.Value("id").Number().Gt(0)
			m.Value("name").String().NotEmpty()
			m.Value("created_at").String().NotEmpty()
		}
	})
}

func TestGetOneMethod(t *testing.T) {

	app := testutils.NewTestApp(t)

	t.Run("Successfully gets one method", func(t *testing.T) {
		res := app.Client("").GET("/v1/methods").Expect(t)

		obj := res.JSON().Object().Value("methods").Array()
		obj.NotEmpty()
		for i := 0; i < len(obj.Iter()); i++ {
			m := obj.Element(i).Object()

			t.Log(m.Value("id").Number().Raw())

			id := int64(m.Value("id").Number().Raw())

			one := app.Client("").
				GET(fmt.Sprintf("/v1/methods/%d", id)).
				Expect(t).
				JSON().Object().Value("method").
				Object()

			one.Value("id").Number().Equal(m.Value("id").Number().Raw())
			one.Value("name").String().Equal(m.Value("name").String().Raw())
			s := m.Value("created_at").String().Raw()
			one.Value("created_at").String().Equal(s)
		}
	})

	t.Run("Method not found", func(t *testing.T) {
		app.Client("").GET("/v1/methods/0").Expect(t).Status(http.StatusNotFound)
	})
}

func TestMethodExists(t *testing.T) {
	app := testutils.NewTestApp(t)

	tests := []struct {
		name   string
		method string
		assert func(*httpexpect.Object)
	}{
		{
			name:   "Hario Switch",
			method: "Hario Switch",
			assert: func(method *httpexpect.Object) {
				method.Value("name").String().Equal("Hario Switch")
			},
		},
		{
			name:   "V60",
			method: "V60",
			assert: func(method *httpexpect.Object) {
				method.Value("name").String().Equal("V60")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			res := app.Client("").GET("/v1/methods").Expect(t)

			obj := res.JSON().Object().Value("methods").Array()
			obj.NotEmpty()

			var method *httpexpect.Object

			for _, v := range obj.Iter() {
				m := v.Object()

				if m.Value("name").String().Raw() == tt.method {
					method = m
					break
				}
			}

			if method == nil {
				t.Fatalf("Method %s not found", tt.method)
			}

			tt.assert(method)
		})
	}
}
