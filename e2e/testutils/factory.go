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

type FixtureGrinder struct {
	ID        int64
	UserID    int64
	Name      string
	CreatedAt string
	Version   int64
}
