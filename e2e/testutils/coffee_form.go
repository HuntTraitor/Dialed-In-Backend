package testutils

type CoffeeForm struct {
	Name         string
	Roaster      string
	Region       string
	Process      string
	Description  string
	OriginType   string
	TastingNotes []string
	Rating       int
	RoastLevel   string
	Cost         float64
	Decaf        bool
	Img          []byte
	Variety      string
}

func ValidCoffeeForm() CoffeeForm {
	return CoffeeForm{
		Name:         "Test Coffee",
		Roaster:      "Test Roaster",
		Region:       "Test Region",
		Process:      "Test Process",
		Description:  "Test Description",
		OriginType:   "Test Origin Type",
		TastingNotes: []string{"Test Tasting Note 1", "Test Tasting Note 2"},
		Rating:       5,
		RoastLevel:   "Medium",
		Cost:         25.99,
		Decaf:        false,
		Img:          []byte("Test Image"),
		Variety:      "Test Variety",
	}
}

func MinimalCoffeeForm(name string) CoffeeForm {
	return CoffeeForm{
		Name: name,
	}
}

func EmptyCoffeeForm() CoffeeForm {
	return CoffeeForm{}
}
