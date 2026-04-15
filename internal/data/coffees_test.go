package data

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func boolPtr(v bool) *bool {
	return &v
}

func float64Ptr(v float64) *float64 {
	return &v
}

func baseCoffeeFilters() CoffeeFilters {
	return CoffeeFilters{
		Rating:       []string{},
		OriginType:   []string{},
		RoastLevel:   []string{},
		TastingNotes: []string{},
		Filters: Filters{
			Sort:         "name",
			SortSafelist: CoffeeSafeSortList,
			Page:         DEFAULT_PAGE,
			PageSize:     DEFAULT_PAGE_SIZE,
		},
	}
}

func TestGetAllCoffeeFilters(t *testing.T) {
	db := newTestDB(t)
	users := UserModel{db}
	coffees := CoffeeModel{db}

	user := &User{
		Name:  "Coffee Tester",
		Email: "coffeetester@example.com",
		Password: password{
			hash: []byte("hash"),
		},
		Activated: true,
	}
	require.NoError(t, users.Insert(user))

	otherUser := &User{
		Name:  "Other User",
		Email: "othercoffeetester@example.com",
		Password: password{
			hash: []byte("hash"),
		},
		Activated: true,
	}
	require.NoError(t, users.Insert(otherUser))

	targetCoffee := &Coffee{
		UserID: int(user.ID),
		Info: CoffeeInfo{
			Name:         "Kenya Sunrise",
			Roaster:      "Onyx",
			Region:       "Nyeri",
			Process:      "Washed",
			Variety:      "SL28",
			OriginType:   "Single Origin",
			RoastLevel:   "Light",
			Decaf:        false,
			Rating:       5,
			TastingNotes: []string{"Citrus", "Berry"},
			Cost:         22.50,
		},
	}
	require.NoError(t, coffees.Insert(targetCoffee))

	// Ensures GetAllForUser always scopes by user_id.
	require.NoError(t, coffees.Insert(&Coffee{
		UserID: int(otherUser.ID),
		Info: CoffeeInfo{
			Name:         "Should Not Appear",
			Roaster:      "Hidden Roaster",
			Region:       "Hidden Region",
			Process:      "Natural",
			Variety:      "Bourbon",
			OriginType:   "Blend",
			RoastLevel:   "Dark",
			Decaf:        true,
			Rating:       1,
			TastingNotes: []string{"Smoke"},
			Cost:         9.00,
		},
	}))

	tests := []struct {
		name      string
		mutate    func(*CoffeeFilters)
		wantNames []string
	}{
		{
			name:      "name no filter",
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "name filter match",
			mutate:    func(f *CoffeeFilters) { f.Name = "Kenya" },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "name filter partial match",
			mutate:    func(f *CoffeeFilters) { f.Name = "Sun" },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "name filter wrong",
			mutate:    func(f *CoffeeFilters) { f.Name = "Colombia" },
			wantNames: []string{},
		},
		{
			name:      "roaster no filter",
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "roaster filter match",
			mutate:    func(f *CoffeeFilters) { f.Roaster = "Onyx" },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "roaster filter partial match",
			mutate:    func(f *CoffeeFilters) { f.Roaster = "Ony" },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "roaster filter wrong",
			mutate:    func(f *CoffeeFilters) { f.Roaster = "Blue Bottle" },
			wantNames: []string{},
		},
		{
			name:      "region no filter",
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "region filter match",
			mutate:    func(f *CoffeeFilters) { f.Region = "Nyeri" },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "region filter partial match",
			mutate:    func(f *CoffeeFilters) { f.Region = "Nye" },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "region filter wrong",
			mutate:    func(f *CoffeeFilters) { f.Region = "Huila" },
			wantNames: []string{},
		},
		{
			name:      "process no filter",
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "process filter match",
			mutate:    func(f *CoffeeFilters) { f.Process = "Washed" },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "process filter wrong",
			mutate:    func(f *CoffeeFilters) { f.Process = "Anaerobic" },
			wantNames: []string{},
		},
		{
			name:      "process filter partial match",
			mutate:    func(f *CoffeeFilters) { f.Process = "W" },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "variety no filter",
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "variety filter match",
			mutate:    func(f *CoffeeFilters) { f.Variety = "SL28" },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "variety filter wrong",
			mutate:    func(f *CoffeeFilters) { f.Variety = "Caturra" },
			wantNames: []string{},
		},
		{
			name:      "variety filter partial match",
			mutate:    func(f *CoffeeFilters) { f.Variety = "SL" },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "origin_type no filter",
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "origin_type filter match",
			mutate:    func(f *CoffeeFilters) { f.OriginType = []string{"single origin"} },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "origin_type filter partial match",
			mutate:    func(f *CoffeeFilters) { f.OriginType = []string{"random origin", "single origin"} },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "origin_type filter wrong",
			mutate:    func(f *CoffeeFilters) { f.OriginType = []string{"blend"} },
			wantNames: []string{},
		},
		{
			name:      "roast_level no filter",
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "roast_level filter match",
			mutate:    func(f *CoffeeFilters) { f.RoastLevel = []string{"light"} },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "roast_level filter partial match",
			mutate:    func(f *CoffeeFilters) { f.RoastLevel = []string{"light", "medium"} },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "roast_level filter wrong",
			mutate:    func(f *CoffeeFilters) { f.RoastLevel = []string{"dark"} },
			wantNames: []string{},
		},
		{
			name:      "decaf no filter",
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "decaf filter match",
			mutate:    func(f *CoffeeFilters) { f.Decaf = boolPtr(false) },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "decaf filter wrong",
			mutate:    func(f *CoffeeFilters) { f.Decaf = boolPtr(true) },
			wantNames: []string{},
		},
		{
			name:      "rating no filter",
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "rating filter match",
			mutate:    func(f *CoffeeFilters) { f.Rating = []string{"5"} },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "rating filter partial match",
			mutate:    func(f *CoffeeFilters) { f.Rating = []string{"4", "5"} },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "rating filter wrong",
			mutate:    func(f *CoffeeFilters) { f.Rating = []string{"1"} },
			wantNames: []string{},
		},
		{
			name:      "tasting_notes no filter",
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "tasting_notes filter match",
			mutate:    func(f *CoffeeFilters) { f.TastingNotes = []string{"citrus"} },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "tasting_notes filter partial match",
			mutate:    func(f *CoffeeFilters) { f.TastingNotes = []string{"citrus", "berry"} },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "tasting_notes filter wrong",
			mutate:    func(f *CoffeeFilters) { f.TastingNotes = []string{"peanut"} },
			wantNames: []string{},
		},
		{
			name:      "min_cost no filter",
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "min_cost filter match",
			mutate:    func(f *CoffeeFilters) { f.MinCost = float64Ptr(20.00) },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "min_cost filter wrong",
			mutate:    func(f *CoffeeFilters) { f.MinCost = float64Ptr(30.00) },
			wantNames: []string{},
		},
		{
			name:      "max_cost no filter",
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "max_cost filter match",
			mutate:    func(f *CoffeeFilters) { f.MaxCost = float64Ptr(25.00) },
			wantNames: []string{"Kenya Sunrise"},
		},
		{
			name:      "max_cost filter wrong",
			mutate:    func(f *CoffeeFilters) { f.MaxCost = float64Ptr(10.00) },
			wantNames: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := baseCoffeeFilters()
			if tt.mutate != nil {
				tt.mutate(&filters)
			}

			got, _, err := coffees.GetAllForUser(user.ID, filters)
			require.NoError(t, err)

			gotNames := make([]string, 0, len(got))
			for _, coffee := range got {
				gotNames = append(gotNames, coffee.Info.Name)
			}

			require.Equal(t, tt.wantNames, gotNames)
		})
	}
}

func TestGetAllCoffeeSort(t *testing.T) {
	db := newTestDB(t)
	users := UserModel{db}
	coffees := CoffeeModel{db}

	user := &User{
		Name:  "Coffee Sort Tester",
		Email: "coffeesorttester@example.com",
		Password: password{
			hash: []byte("hash"),
		},
		Activated: true,
	}
	require.NoError(t, users.Insert(user))

	require.NoError(t, coffees.Insert(&Coffee{
		UserID: int(user.ID),
		Info: CoffeeInfo{
			Name:         "Alpha",
			Roaster:      "Roaster A",
			Region:       "Region A",
			Process:      "Process A",
			Variety:      "Variety A",
			OriginType:   "Blend",
			RoastLevel:   "Level A",
			Decaf:        false,
			Rating:       1,
			TastingNotes: []string{"a"},
			Cost:         10.00,
		},
	}))

	require.NoError(t, coffees.Insert(&Coffee{
		UserID: int(user.ID),
		Info: CoffeeInfo{
			Name:         "Bravo",
			Roaster:      "Roaster B",
			Region:       "Region B",
			Process:      "Process B",
			Variety:      "Variety B",
			OriginType:   "Microlot",
			RoastLevel:   "Level B",
			Decaf:        false,
			Rating:       2,
			TastingNotes: []string{"b"},
			Cost:         20.00,
		},
	}))

	require.NoError(t, coffees.Insert(&Coffee{
		UserID: int(user.ID),
		Info: CoffeeInfo{
			Name:         "Charlie",
			Roaster:      "Roaster C",
			Region:       "Region C",
			Process:      "Process C",
			Variety:      "Variety C",
			OriginType:   "Single Origin",
			RoastLevel:   "Level C",
			Decaf:        true,
			Rating:       3,
			TastingNotes: []string{"c"},
			Cost:         30.00,
		},
	}))

	tests := []struct {
		name      string
		mutate    func(*CoffeeFilters)
		wantNames []string
	}{
		{
			name:      "sort name asc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "name" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort name desc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "-name" },
			wantNames: []string{"Charlie", "Bravo", "Alpha"},
		},
		{
			name:      "sort roaster asc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "roaster" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort roaster desc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "-roaster" },
			wantNames: []string{"Charlie", "Bravo", "Alpha"},
		},
		{
			name:      "sort region asc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "region" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort region desc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "-region" },
			wantNames: []string{"Charlie", "Bravo", "Alpha"},
		},
		{
			name:      "sort process asc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "process" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort process desc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "-process" },
			wantNames: []string{"Charlie", "Bravo", "Alpha"},
		},
		{
			name:      "sort variety asc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "variety" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort variety desc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "-variety" },
			wantNames: []string{"Charlie", "Bravo", "Alpha"},
		},
		{
			name:      "sort origin_type asc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "origin_type" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort origin_type desc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "-origin_type" },
			wantNames: []string{"Charlie", "Bravo", "Alpha"},
		},
		{
			name:      "sort roast_level asc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "roast_level" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort roast_level desc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "-roast_level" },
			wantNames: []string{"Charlie", "Bravo", "Alpha"},
		},
		{
			name:      "sort decaf asc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "decaf" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort decaf desc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "-decaf" },
			wantNames: []string{"Charlie", "Alpha", "Bravo"},
		},
		{
			name:      "sort rating asc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "rating" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort rating desc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "-rating" },
			wantNames: []string{"Charlie", "Bravo", "Alpha"},
		},
		{
			name:      "sort tasting_notes asc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "tasting_notes" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort tasting_notes desc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "-tasting_notes" },
			wantNames: []string{"Charlie", "Bravo", "Alpha"},
		},
		{
			name:      "sort cost asc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "cost" },
			wantNames: []string{"Alpha", "Bravo", "Charlie"},
		},
		{
			name:      "sort cost desc",
			mutate:    func(f *CoffeeFilters) { f.Sort = "-cost" },
			wantNames: []string{"Charlie", "Bravo", "Alpha"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filters := baseCoffeeFilters()
			if tt.mutate != nil {
				tt.mutate(&filters)
			}

			got, _, err := coffees.GetAllForUser(user.ID, filters)
			require.NoError(t, err)

			gotNames := make([]string, 0, len(got))
			for _, coffee := range got {
				gotNames = append(gotNames, coffee.Info.Name)
			}

			require.Equal(t, tt.wantNames, gotNames)
		})
	}
}
