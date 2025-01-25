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
			Img:       "https://example.com/img1.png",
			CreatedAt: "2025-01-25 00:28:23 +00:00",
		},
		{
			ID:        2,
			Name:      "Mock Method 2",
			Img:       "https://example.com/img2.png",
			CreatedAt: "2025-01-25 00:28:23 +00:00",
		},
	}

	return mockMethods, nil
}
