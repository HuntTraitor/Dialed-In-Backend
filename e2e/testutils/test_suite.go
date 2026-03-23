package testutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type TestApp struct {
	BaseURL string
	Factory *FixtureFactory
}

func NewTestApp(t *testing.T) *TestApp {
	t.Helper()

	cleanup, _, err := LaunchTestProgram("3001")
	require.NoError(t, err)

	t.Cleanup(cleanup)

	app := &TestApp{
		BaseURL: "http://localhost:3001",
	}

	app.Factory = &FixtureFactory{
		BaseURL: app.BaseURL,
	}

	return app
}

func (a *TestApp) Client(token string) *APIClient {
	return &APIClient{
		BaseURL: a.BaseURL,
		Token:   token,
	}
}
