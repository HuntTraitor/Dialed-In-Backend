package e2e

import (
	"testing"

	"github.com/hunttraitor/dialed-in-backend/e2e/testutils"
	_ "github.com/lib/pq"
)

func TestHealthcheck(t *testing.T) {
	app := testutils.NewTestApp(t)

	t.Run("App is healthy", func(t *testing.T) {
		res := app.Client("").GET("/v1/healthcheck").Expect(t)

		obj := res.JSON().Object()

		obj.Value("status").String().Equal("available")
		obj.Value("system_info").Object().Value("environment").String().Equal("test")
	})
}
