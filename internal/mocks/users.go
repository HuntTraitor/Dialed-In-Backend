package mocks

import "github.com/hunttraitor/dialed-in-backend/internal/data"

type MockUserModel struct{}

func (m MockUserModel) Insert(user *data.User) error {
	switch user.Email {
	case "dupe@example.com":
		return data.ErrDuplicateEmail
	default:
		return nil
	}
}
