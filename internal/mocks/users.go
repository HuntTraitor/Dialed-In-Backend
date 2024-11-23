package mocks

import (
	"errors"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"time"
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
	switch user.Email {
	case "editconflict@example.com":
		return data.ErrEditConflict
	default:
		return nil
	}
}

func (m MockUserModel) GetForToken(tokenScope, tokenPlainText string) (*data.User, error) {
	switch tokenPlainText {
	case "ASDJKLEPOIURERFJDKSLAIEJG1":
		return &data.User{
			ID:        1,
			CreatedAt: time.Now().UTC(),
			Name:      "Test User",
			Email:     "test@example.com",
			Activated: false,
			Version:   1,
		}, nil
	case "ASDJKLEPOIURERFJDKSLAIEJG3":
		return &data.User{
			ID:        1,
			CreatedAt: time.Now().UTC(),
			Name:      "Test User",
			Email:     "editconflict@example.com",
			Activated: false,
			Version:   1,
		}, nil
	case "ASDJKLEPOIURERFJDKSLAIEJG2":
		return nil, data.ErrRecordNotFound
	default:
		return nil, nil
	}

}

func (m MockUserModel) GetByEmail(email string) (*data.User, error) {
	// TODO test getbyemail
	return nil, nil
}
