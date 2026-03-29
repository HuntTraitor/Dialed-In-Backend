package testutils

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/stretchr/testify/require"
)

type CreateRecipeResponse struct {
	Recipe struct {
		ID        int64           `json:"id"`
		UserID    int64           `json:"user_id"`
		Method    Method          `json:"method"`
		Coffee    data.Coffee     `json:"coffee,omitempty"`
		Grinder   data.Grinder    `json:"grinder,omitempty"`
		Info      json.RawMessage `json:"info"`
		CreatedAt string          `json:"created_at"`
		Version   int             `json:"version"`
	} `json:"recipe"`
}

type CreateRecipeRequest struct {
	MethodId  int64           `json:"method_id,omitempty"`
	CoffeeId  int             `json:"coffee_id,omitempty"`
	GrinderId int64           `json:"grinder_id,omitempty"`
	Info      json.RawMessage `json:"info,omitempty"`
}

func (f *FixtureFactory) CreateRecipe(t *testing.T, token string, methodId int64, r CreateRecipeRequest) CreateRecipeResponse {
	t.Helper()

	grinder := f.CreateGrinder(t, token, ValidGrinder())
	coffee := f.CreateCoffee(t, token, ValidCoffeeForm())

	request := CreateRecipeRequest{
		GrinderId: grinder.ID,
		CoffeeId:  coffee.Coffee.ID,
		MethodId:  methodId,
		Info:      r.Info,
	}

	res := (&APIClient{BaseURL: f.BaseURL, Token: token}).
		POSTJSON("/v1/recipes", request).Expect(t)

	res.Status(http.StatusCreated)

	var body CreateRecipeResponse
	DecodeJSON(t, res, &body)

	require.NotZero(t, body.Recipe.ID)
	return body
}

func ValidSwitchInfo() data.SwitchRecipeInfo {
	return data.SwitchRecipeInfo{
		Name:      "Test Switch Recipe",
		GramIn:    15,
		MlOut:     250,
		GrindSize: "Medium",
		WaterTemp: "93°C",
		Phases: []data.SwitchPhase{
			{Open: Ptr(true), Time: Ptr(30), Amount: Ptr(50)},
		},
	}
}

func ValidV60Info() data.V60RecipeInfo {
	return data.V60RecipeInfo{
		Name:      "Test V60 Recipe",
		GramIn:    15,
		MlOut:     250,
		GrindSize: "Medium",
		WaterTemp: "93°C",
		Phases: []data.V60Phase{
			{Time: Ptr(30), Amount: Ptr(50)},
		},
	}
}

func ValidSwitchRecipe() CreateRecipeRequest {
	info := ValidSwitchInfo()
	infoJSON, _ := json.Marshal(info)
	return CreateRecipeRequest{Info: infoJSON}
}

func ValidV60Recipe() CreateRecipeRequest {
	info := ValidV60Info()
	infoJSON, _ := json.Marshal(info)
	return CreateRecipeRequest{Info: infoJSON}
}
