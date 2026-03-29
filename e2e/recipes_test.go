package e2e

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/hunttraitor/dialed-in-backend/e2e/testutils"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
)

func TestPostRecipes(t *testing.T) {
	app := testutils.NewTestApp(t)
	user := app.Factory.CreateUser(t)
	token := app.Factory.Login(t, user.Email, user.Password)

	coffee := app.Factory.CreateCoffee(t, token, testutils.ValidCoffeeForm())
	grinder := app.Factory.CreateGrinder(t, token, testutils.ValidGrinder())
	v60InfoJSON, _ := json.Marshal(testutils.ValidV60Info())

	commonTests := []struct {
		name   string
		mutate func(*testutils.CreateRecipeRequest)
		assert func(*httpexpect.Response)
	}{
		{
			name:   "Successfully posts recipe with all fields",
			mutate: func(req *testutils.CreateRecipeRequest) {},
			assert: func(res *httpexpect.Response) {
				recipe := res.Status(http.StatusCreated).JSON().Object().Value("recipe").Object()
				recipe.Value("id").Number().Gt(0)
				recipe.Value("user_id").Number().Equal(user.ID)
				method := recipe.Value("method").Object()
				method.Value("id").Number().Equal(float64(1))
				method.Value("name").String().NotEmpty()
				coff := recipe.Value("coffee").Object()
				coff.Value("id").Number().Equal(coffee.Coffee.ID)
				coff.Value("info").Object().Value("name").String().Equal(coffee.Coffee.Info.Name)
				recipe.Value("grinder").Object().Value("id").Number().Equal(grinder.ID)
				recipe.Value("info").Object().NotEmpty()
				recipe.Value("created_at").String().NotEmpty()
				recipe.Value("version").Number().Equal(1)
			},
		},
		{
			name: "Missing coffee and grinder is successful",
			mutate: func(req *testutils.CreateRecipeRequest) {
				req.CoffeeId = 0
				req.GrinderId = 0
			},
			assert: func(res *httpexpect.Response) {
				recipe := res.Status(http.StatusCreated).JSON().Object().Value("recipe").Object()
				recipe.Value("id").Number().Gt(0)
				recipe.Value("user_id").Number().Equal(user.ID)
				recipe.NotContainsKey("coffee")
				recipe.NotContainsKey("grinder")
			},
		},
		{
			name: "Empty body returns an error",
			mutate: func(req *testutils.CreateRecipeRequest) {
				*req = testutils.CreateRecipeRequest{}
			},
			assert: func(res *httpexpect.Response) {
				err := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
				err.Value("info").String().NotEmpty()
				err.Value("method_id").String().NotEmpty()
			},
		},
		{
			name: "Unknown method id returns an error",
			mutate: func(req *testutils.CreateRecipeRequest) {
				req.MethodId = 99999
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusNotFound).JSON().Object().
					Value("error").String().Equal("the requested method could not be found")
			},
		},
		{
			name: "Unknown coffee id returns an error",
			mutate: func(req *testutils.CreateRecipeRequest) {
				req.CoffeeId = 99999
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusNotFound).JSON().Object().
					Value("error").String().Equal("the requested coffee could not be found")
			},
		},
		{
			name: "Unknown grinder id returns an error",
			mutate: func(req *testutils.CreateRecipeRequest) {
				req.GrinderId = 99999
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusNotFound).JSON().Object().
					Value("error").String().Equal("the requested grinder could not be found")
			},
		},
	}

	for _, tt := range commonTests {
		t.Run(tt.name, func(t *testing.T) {
			req := testutils.CreateRecipeRequest{
				MethodId:  1,
				CoffeeId:  coffee.Coffee.ID,
				GrinderId: grinder.ID,
				Info:      v60InfoJSON,
			}
			tt.mutate(&req)
			b, _ := json.Marshal(req)
			t.Logf("%s", b)
			t.Log("after...")
			tt.assert(app.Client(token).POSTJSON("/v1/recipes", req).Expect(t))
		})
	}

	t.Run("Not authenticated returns an error", func(t *testing.T) {
		req := testutils.CreateRecipeRequest{MethodId: 1, Info: v60InfoJSON}
		app.Client("").POSTJSON("/v1/recipes", req).Expect(t).Status(http.StatusUnauthorized)
	})

	t.Run("ID fields as strings returns an error", func(t *testing.T) {
		body := map[string]any{"method_id": "not-an-int", "coffee_id": "abc", "grinder_id": "def"}
		app.Client(token).POSTJSON("/v1/recipes", body).Expect(t).Status(http.StatusBadRequest)
	})

	// Switch tests
	t.Run("Switch", func(t *testing.T) {
		switchTests := []struct {
			name   string
			mutate func(*data.SwitchRecipeInfo)
			assert func(*httpexpect.Response)
		}{
			{
				name:   "Successfully posts with all fields",
				mutate: func(info *data.SwitchRecipeInfo) {},
				assert: func(res *httpexpect.Response) {
					recipe := res.Status(http.StatusCreated).JSON().Object().Value("recipe").Object()
					recipe.Value("id").Number().Gt(0)
					recipe.Value("user_id").Number().Equal(user.ID)
					recipe.Value("version").Number().Equal(1)
					info := recipe.Value("info").Object()
					info.Value("name").String().Equal(testutils.ValidSwitchInfo().Name)
					info.Value("grams_in").Number().Equal(testutils.ValidSwitchInfo().GramIn)
					info.Value("ml_out").Number().Equal(testutils.ValidSwitchInfo().MlOut)
					info.Value("grind_size").String().Equal(testutils.ValidSwitchInfo().GrindSize)
					info.Value("phases").Array().Length().Equal(len(testutils.ValidSwitchInfo().Phases))
				},
			},
			{
				name: "Missing info fields returns errors",
				mutate: func(info *data.SwitchRecipeInfo) {
					*info = data.SwitchRecipeInfo{}
				},
				assert: func(res *httpexpect.Response) {
					errs := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
					errs.Value("grams_in").String().NotEmpty()
					errs.Value("ml_out").String().NotEmpty()
					errs.Value("phases").String().NotEmpty()
					errs.Value("name").String().NotEmpty()
				},
			},
			{
				name: "Empty phase fields returns errors",
				mutate: func(info *data.SwitchRecipeInfo) {
					info.Phases = []data.SwitchPhase{{}}
				},
				assert: func(res *httpexpect.Response) {
					errs := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
					errs.Value("amount").String().NotEmpty()
					errs.Value("open").String().NotEmpty()
					errs.Value("time").String().NotEmpty()
				},
			},
			{
				name: "Name too long returns an error",
				mutate: func(info *data.SwitchRecipeInfo) {
					info.Name = strings.Repeat("a", 101)
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.name").String().NotEmpty()
				},
			},
			{
				name: "Grams in too big returns an error",
				mutate: func(info *data.SwitchRecipeInfo) {
					info.GramIn = 10000
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.grams_in").String().NotEmpty()
				},
			},
			{
				name: "Grams in at 0 returns an error",
				mutate: func(info *data.SwitchRecipeInfo) {
					info.GramIn = 0
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.grams_in").String().NotEmpty()
				},
			},
			{
				name: "Ml out too big returns an error",
				mutate: func(info *data.SwitchRecipeInfo) {
					info.MlOut = 1000
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.ml_out").String().NotEmpty()
				},
			},
			{
				name: "Ml out at 0 returns an error",
				mutate: func(info *data.SwitchRecipeInfo) {
					info.MlOut = 0
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.ml_out").String().NotEmpty()
				},
			},
			{
				name: "No phases returns an error",
				mutate: func(info *data.SwitchRecipeInfo) {
					info.Phases = nil
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.phases").String().NotEmpty()
				},
			},
			{
				name: "Grind size too long returns an error",
				mutate: func(info *data.SwitchRecipeInfo) {
					info.GrindSize = strings.Repeat("a", 51)
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.grind_size").String().NotEmpty()
				},
			},
			{
				name: "Phase time 0 returns an error",
				mutate: func(info *data.SwitchRecipeInfo) {
					info.Phases = []data.SwitchPhase{
						{Open: testutils.Ptr(true), Time: testutils.Ptr(0), Amount: testutils.Ptr(50)},
					}
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.time").String().NotEmpty()
				},
			},
			{
				name: "Phase amount 0 passes",
				mutate: func(info *data.SwitchRecipeInfo) {
					info.Phases = []data.SwitchPhase{
						{Open: testutils.Ptr(true), Time: testutils.Ptr(30), Amount: testutils.Ptr(0)},
					}
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusCreated)
				},
			},
			{
				name: "Phase amount less than 0 returns an error",
				mutate: func(info *data.SwitchRecipeInfo) {
					info.Phases = []data.SwitchPhase{
						{Open: testutils.Ptr(true), Time: testutils.Ptr(30), Amount: testutils.Ptr(-1)},
					}
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.amount").String().NotEmpty()
				},
			},
			{
				name: "Phase amount greater than 1000 returns an error",
				mutate: func(info *data.SwitchRecipeInfo) {
					info.Phases = []data.SwitchPhase{
						{Open: testutils.Ptr(true), Time: testutils.Ptr(30), Amount: testutils.Ptr(1001)},
					}
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.amount").String().NotEmpty()
				},
			},
			{
				name: "Phase time greater than 10000 returns an error",
				mutate: func(info *data.SwitchRecipeInfo) {
					info.Phases = []data.SwitchPhase{
						{Open: testutils.Ptr(true), Time: testutils.Ptr(10001), Amount: testutils.Ptr(50)},
					}
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.time").String().NotEmpty()
				},
			},
		}

		for _, tt := range switchTests {
			t.Run(tt.name, func(t *testing.T) {
				info := testutils.ValidSwitchInfo()
				tt.mutate(&info)
				infoJSON, _ := json.Marshal(info)
				req := testutils.CreateRecipeRequest{MethodId: 2, Info: infoJSON}
				tt.assert(app.Client(token).POSTJSON("/v1/recipes", req).Expect(t))
			})
		}

		t.Run("Phase open not bool returns an error", func(t *testing.T) {
			infoMap := map[string]any{
				"name": "Test Switch", "grams_in": 15, "ml_out": 250,
				"phases": []map[string]any{{"open": "yes", "time": 30, "amount": 50}},
			}
			infoJSON, _ := json.Marshal(infoMap)
			res := app.Client(token).POSTJSON("/v1/recipes", testutils.CreateRecipeRequest{MethodId: 2, Info: infoJSON}).Expect(t)
			res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object().
				Value("info").String().NotEmpty()
		})
	})

	// V60 tests
	t.Run("V60", func(t *testing.T) {
		v60Tests := []struct {
			name   string
			mutate func(*data.V60RecipeInfo)
			assert func(*httpexpect.Response)
		}{
			{
				name:   "Successfully posts with all fields",
				mutate: func(info *data.V60RecipeInfo) {},
				assert: func(res *httpexpect.Response) {
					recipe := res.Status(http.StatusCreated).JSON().Object().Value("recipe").Object()
					recipe.Value("id").Number().Gt(0)
					recipe.Value("user_id").Number().Equal(user.ID)
					recipe.Value("version").Number().Equal(1)
					info := recipe.Value("info").Object()
					info.Value("name").String().Equal(testutils.ValidV60Info().Name)
					info.Value("grams_in").Number().Equal(testutils.ValidV60Info().GramIn)
					info.Value("ml_out").Number().Equal(testutils.ValidV60Info().MlOut)
					info.Value("grind_size").String().Equal(testutils.ValidV60Info().GrindSize)
					info.Value("phases").Array().Length().Equal(len(testutils.ValidV60Info().Phases))
				},
			},
			{
				name: "Missing info fields returns errors",
				mutate: func(info *data.V60RecipeInfo) {
					*info = data.V60RecipeInfo{}
				},
				assert: func(res *httpexpect.Response) {
					errs := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
					errs.Value("grams_in").String().NotEmpty()
					errs.Value("ml_out").String().NotEmpty()
					errs.Value("phases").String().NotEmpty()
					errs.Value("name").String().NotEmpty()
				},
			},
			{
				name: "Empty phase fields returns errors",
				mutate: func(info *data.V60RecipeInfo) {
					info.Phases = []data.V60Phase{{}}
				},
				assert: func(res *httpexpect.Response) {
					errs := res.Status(http.StatusUnprocessableEntity).JSON().Object().Value("error").Object()
					errs.Value("amount").String().NotEmpty()
					errs.Value("time").String().NotEmpty()
				},
			},
			{
				name: "Name too long returns an error",
				mutate: func(info *data.V60RecipeInfo) {
					info.Name = strings.Repeat("a", 101)
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.name").String().Equal("must not be more than 100 bytes")
				},
			},
			{
				name: "Grams in too big returns an error",
				mutate: func(info *data.V60RecipeInfo) {
					info.GramIn = 10000
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.grams_in").String().NotEmpty()
				},
			},
			{
				name: "Grams in at 0 returns an error",
				mutate: func(info *data.V60RecipeInfo) {
					info.GramIn = 0
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.grams_in").String().NotEmpty()
				},
			},
			{
				name: "Ml out too big returns an error",
				mutate: func(info *data.V60RecipeInfo) {
					info.MlOut = 1000
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.ml_out").String().NotEmpty()
				},
			},
			{
				name: "Ml out at 0 returns an error",
				mutate: func(info *data.V60RecipeInfo) {
					info.MlOut = 0
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.ml_out").String().NotEmpty()
				},
			},
			{
				name: "No phases returns an error",
				mutate: func(info *data.V60RecipeInfo) {
					info.Phases = nil
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.phases").String().NotEmpty()
				},
			},
			{
				name: "Grind size too long returns an error",
				mutate: func(info *data.V60RecipeInfo) {
					info.GrindSize = strings.Repeat("a", 51)
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.grind_size").String().NotEmpty()
				},
			},
			{
				name: "Phase time 0 returns an error",
				mutate: func(info *data.V60RecipeInfo) {
					info.Phases = []data.V60Phase{
						{Time: testutils.Ptr(0), Amount: testutils.Ptr(50)},
					}
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.time").String().NotEmpty()
				},
			},
			{
				name: "Phase amount 0 passes",
				mutate: func(info *data.V60RecipeInfo) {
					info.Phases = []data.V60Phase{
						{Time: testutils.Ptr(30), Amount: testutils.Ptr(0)},
					}
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusCreated)
				},
			},
			{
				name: "Phase amount less than 0 returns an error",
				mutate: func(info *data.V60RecipeInfo) {
					info.Phases = []data.V60Phase{
						{Time: testutils.Ptr(30), Amount: testutils.Ptr(-1)},
					}
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.amount").String().NotEmpty()
				},
			},
			{
				name: "Phase amount greater than 100 returns an error",
				mutate: func(info *data.V60RecipeInfo) {
					info.Phases = []data.V60Phase{
						{Time: testutils.Ptr(30), Amount: testutils.Ptr(1001)},
					}
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.amount").String().NotEmpty()
				},
			},
			{
				name: "Phase time greater than 10000 returns an error",
				mutate: func(info *data.V60RecipeInfo) {
					info.Phases = []data.V60Phase{
						{Time: testutils.Ptr(10001), Amount: testutils.Ptr(50)},
					}
				},
				assert: func(res *httpexpect.Response) {
					res.Status(http.StatusUnprocessableEntity).JSON().Object().
						Path("$.error.time").String().NotEmpty()
				},
			},
		}

		for _, tt := range v60Tests {
			t.Run(tt.name, func(t *testing.T) {
				info := testutils.ValidV60Info()
				tt.mutate(&info)
				infoJSON, _ := json.Marshal(info)
				req := testutils.CreateRecipeRequest{MethodId: 1, Info: infoJSON}
				tt.assert(app.Client(token).POSTJSON("/v1/recipes", req).Expect(t))
			})
		}
	})
}
