package mocks

import "github.com/hunttraitor/dialed-in-backend/internal/data"

type MockCoffeeModel struct{}

var MockCoffee = &data.Coffee{
	ID:        1,
	UserID:    1,
	Version:   1,
	CreatedAt: "2025-01-01T00:00:00Z",
	Info: data.CoffeeInfo{
		Name:         "Mock Coffee",
		Roaster:      "Mock Roaster",
		Region:       "Mock Region",
		Process:      "Washed",
		Decaf:        false,
		OriginType:   "Single Origin",
		TastingNotes: []string{"floral", "citrus"},
		Rating:       5,
		RoastLevel:   "Medium-Light",
		Cost:         13.75,
		Img:          "https://example.com/mock-coffee.png",
		Description:  "A test mock coffee used for unit testing.",
	},
}

var MockEmptyCoffee = data.Coffee{
	ID:        1,
	UserID:    1,
	Version:   1,
	CreatedAt: "2025-01-01T00:00:00Z",
	Info: data.CoffeeInfo{
		Name:  "Mock Empty Coffee",
		Decaf: false,
	},
}

func (m MockCoffeeModel) GetAllForUser(userID int64) ([]*data.Coffee, error) {
	mockCoffees := []*data.Coffee{
		{
			ID:        1,
			UserID:    1,
			Version:   1,
			CreatedAt: "2025-01-01T00:00:00Z",
			Info: data.CoffeeInfo{
				Name:         "Mock Coffee 1",
				Region:       "Region 1",
				Img:          "www.example.com",
				Description:  "Example Description",
				Decaf:        false,
				OriginType:   "Single Origin",
				TastingNotes: []string{"chocolate", "berry"},
				Rating:       4,
				RoastLevel:   "Medium",
				Cost:         12.99,
				Process:      "Washed",
			},
		},
		{
			ID:        2,
			UserID:    1,
			Version:   1,
			CreatedAt: "2025-01-01T00:00:00Z",
			Info: data.CoffeeInfo{
				Name:         "Mock Coffee 2",
				Region:       "Region 2",
				Img:          "www.example.com",
				Description:  "Example Description",
				Decaf:        true,
				OriginType:   "Blend",
				TastingNotes: []string{"nutty", "caramel"},
				Rating:       5,
				RoastLevel:   "Dark",
				Cost:         14.50,
				Process:      "Natural",
			},
		},
	}
	return mockCoffees, nil
}

func (m MockCoffeeModel) Insert(coffee *data.Coffee) error {
	return nil
}

func (m MockCoffeeModel) GetOne(id int64, userId int64) (*data.Coffee, error) {
	return MockCoffee, nil
}

func (m MockCoffeeModel) Update(coffee *data.Coffee) error {
	return nil
}

func (m MockCoffeeModel) Delete(id int64, userId int64) error {
	return nil
}
