package mocks

import "github.com/hunttraitor/dialed-in-backend/internal/data"

type MockCoffeeModel struct{}

func (m MockCoffeeModel) GetAllForUser(userID int64) ([]*data.Coffee, error) {
	mockCoffees := []*data.Coffee{
		{
			ID:          1,
			UserID:      1,
			Name:        "Mock Coffee 1",
			Region:      "Region 1",
			Img:         "www.example.com",
			Description: "Example Description",
		},
		{
			ID:          2,
			UserID:      1,
			Name:        "Mock Coffee 2",
			Region:      "Region 2",
			Img:         "www.example.com",
			Description: "Example Description",
		},
	}
	return mockCoffees, nil
}
