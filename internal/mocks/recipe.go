package mocks

import "github.com/hunttraitor/dialed-in-backend/internal/data"

type MockRecipeModel struct{}

func (m MockRecipeModel) Insert(recipe *data.Recipe) error {
	return nil
}

func (m MockRecipeModel) GetAllForUser(userID int64, methodID int64, coffeeID int64) ([]*data.Recipe, error) {
	return []*data.Recipe{}, nil
}
