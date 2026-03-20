package mocks

import "github.com/hunttraitor/dialed-in-backend/internal/data"

type MockGrinderModel struct{}

func (m MockGrinderModel) Insert(grinder *data.Grinder) error {
	return nil
}

func (m MockGrinderModel) GetOne(id int64, userId int64) (*data.Grinder, error) {
	return nil, nil
}

func (m MockGrinderModel) GetAllForUser(userId int64) ([]*data.Grinder, error) {
	return []*data.Grinder{}, nil
}
