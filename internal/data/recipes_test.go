package data

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func int64Ptr(v int64) *int64 {
	return &v
}

func baseRecipeFilters() RecipeFilters {
	return RecipeFilters{
		Filters: Filters{
			Sort:         "name",
			SortSafelist: RecipeSafeSortList,
			Page:         DEFAULT_PAGE,
			PageSize:     DEFAULT_PAGE_SIZE,
		},
	}
}

func recipeInfo(t *testing.T, name string, gramsIn int, mlOut int, waterTemp string, grindSize string, phaseTime int) json.RawMessage {
	t.Helper()

	info, err := json.Marshal(map[string]any{
		"name":       name,
		"grams_in":   gramsIn,
		"ml_out":     mlOut,
		"water_temp": waterTemp,
		"grind_size": grindSize,
		"phases": []map[string]any{
			{
				"time":   phaseTime,
				"amount": mlOut,
			},
		},
	})
	require.NoError(t, err)

	return info
}

func recipeName(t *testing.T, info json.RawMessage) string {
	t.Helper()

	var decoded struct {
		Name string `json:"name"`
	}
	require.NoError(t, json.Unmarshal(info, &decoded))

	return decoded.Name
}

func TestGetAllRecipeFilters(t *testing.T) {
	db := newTestDB(t)
	users := UserModel{db}
	coffees := CoffeeModel{db}
	recipes := RecipeModel{db}

	user := &User{
		Name:  "Recipe Tester",
		Email: "recipetester@example.com",
		Password: password{
			hash: []byte("password"),
		},
		Activated: true,
	}
	require.NoError(t, users.Insert(user))

	otherUser := &User{
		Name:  "Other Recipe User",
		Email: "otherrecipetester@example.com",
		Password: password{
			hash: []byte("password"),
		},
		Activated: true,
	}
	require.NoError(t, users.Insert(otherUser))

	targetCoffee := &Coffee{
		UserID: int(user.ID),
		Info: CoffeeInfo{
			Name: "Recipe Coffee",
		},
	}
	require.NoError(t, coffees.Insert(targetCoffee))

	otherCoffee := &Coffee{
		UserID: int(user.ID),
		Info: CoffeeInfo{
			Name: "Other Recipe Coffee",
		},
	}
	require.NoError(t, coffees.Insert(otherCoffee))

	targetRecipe := &Recipe{
		UserID:   user.ID,
		MethodID: 1,
		CoffeeID: int64Ptr(int64(targetCoffee.ID)),
		Info:     recipeInfo(t, "Morning V60", 18, 300, "96°C", "Medium Fine", 45),
	}
	require.NoError(t, recipes.Insert(targetRecipe))

	// Ensures GetAllForUser always scopes by user_id.
	require.NoError(t, recipes.Insert(&Recipe{
		UserID:   otherUser.ID,
		MethodID: 1,
		CoffeeID: int64Ptr(int64(targetCoffee.ID)),
		Info:     recipeInfo(t, "Should Not Appear", 18, 300, "96°C", "Medium Fine", 45),
	}))

	tests := []struct {
		name      string
		mutate    func(*RecipeFilters)
		wantNames []string
	}{
		{
			name:      "method_id no filter",
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "method_id filter match",
			mutate:    func(f *RecipeFilters) { f.MethodID = 1 },
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "method_id filter wrong",
			mutate:    func(f *RecipeFilters) { f.MethodID = 2 },
			wantNames: []string{},
		},
		{
			name:      "coffee_id no filter",
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "coffee_id filter match",
			mutate:    func(f *RecipeFilters) { f.CoffeeID = int(targetCoffee.ID) },
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "coffee_id filter wrong",
			mutate:    func(f *RecipeFilters) { f.CoffeeID = int(otherCoffee.ID) },
			wantNames: []string{},
		},
		{
			name:      "grinder_id no filter",
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "grinder_id filter wrong",
			mutate:    func(f *RecipeFilters) { f.GrinderID = 1 },
			wantNames: []string{},
		},
		{
			name:      "name no filter",
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "name filter match",
			mutate:    func(f *RecipeFilters) { f.Name = "Morning" },
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "name filter partial match",
			mutate:    func(f *RecipeFilters) { f.Name = "V6" },
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "name filter wrong",
			mutate:    func(f *RecipeFilters) { f.Name = "Evening" },
			wantNames: []string{},
		},
		{
			name:      "search filter by name",
			mutate:    func(f *RecipeFilters) { f.Search = "Morning" },
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "search filter wrong",
			mutate:    func(f *RecipeFilters) { f.Search = "Evening" },
			wantNames: []string{},
		},
		{
			name:      "water_temp no filter",
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "water_temp filter match",
			mutate:    func(f *RecipeFilters) { f.WaterTemp = "96°C" },
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "water_temp filter partial match",
			mutate:    func(f *RecipeFilters) { f.WaterTemp = "96" },
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "water_temp filter wrong",
			mutate:    func(f *RecipeFilters) { f.WaterTemp = "88°C" },
			wantNames: []string{},
		},
		{
			name:      "grind_size no filter",
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "grind_size filter match",
			mutate:    func(f *RecipeFilters) { f.GrindSize = "Medium Fine" },
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "grind_size filter partial match",
			mutate:    func(f *RecipeFilters) { f.GrindSize = "Medium" },
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "grind_size filter wrong",
			mutate:    func(f *RecipeFilters) { f.GrindSize = "Coarse" },
			wantNames: []string{},
		},
		{
			name:      "grams_in no filter",
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "grams_in filter match",
			mutate:    func(f *RecipeFilters) { f.GramsIn = 18 },
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "grams_in filter wrong",
			mutate:    func(f *RecipeFilters) { f.GramsIn = 20 },
			wantNames: []string{},
		},
		{
			name:      "ml_out no filter",
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "ml_out filter match",
			mutate:    func(f *RecipeFilters) { f.MlOut = 300 },
			wantNames: []string{"Morning V60"},
		},
		{
			name:      "ml_out filter wrong",
			mutate:    func(f *RecipeFilters) { f.MlOut = 250 },
			wantNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := baseRecipeFilters()
			if tt.mutate != nil {
				tt.mutate(&filters)
			}

			got, _, err := recipes.GetAllForUser(user.ID, filters)
			require.NoError(t, err)

			gotNames := make([]string, 0, len(got))
			for _, recipe := range got {
				gotNames = append(gotNames, recipeName(t, recipe.Info))
			}

			require.Equal(t, tt.wantNames, gotNames)
		})
	}
}

func TestGetAllRecipeSort(t *testing.T) {
	db := newTestDB(t)
	users := UserModel{db}
	recipes := RecipeModel{db}

	user := &User{
		Name:  "Recipe Sort Tester",
		Email: "recipesorttester@example.com",
		Password: password{
			hash: []byte("hash"),
		},
		Activated: true,
	}
	require.NoError(t, users.Insert(user))

	require.NoError(t, recipes.Insert(&Recipe{
		UserID:   user.ID,
		MethodID: 1,
		Info:     recipeInfo(t, "Alpha", 10, 200, "90°C", "Fine", 10),
	}))

	require.NoError(t, recipes.Insert(&Recipe{
		UserID:   user.ID,
		MethodID: 1,
		Info:     recipeInfo(t, "Bravo", 20, 300, "95°C", "Medium", 20),
	}))

	require.NoError(t, recipes.Insert(&Recipe{
		UserID:   user.ID,
		MethodID: 1,
		Info:     recipeInfo(t, "Charlie", 30, 400, "99°C", "Coarse", 30),
	}))

	tests := []struct {
		name      string
		mutate    func(*RecipeFilters)
		wantNames []string
	}{
		{
			name:      "sort name asc",
			mutate:    func(f *RecipeFilters) { f.Sort = "name" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort name desc",
			mutate:    func(f *RecipeFilters) { f.Sort = "-name" },
			wantNames: []string{"Charlie", "Bravo", "Alpha"},
		},
		{
			name:      "sort grams_in asc",
			mutate:    func(f *RecipeFilters) { f.Sort = "grams_in" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort grams_in desc",
			mutate:    func(f *RecipeFilters) { f.Sort = "-grams_in" },
			wantNames: []string{"Charlie", "Bravo", "Alpha"},
		},
		{
			name:      "sort ml_out asc",
			mutate:    func(f *RecipeFilters) { f.Sort = "ml_out" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort ml_out desc",
			mutate:    func(f *RecipeFilters) { f.Sort = "-ml_out" },
			wantNames: []string{"Charlie", "Bravo", "Alpha"},
		},
		{
			name:      "sort water_temp asc",
			mutate:    func(f *RecipeFilters) { f.Sort = "water_temp" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort water_temp desc",
			mutate:    func(f *RecipeFilters) { f.Sort = "-water_temp" },
			wantNames: []string{"Charlie", "Bravo", "Alpha"},
		},
		{
			name:      "sort grind_size asc",
			mutate:    func(f *RecipeFilters) { f.Sort = "grind_size" },
			wantNames: []string{"Charlie", "Alpha", "Bravo"},
		},
		{
			name:      "sort grind_size desc",
			mutate:    func(f *RecipeFilters) { f.Sort = "-grind_size" },
			wantNames: []string{"Bravo", "Alpha", "Charlie"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := baseRecipeFilters()
			if tt.mutate != nil {
				tt.mutate(&filters)
			}

			got, _, err := recipes.GetAllForUser(user.ID, filters)
			require.NoError(t, err)

			gotNames := make([]string, 0, len(got))
			for _, recipe := range got {
				gotNames = append(gotNames, recipeName(t, recipe.Info))
			}

			require.Equal(t, tt.wantNames, gotNames)
		})
	}
}
