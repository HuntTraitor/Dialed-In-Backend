package mocks

import "github.com/hunttraitor/dialed-in-backend/internal/data"

type MockRecipeModel struct{}

func (m MockRecipeModel) Insert(recipe *data.Recipe) error {
	return nil
}
