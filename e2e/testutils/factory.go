package testutils

type FixtureFactory struct {
	BaseURL string
}

type FixtureUser struct {
	ID       int64
	Name     string
	Email    string
	Password string
}
