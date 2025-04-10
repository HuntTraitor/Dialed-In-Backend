package mocks

import (
	"github.com/hunttraitor/dialed-in-backend/internal/data"
)

type MockMethodModel struct {
}

func (m MockMethodModel) GetAll() ([]*data.Method, error) {
	mockMethods := []*data.Method{
		{
			ID:        1,
			Name:      "Mock Method 1",
			CreatedAt: "2025-01-25 00:28:23 +00:00",
		},
		{
			ID:        2,
			Name:      "Mock Method 2",
			CreatedAt: "2025-01-25 00:28:23 +00:00",
		},
	}

	return mockMethods, nil
}

func (m MockMethodModel) GetOne(id int64) (*data.Method, error) {
	mockMethod := &data.Method{
		ID:        id,
		Name:      "Mock Method 1",
		CreatedAt: "2025-01-25 00:28:23 +00:00",
	}
	return mockMethod, nil
}
