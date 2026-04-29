package e2e

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/hunttraitor/dialed-in-backend/e2e/testutils"
	"github.com/hunttraitor/dialed-in-backend/internal/data"
	"github.com/stretchr/testify/assert"
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
				name: "Successfully posts without grinder_size field",
				mutate: func(info *data.SwitchRecipeInfo) {
					info.GrindSize = ""
				},
				assert: func(res *httpexpect.Response) {
					recipe := res.Status(http.StatusCreated).JSON().Object().Value("recipe").Object()
					recipe.Value("id").Number().Gt(0)
					recipe.Value("user_id").Number().Equal(user.ID)
					recipe.Value("version").Number().Equal(1)
					info := recipe.Value("info").Object()
					info.Value("name").String().Equal(testutils.ValidSwitchInfo().Name)
					info.NotContainsKey("grind_size")
					info.Value("grams_in").Number().Equal(testutils.ValidSwitchInfo().GramIn)
					info.Value("ml_out").Number().Equal(testutils.ValidSwitchInfo().MlOut)
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
					errs.Value("water_temp").String().NotEmpty()
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
				name: "Successfully posts without grinder_size field",
				mutate: func(info *data.V60RecipeInfo) {
					info.GrindSize = ""
				},
				assert: func(res *httpexpect.Response) {
					recipe := res.Status(http.StatusCreated).JSON().Object().Value("recipe").Object()
					recipe.Value("id").Number().Gt(0)
					recipe.Value("user_id").Number().Equal(user.ID)
					recipe.Value("version").Number().Equal(1)
					info := recipe.Value("info").Object()
					info.Value("name").String().Equal(testutils.ValidV60Info().Name)
					info.NotContainsKey("grind_size")
					info.Value("grams_in").Number().Equal(testutils.ValidV60Info().GramIn)
					info.Value("ml_out").Number().Equal(testutils.ValidV60Info().MlOut)
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
					errs.Value("water_temp").String().NotEmpty()
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

func TestGetAllRecipes(t *testing.T) {
	app := testutils.NewTestApp(t)

	t.Run("GET all recipes returns both v60 and switch recipes", func(t *testing.T) {
		user := app.Factory.CreateUser(t)
		token := app.Factory.Login(t, user.Email, user.Password)
		cof := app.Factory.CreateCoffee(t, token, testutils.ValidCoffeeForm())
		gr := app.Factory.CreateGrinder(t, token, testutils.ValidGrinder())

		v60Insert := &testutils.CreateRecipeRequest{}
		*v60Insert = testutils.ValidV60Recipe()
		v60Insert.CoffeeId = cof.Coffee.ID
		v60Insert.GrinderId = gr.ID

		switchInsert := &testutils.CreateRecipeRequest{}
		*switchInsert = testutils.ValidSwitchRecipe()
		switchInsert.CoffeeId = cof.Coffee.ID
		switchInsert.GrinderId = gr.ID

		v60 := app.Factory.CreateRecipe(t, token, 1, *v60Insert)
		sw := app.Factory.CreateRecipe(t, token, 2, *switchInsert)

		arr := app.Client(token).GET("/v1/recipes").Expect(t).
			Status(http.StatusOK).JSON().Object().Value("recipes").Array()

		arr.Length().Equal(2)

		byMethodName := make(map[string]*httpexpect.Object, 2)
		for _, v := range arr.Iter() {
			obj := v.Object()
			name := obj.Value("method").Object().Value("name").String().Raw()
			byMethodName[name] = obj
		}

		assert.ElementsMatch(t, []string{"V60", "Hario Switch"}, []string{
			byMethodName["V60"].Value("method").Object().Value("name").String().Raw(),
			byMethodName["Hario Switch"].Value("method").Object().Value("name").String().Raw(),
		})

		for _, tc := range []struct {
			recipe     testutils.CreateRecipeResponse
			obj        *httpexpect.Object
			assertInfo func(*httpexpect.Object)
		}{
			{
				recipe: v60,
				obj:    byMethodName["V60"],
				assertInfo: func(info *httpexpect.Object) {
					vi := testutils.ValidV60Info()
					info.Value("name").String().Equal(vi.Name)
					info.Value("grams_in").Number().Equal(vi.GramIn)
					info.Value("ml_out").Number().Equal(vi.MlOut)
					info.Value("grind_size").String().Equal(vi.GrindSize)
					info.Value("water_temp").String().Equal(vi.WaterTemp)
					phases := info.Value("phases").Array()
					phases.Length().Equal(1)
					phase := phases.Element(0).Object()
					phase.Value("time").Number().Equal(*vi.Phases[0].Time)
					phase.Value("amount").Number().Equal(*vi.Phases[0].Amount)
				},
			},
			{
				recipe: sw,
				obj:    byMethodName["Hario Switch"],
				assertInfo: func(info *httpexpect.Object) {
					si := testutils.ValidSwitchInfo()
					info.Value("name").String().Equal(si.Name)
					info.Value("grams_in").Number().Equal(si.GramIn)
					info.Value("ml_out").Number().Equal(si.MlOut)
					info.Value("grind_size").String().Equal(si.GrindSize)
					info.Value("water_temp").String().Equal(si.WaterTemp)
					phases := info.Value("phases").Array()
					phases.Length().Equal(1)
					phase := phases.Element(0).Object()
					phase.Value("open").Boolean().Equal(*si.Phases[0].Open)
					phase.Value("time").Number().Equal(*si.Phases[0].Time)
					phase.Value("amount").Number().Equal(*si.Phases[0].Amount)
				},
			},
		} {
			obj := tc.obj
			obj.Value("id").Number().Equal(tc.recipe.Recipe.ID)
			obj.Value("user_id").Number().Equal(user.ID)
			obj.Value("method").Object().Value("id").Number().Gt(0)
			obj.Value("method").Object().Value("name").String().NotEmpty()
			obj.Value("coffee").Object().Value("id").Number().Equal(tc.recipe.Recipe.Coffee.ID)
			obj.Value("coffee").Object().Value("info").Object().Value("name").String().Equal(tc.recipe.Recipe.Coffee.Info.Name)
			obj.Value("grinder").Object().Value("id").Number().Equal(tc.recipe.Recipe.Grinder.ID)
			tc.assertInfo(obj.Value("info").Object())
			obj.Value("created_at").String().NotEmpty()
			obj.Value("version").Number().Equal(1)
		}
	})

	t.Run("GET all recipes without coffees or grinders is successful", func(t *testing.T) {
		user := app.Factory.CreateUser(t)
		token := app.Factory.Login(t, user.Email, user.Password)

		v60InfoJSON, _ := json.Marshal(testutils.ValidV60Info())
		req := testutils.CreateRecipeRequest{MethodId: 1, Info: v60InfoJSON}
		app.Client(token).POSTJSON("/v1/recipes", req).Expect(t).Status(http.StatusCreated)

		arr := app.Client(token).GET("/v1/recipes").Expect(t).
			Status(http.StatusOK).JSON().Object().Value("recipes").Array()

		arr.Length().Equal(1)
		recipe := arr.Element(0).Object()
		recipe.NotContainsKey("coffee")
		recipe.NotContainsKey("grinder")
	})

	t.Run("GET all recipes unauthorized returns error", func(t *testing.T) {
		app.Client("").GET("/v1/recipes").Expect(t).Status(http.StatusUnauthorized)
	})

	t.Run("GET all recipes no recipes returns an empty array", func(t *testing.T) {
		user := app.Factory.CreateUser(t)
		token := app.Factory.Login(t, user.Email, user.Password)

		app.Client(token).GET("/v1/recipes").Expect(t).
			Status(http.StatusOK).JSON().Object().Value("recipes").Array().Empty()
	})

	t.Run("Successfully applies all recipe filters in one query", func(t *testing.T) {
		user := app.Factory.CreateUser(t)
		token := app.Factory.Login(t, user.Email, user.Password)

		coffee := app.Factory.CreateCoffee(t, token, testutils.ValidCoffeeForm())
		grinder := app.Factory.CreateGrinder(t, token, testutils.ValidGrinder())

		targetInfo := testutils.ValidV60Info()
		targetInfo.Name = "FilterTarget Morning V60"
		targetInfo.WaterTemp = "96°C"
		targetInfo.GrindSize = "Medium Fine"
		targetInfo.GramIn = 18
		targetInfo.MlOut = 300
		targetInfoJSON, _ := json.Marshal(targetInfo)

		target := app.Factory.CreateRecipe(t, token, 1, testutils.CreateRecipeRequest{
			MethodId:  1,
			CoffeeId:  coffee.Coffee.ID,
			GrinderId: grinder.ID,
			Info:      targetInfoJSON,
		})

		decoyInfo := testutils.ValidV60Info()
		decoyInfo.Name = "FilterTarget Morning V60 Decoy"
		decoyInfo.WaterTemp = "96°C"
		decoyInfo.GrindSize = "Medium Fine"
		decoyInfo.GramIn = 19
		decoyInfo.MlOut = 300
		decoyInfoJSON, _ := json.Marshal(decoyInfo)
		app.Factory.CreateRecipe(t, token, 1, testutils.CreateRecipeRequest{
			MethodId:  1,
			CoffeeId:  coffee.Coffee.ID,
			GrinderId: grinder.ID,
			Info:      decoyInfoJSON,
		})

		res := app.Client(token).
			GET("/v1/recipes").
			WithQuery("method_id", "1").
			WithQuery("coffee_id", fmt.Sprintf("%d", coffee.Coffee.ID)).
			WithQuery("grinder_id", fmt.Sprintf("%d", grinder.ID)).
			WithQuery("search", "FilterTarget Morning").
			WithQuery("name", "FilterTarget Morn").
			WithQuery("water_temp", "96").
			WithQuery("grind_size", "Medium Fi").
			WithQuery("grams_in", "18").
			WithQuery("ml_out", "300").
			Expect(t).
			Status(http.StatusOK)

		recipes := res.JSON().Object().Value("recipes").Array()
		recipes.Length().Equal(1)

		recipe := recipes.Element(0).Object()
		recipe.Value("id").Number().Equal(target.Recipe.ID)
		info := recipe.Value("info").Object()
		info.Value("name").String().Equal(targetInfo.Name)
		info.Value("water_temp").String().Equal(targetInfo.WaterTemp)
		info.Value("grind_size").String().Equal(targetInfo.GrindSize)
		info.Value("grams_in").Number().Equal(float64(targetInfo.GramIn))
		info.Value("ml_out").Number().Equal(float64(targetInfo.MlOut))
	})
}

func TestGetAllRecipesSort(t *testing.T) {
	t.Run("Successfully sorts recipes by sort value", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		user := app.Factory.CreateUser(t)
		token := app.Factory.Login(t, user.Email, user.Password)

		alphaInfo := testutils.ValidV60Info()
		alphaInfo.Name = "SortCase Alpha"
		alphaJSON, _ := json.Marshal(alphaInfo)
		app.Factory.CreateRecipe(t, token, 1, testutils.CreateRecipeRequest{
			MethodId: 1,
			Info:     alphaJSON,
		})

		bravoInfo := testutils.ValidV60Info()
		bravoInfo.Name = "SortCase Bravo"
		bravoJSON, _ := json.Marshal(bravoInfo)
		app.Factory.CreateRecipe(t, token, 1, testutils.CreateRecipeRequest{
			MethodId: 1,
			Info:     bravoJSON,
		})

		res := app.Client(token).
			GET("/v1/recipes").
			WithQuery("name", "SortCase").
			WithQuery("sort", "-name").
			Expect(t).
			Status(http.StatusOK)

		recipes := res.JSON().Object().Value("recipes").Array()
		recipes.Length().Equal(2)
		recipes.Element(0).Object().Path("$.info.name").String().Equal("SortCase Bravo")
		recipes.Element(1).Object().Path("$.info.name").String().Equal("SortCase Alpha")
	})

	t.Run("Fails when sort value is invalid", func(t *testing.T) {
		app := testutils.NewTestApp(t)
		user := app.Factory.CreateUser(t)
		token := app.Factory.Login(t, user.Email, user.Password)

		app.Client(token).
			GET("/v1/recipes").
			WithQuery("sort", "not-a-real-sort").
			Expect(t).
			Status(http.StatusUnprocessableEntity).
			JSON().Object().
			Path("$.error.sort").String().Equal("invalid sort value")
	})
}

func TestGetAllRecipesPagination(t *testing.T) {
	app := testutils.NewTestApp(t)
	user := app.Factory.CreateUser(t)
	token := app.Factory.Login(t, user.Email, user.Password)
	for i := 0; i < 11; i++ {
		app.Factory.CreateRecipe(t, token, 1, testutils.ValidV60Recipe())
	}

	tests := []struct {
		name   string
		mutate func(*testutils.RequestBuilder)
		assert func(*httpexpect.Response)
	}{
		{
			name: "No pagination counter defaults to 10 pages",
			mutate: func(req *testutils.RequestBuilder) {
			},
			assert: func(res *httpexpect.Response) {
				obj := res.Status(http.StatusOK).JSON().Object()
				obj.Value("recipes").Array().Length().Equal(10)
				metadata := obj.Value("metadata").Object()
				metadata.Value("current_page").Number().Equal(1)
				metadata.Value("page_size").Number().Equal(10)
				metadata.Value("first_page").Number().Equal(1)
				metadata.Value("last_page").Number().Equal(2)
				metadata.Value("total_records").Number().Equal(11)
			},
		},
		{
			name: "Pagination set to 11 pages",
			mutate: func(req *testutils.RequestBuilder) {
				req.WithQuery("page_size", "11")
			},
			assert: func(res *httpexpect.Response) {
				obj := res.Status(http.StatusOK).JSON().Object()
				obj.Value("recipes").Array().Length().Equal(11)
				metadata := obj.Value("metadata").Object()
				metadata.Value("current_page").Number().Equal(1)
				metadata.Value("page_size").Number().Equal(11)
				metadata.Value("first_page").Number().Equal(1)
				metadata.Value("last_page").Number().Equal(1)
				metadata.Value("total_records").Number().Equal(11)
			},
		},
		{
			name: "Pagination page number correctly gives values",
			mutate: func(req *testutils.RequestBuilder) {
				req.WithQuery("page_size", "10")
				req.WithQuery("page", "2")
			},
			assert: func(res *httpexpect.Response) {
				obj := res.Status(http.StatusOK).JSON().Object()
				obj.Value("recipes").Array().Length().Equal(1)
				metadata := obj.Value("metadata").Object()
				metadata.Value("current_page").Number().Equal(2)
				metadata.Value("page_size").Number().Equal(10)
				metadata.Value("first_page").Number().Equal(1)
				metadata.Value("last_page").Number().Equal(2)
				metadata.Value("total_records").Number().Equal(11)
			},
		},
		{
			name: "Pagination page number too high returns empty values",
			mutate: func(req *testutils.RequestBuilder) {
				req.WithQuery("page", "10000")
			},
			assert: func(res *httpexpect.Response) {
				obj := res.Status(http.StatusOK).JSON().Object()
				obj.Value("recipes").Array().Length().Equal(0)
				obj.Value("metadata").Object().Empty()
			},
		},
		{
			name: "Pagination page number negative returns error",
			mutate: func(req *testutils.RequestBuilder) {
				req.WithQuery("page", "-1")
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path("$.error.page").String().
					Contains("greater than zero")
			},
		},
		{
			name: "Pagination page size too high returns error",
			mutate: func(req *testutils.RequestBuilder) {
				req.WithQuery("page_size", "10000")
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path("$.error.page_size").String().
					Contains("maximum")
			},
		},
		{
			name: "Pagination page size too low returns error",
			mutate: func(req *testutils.RequestBuilder) {
				req.WithQuery("page_size", "-1")
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path("$.error.page_size").String().
					Contains("maximum")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := app.Client(token).GET("/v1/recipes")
			tt.mutate(req)
			res := req.Expect(t)
			tt.assert(res)
		})
	}
}

func TestPatchRecipe(t *testing.T) {
	app := testutils.NewTestApp(t)
	user := app.Factory.CreateUser(t)
	token := app.Factory.Login(t, user.Email, user.Password)

	coffee := app.Factory.CreateCoffee(t, token, testutils.ValidCoffeeForm())
	grinder := app.Factory.CreateGrinder(t, token, testutils.ValidGrinder())

	tests := []struct {
		name           string
		sourceMethodID int64
		mutate         func(*testutils.PatchRecipeRequest)
		assert         func(*httpexpect.Response)
	}{
		{
			name:           "Successfully editing info name",
			sourceMethodID: 1,
			mutate: func(req *testutils.PatchRecipeRequest) {
				req.Info.Name = "Updated Name"
			},
			assert: func(res *httpexpect.Response) {
				r := res.Status(http.StatusOK).JSON().Object().Value("recipe").Object()
				r.Value("info").Object().Value("name").String().Equal("Updated Name")
				r.Value("version").Number().Equal(2)
			},
		},
		{
			name:           "Successfully changing coffee and grinder id",
			sourceMethodID: 1,
			mutate: func(req *testutils.PatchRecipeRequest) {
				req.CoffeeID = testutils.Ptr(coffee.Coffee.ID)
				req.GrinderID = testutils.Ptr(grinder.ID)
			},
			assert: func(res *httpexpect.Response) {
				r := res.Status(http.StatusOK).JSON().Object().Value("recipe").Object()
				r.Value("coffee").Object().Value("id").Number().Equal(coffee.Coffee.ID)
				r.Value("grinder").Object().Value("id").Number().Equal(grinder.ID)
				r.Value("version").Number().Equal(2)
			},
		},
		{
			name:           "Successfully modifying phases of v60",
			sourceMethodID: 1,
			mutate: func(req *testutils.PatchRecipeRequest) {
				req.Info.Phases = []testutils.PatchPhase{
					{Time: testutils.Ptr(30), Amount: testutils.Ptr(50)},
					{Time: testutils.Ptr(60), Amount: testutils.Ptr(100)},
				}
			},
			assert: func(res *httpexpect.Response) {
				r := res.Status(http.StatusOK).JSON().Object().Value("recipe").Object()
				phases := r.Value("info").Object().Value("phases").Array()
				phases.Length().Equal(2)
				phases.Element(0).Object().Value("time").Number().Equal(30)
				phases.Element(0).Object().Value("amount").Number().Equal(50)
				phases.Element(1).Object().Value("time").Number().Equal(60)
				phases.Element(1).Object().Value("amount").Number().Equal(100)
				r.Value("version").Number().Equal(2)
			},
		},
		{
			name:           "Modifying the open boolean off of a switch recipe fails",
			sourceMethodID: 2,
			mutate: func(req *testutils.PatchRecipeRequest) {
				// Info has phases without "open" — switch recipe requires it
				req.Info.Phases = []testutils.PatchPhase{
					{Time: testutils.Ptr(30), Amount: testutils.Ptr(50)},
				}
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).JSON().Object().
					Path("$.error.open").String().Contains("must be provided")
			},
		},
		{
			name:           "Modifying the open boolean AND changing method_id to v60 succeeds",
			sourceMethodID: 2,
			mutate: func(req *testutils.PatchRecipeRequest) {
				req.MethodID = testutils.Ptr(int64(1))
				req.Info.Phases = []testutils.PatchPhase{
					{Time: testutils.Ptr(30), Amount: testutils.Ptr(50)},
				}
			},
			assert: func(res *httpexpect.Response) {
				r := res.Status(http.StatusOK).JSON().Object().Value("recipe").Object()
				r.Value("method").Object().Value("name").String().Equal("V60")
				r.Value("info").Object().Value("phases").Array().Element(0).Object().NotContainsKey("open")
				r.Value("version").Number().Equal(2)
			},
		},
		{
			name:           "Successfully setting coffee_id, grinder_id, and grind_size to nil",
			sourceMethodID: 1,
			mutate: func(req *testutils.PatchRecipeRequest) {
				req.NullCoffeeID = true
				req.NullGrinderID = true
				req.Info.GrindSize = ""
			},
			assert: func(res *httpexpect.Response) {
				r := res.Status(http.StatusOK).JSON().Object().Value("recipe").Object()
				r.NotContainsKey("coffee")
				r.NotContainsKey("grinder")
				r.Value("info").Object().NotContainsKey("grind_size")
				r.Value("version").Number().Equal(2)
			},
		},
		{
			name:           "Adding an unknown coffee id fails",
			sourceMethodID: 1,
			mutate: func(req *testutils.PatchRecipeRequest) {
				req.CoffeeID = testutils.Ptr(99999)
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusNotFound).JSON().Object().
					Value("error").String().Equal("the requested coffee could not be found")
			},
		},
		{
			name:           "Adding an unknown method id fails",
			sourceMethodID: 1,
			mutate: func(req *testutils.PatchRecipeRequest) {
				req.MethodID = testutils.Ptr(int64(99999))
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusNotFound).JSON().Object().
					Value("error").String().Equal("the requested method could not be found")
			},
		},
		{
			name:           "Adding an unknown grinder id fails",
			sourceMethodID: 1,
			mutate: func(req *testutils.PatchRecipeRequest) {
				req.GrinderID = testutils.Ptr(int64(99999))
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusNotFound).JSON().Object().
					Value("error").String().Equal("the requested grinder could not be found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var recipeReq testutils.CreateRecipeRequest

			switch tt.sourceMethodID {
			case 1:
				recipeReq = testutils.ValidV60Recipe()
			case 2:
				recipeReq = testutils.ValidSwitchRecipe()
			default:
				t.Fatalf("invalid sourceMethodID %d", tt.sourceMethodID)
			}

			recipe := app.Factory.CreateRecipe(t, token, tt.sourceMethodID, recipeReq)
			req := testutils.ValidPatchRecipeRequest()
			tt.mutate(&req)
			res := app.Client(token).PATCHJSON(fmt.Sprintf("/v1/recipes/%d", recipe.Recipe.ID), req).Expect(t)
			tt.assert(res)
		})
	}

	t.Run("Patching a recipe you don't own returns an error", func(t *testing.T) {
		otherUser := app.Factory.CreateUser(t)
		otherToken := app.Factory.Login(t, otherUser.Email, otherUser.Password)
		recipe := app.Factory.CreateRecipe(t, otherToken, 1, testutils.ValidV60Recipe())

		app.Client(token).PATCHJSON(fmt.Sprintf("/v1/recipes/%d", recipe.Recipe.ID), testutils.ValidPatchRecipeRequest()).
			Expect(t).Status(http.StatusNotFound).JSON().Object().
			Value("error").String().Equal("the requested recipe could not be found")
	})

	t.Run("Patching an unknown recipe returns an error", func(t *testing.T) {
		app.Client(token).PATCHJSON("/v1/recipes/99999", testutils.ValidPatchRecipeRequest()).
			Expect(t).Status(http.StatusNotFound).JSON().Object().
			Value("error").String().Equal("the requested recipe could not be found")
	})

	t.Run("Patching a recipe when not logged in returns an error", func(t *testing.T) {
		recipe := app.Factory.CreateRecipe(t, token, 1, testutils.ValidV60Recipe())
		app.Client("").PATCHJSON(fmt.Sprintf("/v1/recipes/%d", recipe.Recipe.ID), testutils.ValidPatchRecipeRequest()).
			Expect(t).Status(http.StatusUnauthorized)
	})
}

func TestGetOneRecipe(t *testing.T) {
	app := testutils.NewTestApp(t)
	user := app.Factory.CreateUser(t)
	token := app.Factory.Login(t, user.Email, user.Password)

	t.Run("Get one recipe works with valid values", func(t *testing.T) {
		cof := app.Factory.CreateCoffee(t, token, testutils.ValidCoffeeForm())
		gr := app.Factory.CreateGrinder(t, token, testutils.ValidGrinder())

		recipe := &testutils.CreateRecipeRequest{}
		*recipe = testutils.ValidV60Recipe()
		recipe.CoffeeId = cof.Coffee.ID
		recipe.GrinderId = gr.ID
		rec := app.Factory.CreateRecipe(t, token, 1, *recipe)

		res := app.Client(token).GET(fmt.Sprintf("/v1/recipes/%d", rec.Recipe.ID)).
			Expect(t).Status(http.StatusOK).JSON().Object().Value("recipe").Object()

		// top level values ALL exist
		res.Value("id").Number().Equal(rec.Recipe.ID)
		res.Value("user_id").Number().Equal(user.ID)
		res.Value("created_at").String().NotEmpty()
		res.Value("version").Number().Equal(1)

		// Method value check
		method := res.Value("method").Object()
		method.Value("id").Number().Equal(1)
		method.Value("name").String().Equal("V60")
		method.Value("created_at").String().NotEmpty()

		coffee := res.Value("coffee").Object()
		coffee.Value("id").Number().Equal(cof.Coffee.ID)
		coffee.Value("user_id").Number().Equal(user.ID)
		coffee.Value("info").Object().Value("name").String().Equal(cof.Coffee.Info.Name)
		coffee.Value("info").Object().Value("roaster").String().Equal(cof.Coffee.Info.Roaster)
		coffee.Value("info").Object().Value("region").String().Equal(cof.Coffee.Info.Region)
		coffee.Value("info").Object().Value("process").String().Equal(cof.Coffee.Info.Process)
		coffee.Value("info").Object().Value("decaf").Boolean().Equal(cof.Coffee.Info.Decaf)
		coffee.Value("info").Object().Value("origin_type").String().Equal(cof.Coffee.Info.OriginType)
		coffee.Value("info").Object().Value("rating").Number().Equal(float64(cof.Coffee.Info.Rating))
		coffee.Value("info").Object().Value("tasting_notes").Array().Length().Equal(len(cof.Coffee.Info.TastingNotes))
		coffee.Value("info").Object().Value("roast_level").String().Equal(cof.Coffee.Info.RoastLevel)
		coffee.Value("info").Object().Value("cost").Number().Equal(cof.Coffee.Info.Cost)
		coffee.Value("info").Object().Value("img").String().Equal(cof.Coffee.Info.Img)
		coffee.Value("info").Object().Value("description").String().Equal(cof.Coffee.Info.Description)
		coffee.Value("info").Object().Value("variety").String().Equal(cof.Coffee.Info.Variety)

		grinder := res.Value("grinder").Object()
		grinder.Value("id").Number().Equal(gr.ID)
		grinder.Value("user_id").Number().Equal(user.ID)
		grinder.Value("name").String().Equal(gr.Name)
		grinder.Value("created_at").String().NotEmpty()
		grinder.Value("version").Number().Equal(1)

		info := res.Value("info").Object()
		info.Value("name").String().Equal(testutils.ValidV60Info().Name)
		info.Value("grams_in").Number().Equal(testutils.ValidV60Info().GramIn)
		info.Value("ml_out").Number().Equal(testutils.ValidV60Info().MlOut)
		info.Value("water_temp").String().Equal(testutils.ValidV60Info().WaterTemp)
		info.Value("grind_size").String().Equal(testutils.ValidV60Info().GrindSize)
		info.Value("phases").Array().Length().Equal(len(testutils.ValidV60Info().Phases))
	})

	t.Run("Get one recipe works if there is no coffee or grinder", func(t *testing.T) {
		rec := app.Factory.CreateRecipe(t, token, 1, testutils.ValidV60Recipe())

		res := app.Client(token).GET(fmt.Sprintf("/v1/recipes/%d", rec.Recipe.ID)).
			Expect(t).Status(http.StatusOK).JSON().Object().Value("recipe").Object()

		res.Value("id").Number().Equal(rec.Recipe.ID)
		res.NotContainsKey("coffee")
		res.NotContainsKey("grinder")
		res.Value("info").Object().NotEmpty()
	})

	t.Run("Get one recipe return an error if it does not exist", func(t *testing.T) {
		app.Client(token).GET("/v1/recipes/99999").
			Expect(t).Status(http.StatusNotFound).JSON().Object().
			Value("error").String().Equal("the requested recipe could not be found")
	})

	t.Run("Get one recipe returns an error if the user does not own one", func(t *testing.T) {
		recipe := app.Factory.CreateRecipe(t, token, 1, testutils.ValidV60Recipe())
		otherUser := app.Factory.CreateUser(t)
		otherToken := app.Factory.Login(t, otherUser.Email, otherUser.Password)
		app.Client(otherToken).GET(fmt.Sprintf("/v1/recipes/%d", recipe.Recipe.ID)).
			Expect(t).Status(http.StatusNotFound).JSON().Object().
			Value("error").String().Equal("the requested recipe could not be found")
	})
}

func TestDeleteRecipe(t *testing.T) {
	app := testutils.NewTestApp(t)
	user := app.Factory.CreateUser(t)
	token := app.Factory.Login(t, user.Email, user.Password)

	t.Run("DELETE recipe successfully deletes a recipe", func(t *testing.T) {
		recipe := app.Factory.CreateRecipe(t, token, 1, testutils.ValidV60Recipe())
		app.Client(token).GET(fmt.Sprintf("/v1/recipes/%d", recipe.Recipe.ID)).Expect(t).Status(http.StatusOK)
		app.Client(token).DELETE(fmt.Sprintf("/v1/recipes/%d", recipe.Recipe.ID)).Expect(t).Status(http.StatusOK).
			JSON().Object().Value("message").String().Contains("deleted")
		app.Client(token).GET(fmt.Sprintf("/v1/recipes/%d", recipe.Recipe.ID)).Expect(t).Status(http.StatusNotFound)
	})

	t.Run("DELETE recipe that is not found returns an error", func(t *testing.T) {
		app.Client(token).DELETE("/v1/recipes/99999").Expect(t).Status(http.StatusNotFound).
			JSON().Object().Value("error").String().Equal("the requested recipe could not be found")
	})

	t.Run("DELETE recipe that you do not own returns an error", func(t *testing.T) {
		recipe := app.Factory.CreateRecipe(t, token, 1, testutils.ValidV60Recipe())
		otherUser := app.Factory.CreateUser(t)
		otherToken := app.Factory.Login(t, otherUser.Email, otherUser.Password)
		app.Client(otherToken).DELETE(fmt.Sprintf("/v1/recipes/%d", recipe.Recipe.ID)).Expect(t).Status(http.StatusNotFound)
	})

	t.Run("DELETE recipe when you are not logged in returns an error", func(t *testing.T) {
		recipe := app.Factory.CreateRecipe(t, token, 1, testutils.ValidV60Recipe())
		app.Client("").DELETE(fmt.Sprintf("/v1/recipes/%d", recipe.Recipe.ID)).Expect(t).Status(http.StatusUnauthorized)
	})
}
