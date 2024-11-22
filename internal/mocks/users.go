package mocks

import (
	"errors"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
)

type MockUserModel struct{}

func (m MockUserModel) Insert(user *data.User) error {
	user.ID = 1
	switch user.Email {
	case "dupe@example.com":
		return data.ErrDuplicateEmail
	case "error@example.com":
		return errors.New("test error")
	default:
		return nil
	}
}

func (m MockUserModel) Update(user *data.User) error {
	// TODO test update
	return nil
}

func (m MockUserModel) GetForToken(tokenScope, tokenPlainText string) (*data.User, error) {
	// TODO test GetForToken
	return nil, nil
}

func (m MockUserModel) GetByEmail(email string) (*data.User, error) {
	// TODO test getbyemail
	return nil, nil
}
