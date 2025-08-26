package mocks

import (
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"net/url"
)

type MockRecipeModel struct{}

func (m MockRecipeModel) Insert(recipe *data.Recipe) error {
	return nil
}

func (m MockRecipeModel) GetAllForUser(userID int64, params url.Values) ([]*data.Recipe, error) {
	return []*data.Recipe{}, nil
}

func (m MockRecipeModel) Update(recipe *data.Recipe) error {
	return nil
}

func (m MockRecipeModel) Get(id int64, userId int64) (*data.Recipe, error) {
	return nil, nil
}

func (m MockRecipeModel) Delete(id int64, userID int64) error {
	return nil
}
