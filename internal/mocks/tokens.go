package mocks

import (
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"time"
)

type MockTokenModel struct {
	TokenCreated int
}

func (m *MockTokenModel) New(userID int64, ttl time.Duration, scope string) (*data.Token, error) {
	err := m.Insert(nil)
	return &data.Token{
		Plaintext: "1234",
	}, err
}

func (m *MockTokenModel) Insert(token *data.Token) error {
	m.TokenCreated++
	return nil
}

func (m *MockTokenModel) DeleteAllForUser(scope string, userID int64) error {
	// TODO test DeleteAllForUser
	return nil
}
