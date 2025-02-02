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
			Version:     1,
		},
		{
			ID:          2,
			UserID:      1,
			Name:        "Mock Coffee 2",
			Region:      "Region 2",
			Img:         "www.example.com",
			Description: "Example Description",
			Version:     1,
		},
	}
	return mockCoffees, nil
}

func (m MockCoffeeModel) Insert(userID int64, coffee *data.Coffee) (*data.Coffee, error) {
	return &data.Coffee{
		ID:          1,
		UserID:      2,
		Name:        "Inserted Coffee",
		Region:      "Inserted Region",
		CreatedAt:   "Inserted Created At",
		Description: "Inserted description",
		Version:     1,
	}, nil
}

func (m MockCoffeeModel) GetOne(id int64, userId int64) (*data.Coffee, error) {
	return &data.Coffee{
		ID:          1,
		UserID:      2,
		Name:        "Mock Coffee",
		Region:      "Mock Region",
		CreatedAt:   "Mock Created At",
		Description: "Mock description",
		Version:     1,
	}, nil
}

func (m MockCoffeeModel) Update(coffee *data.Coffee) error {
	return nil
}
