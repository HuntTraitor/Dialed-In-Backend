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
