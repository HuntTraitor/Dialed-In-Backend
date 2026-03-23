package testutils

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

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

type CreateCoffeeResponse struct {
	Coffee struct {
		ID int64 `json:"id"`
	} `json:"coffee"`
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

func (f *FixtureFactory) CreateCoffee(t *testing.T, token string, form CoffeeForm) CreateCoffeeResponse {
	t.Helper()

	res := (&APIClient{BaseURL: f.BaseURL, Token: token}).
		POSTMultipart("/v1/coffees", form).Expect(t)

	res.Status(http.StatusCreated)

	var body CreateCoffeeResponse
	DecodeJSON(t, res, &body)

	require.NotZero(t, body.Coffee.ID)
	return body
}
