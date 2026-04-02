package testutils

import (
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

type TestApp struct {
	BaseURL string
	Factory *FixtureFactory
}

func NewTestApp(t *testing.T) *TestApp {
	t.Helper()

	port := getFreePort(t)
	cleanup, _, err := LaunchTestProgram(port)
	require.NoError(t, err)
	t.Cleanup(cleanup)

	app := &TestApp{
		BaseURL: fmt.Sprintf("http://localhost:%s", port),
	}

	app.Factory = &FixtureFactory{
		BaseURL: app.BaseURL,
	}

	return app
}

func getFreePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", ":0")
	require.NoError(t, err)
	port := l.Addr().(*net.TCPAddr).Port
	l.Close()
	return fmt.Sprintf("%d", port)
}

func (a *TestApp) Client(token string) *APIClient {
	return &APIClient{
		BaseURL: a.BaseURL,
		Token:   token,
	}
}
