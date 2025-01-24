package mocks

import (
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"time"
)

type MockMethodModel struct {
}

func (m MockMethodModel) GetAll() ([]*data.Method, error) {
	mockMethods := []*data.Method{
		{
			ID:        1,
			Name:      "Mock Method 1",
			Img:       "https://example.com/img1.png",
			CreatedAt: time.Now(),
		},
		{
			ID:        2,
			Name:      "Mock Method 2",
			Img:       "https://example.com/img2.png",
			CreatedAt: time.Now(),
		},
	}

	return mockMethods, nil
}
