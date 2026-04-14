package e2e

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/hunttraitor/dialed-in-backend/e2e/testutils"
	"github.com/stretchr/testify/assert"
)

func TestGetAllCoffees(t *testing.T) {
	app := testutils.NewTestApp(t)

	user := app.Factory.CreateUser(t)
	token := app.Factory.Login(t, user.Email, user.Password)
	app.Factory.CreateCoffee(t, token, testutils.ValidCoffeeForm())

	t.Run("Successfully gets list of coffees", func(t *testing.T) {
		res := app.Client(token).GET("/v1/coffees").Expect(t)

		obj := res.Status(http.StatusOK).JSON().Object()

		coffees := obj.Value("coffees").Array()
		coffees.NotEmpty()

		for i := 0; i < len(coffees.Iter()); i++ {
			c := coffees.Element(i).Object()

			c.Value("id").Number().Gt(0)
			c.Value("user_id").Number().Equal(user.ID)
			c.Value("created_at").String().NotEmpty()
			c.Value("version").Number().Equal(1)

			info := c.Value("info").Object()
			info.Value("img").String().NotEmpty()
			info.Value("name").String().Equal(testutils.ValidCoffeeForm().Name)
			info.Value("roaster").String().Equal(testutils.ValidCoffeeForm().Roaster)
			info.Value("region").String().Equal(testutils.ValidCoffeeForm().Region)
			info.Value("process").String().Equal(testutils.ValidCoffeeForm().Process)
			info.Value("description").String().Equal(testutils.ValidCoffeeForm().Description)
			info.Value("origin_type").String().Equal(testutils.ValidCoffeeForm().OriginType)
			info.Value("rating").Number().Equal(testutils.ValidCoffeeForm().Rating)
			info.Value("cost").Number().Equal(testutils.ValidCoffeeForm().Cost)
			info.Value("roast_level").String().Equal(testutils.ValidCoffeeForm().RoastLevel)
			info.Value("decaf").Boolean().Equal(testutils.ValidCoffeeForm().Decaf)
			info.Value("variety").String().Equal(testutils.ValidCoffeeForm().Variety)

			notes := info.Value("tasting_notes").Array().Raw()
			actualNotes := make([]string, len(notes))
			for j, note := range notes {
				actualNotes[j] = note.(string)
			}
			assert.ElementsMatch(t, testutils.ValidCoffeeForm().TastingNotes, actualNotes)
		}
	})

	t.Run("Successfully applies all coffee filters in one query", func(t *testing.T) {
		target := testutils.ValidCoffeeForm()
		target.Name = "Blueberry Smoothie Bomb"
		target.Roaster = "Roasters Lab Collective"
		target.Region = "Yirgacheffe Chelbesa"
		target.Process = "Honey Anaerobic"
		target.Description = "Juicy cup with loud fruit sweetness"
		target.OriginType = "Single Origin"
		target.TastingNotes = []string{"Blueberry Smoothie", "Jasmine Blossom", "Molasses"}
		target.Rating = 5
		target.RoastLevel = "Light"
		target.Cost = 27.75
		target.Decaf = false
		target.Variety = "Heirloom 74110"

		targetCoffee := app.Factory.CreateCoffee(t, token, target)

		decoy := testutils.ValidCoffeeForm()
		decoy.Name = "Blueberry Smoothie Decaf"
		decoy.Roaster = "Roasters Lab Collective"
		decoy.Region = "Yirgacheffe Chelbesa"
		decoy.Process = "Honey Anaerobic"
		decoy.OriginType = "Single Origin"
		decoy.TastingNotes = []string{"Blueberry Smoothie", "Jasmine Blossom"}
		decoy.Rating = 5
		decoy.RoastLevel = "Light"
		decoy.Cost = 27.75
		decoy.Decaf = true
		decoy.Variety = "Heirloom 74110"
		app.Factory.CreateCoffee(t, token, decoy)

		res := app.Client(token).
			GET("/v1/coffees").
			WithQuery("name", "Blueberry sm").
			WithQuery("roaster", "Roasters Lab").
			WithQuery("region", "Yirgach").
			WithQuery("process", "Honey Anaero").
			WithQuery("variety", "Heirloom 741").
			WithQuery("origin_type", "blend,Single Origin,micro-lot").
			WithQuery("roast_level", "dark,Light,medium-dark").
			WithQuery("decaf", "false").
			WithQuery("rating", "1,5,3").
			WithQuery("tasting_notes", "grapefruit,BLUEBERRY SMOOTHIE,molasses").
			WithQuery("min_cost", "27.70").
			WithQuery("max_cost", "27.80").
			Expect(t).
			Status(http.StatusOK)
		coffees := res.JSON().Object().Value("coffees").Array()
		coffees.Length().Equal(1)

		coffee := coffees.Element(0).Object()
		coffee.Value("id").Number().Equal(targetCoffee.Coffee.ID)

		info := coffee.Value("info").Object()
		info.Value("name").String().Equal(target.Name)
		info.Value("roaster").String().Equal(target.Roaster)
		info.Value("region").String().Equal(target.Region)
		info.Value("process").String().Equal(target.Process)
		info.Value("origin_type").String().Equal(target.OriginType)
		info.Value("roast_level").String().Equal(target.RoastLevel)
		info.Value("rating").Number().Equal(target.Rating)
		info.Value("decaf").Boolean().Equal(target.Decaf)
		info.Value("cost").Number().Equal(target.Cost)
		info.Value("variety").String().Equal(target.Variety)
	})

	t.Run("Fails to get coffees when not logged in", func(t *testing.T) {
		app.Client("").GET("/v1/coffees").Expect(t).Status(http.StatusUnauthorized)
	})
}

func TestGetOneCoffee(t *testing.T) {
	app := testutils.NewTestApp(t)

	t.Run("Getting successful coffee returns 200", func(t *testing.T) {
		user := app.Factory.CreateUser(t)
		token := app.Factory.Login(t, user.Email, user.Password)
		coffee := app.Factory.CreateCoffee(t, token, testutils.ValidCoffeeForm())
		res := app.Client(token).GET(fmt.Sprintf("/v1/coffees/%d", coffee.Coffee.ID)).
			Expect(t).Status(http.StatusOK)

		c := res.JSON().Object().Value("coffee").Object()
		c.Value("id").Number().Equal(coffee.Coffee.ID)
		c.Value("user_id").Number().Equal(user.ID)
		c.Value("created_at").String().NotEmpty()
		c.Value("version").Number().Equal(1)

		info := c.Value("info").Object()
		info.Value("name").String().Equal(testutils.ValidCoffeeForm().Name)
		info.Value("roaster").String().Equal(testutils.ValidCoffeeForm().Roaster)
		info.Value("region").String().Equal(testutils.ValidCoffeeForm().Region)
		info.Value("process").String().Equal(testutils.ValidCoffeeForm().Process)
		info.Value("description").String().Equal(testutils.ValidCoffeeForm().Description)
		info.Value("origin_type").String().Equal(testutils.ValidCoffeeForm().OriginType)
		info.Value("rating").Number().Equal(testutils.ValidCoffeeForm().Rating)
		info.Value("cost").Number().Equal(testutils.ValidCoffeeForm().Cost)
		info.Value("roast_level").String().Equal(testutils.ValidCoffeeForm().RoastLevel)
		info.Value("decaf").Boolean().Equal(testutils.ValidCoffeeForm().Decaf)
		info.Value("variety").String().Equal(testutils.ValidCoffeeForm().Variety)
		notes := info.Value("tasting_notes").Array().Raw()
		actualNotes := make([]string, len(notes))
		for j, note := range notes {
			actualNotes[j] = note.(string)
		}
		assert.ElementsMatch(t, testutils.ValidCoffeeForm().TastingNotes, actualNotes)
	})

	t.Run("Getting coffee that doesn't exist returns 404", func(t *testing.T) {
		user := app.Factory.CreateUser(t)
		token := app.Factory.Login(t, user.Email, user.Password)
		app.Client(token).GET("/v1/coffees/0").Expect(t).Status(http.StatusNotFound)
	})

	t.Run("Getting coffee that doesn't belong to user returns 404", func(t *testing.T) {
		user := app.Factory.CreateUser(t)
		token := app.Factory.Login(t, user.Email, user.Password)

		user2 := app.Factory.CreateUser(t)
		token2 := app.Factory.Login(t, user2.Email, user2.Password)

		coffee := app.Factory.CreateCoffee(t, token, testutils.ValidCoffeeForm())
		app.Client(token).GET(fmt.Sprintf("/v1/coffees/%d", coffee.Coffee.ID)).Expect(t).Status(http.StatusOK)
		app.Client(token2).GET(fmt.Sprintf("/v1/coffees/%d", coffee.Coffee.ID)).Expect(t).Status(http.StatusNotFound)
	})

	t.Run("Getting coffee unauthenticated is a 403", func(t *testing.T) {
		app.Client("").GET(fmt.Sprintf("/v1/coffees/%d", 1)).Expect(t).Status(http.StatusUnauthorized)
	})
}

func TestPostCoffee(t *testing.T) {
	app := testutils.NewTestApp(t)
	user := app.Factory.CreateUser(t)
	token := app.Factory.Login(t, user.Email, user.Password)

	longName := strings.Repeat("A", 510)
	longRoaster := strings.Repeat("A", 510)
	longRegion := strings.Repeat("B", 110)
	longProcess := strings.Repeat("p", 210)
	longDescription := strings.Repeat("C", 1010)
	longOriginType := strings.Repeat("o", 110)
	longTastingNote := []string{strings.Repeat("t", 110)}
	longVariety := strings.Repeat("v", 210)
	longTastingNotes := make([]string, 51)

	for i := range longTastingNotes {
		longTastingNotes[i] = fmt.Sprintf("note%d", i+1)
	}

	tests := []struct {
		name   string
		mutate func(*testutils.CoffeeForm)
		assert func(*httpexpect.Response)
	}{
		{
			name:   "Successfully posts a new coffee",
			mutate: func(form *testutils.CoffeeForm) {},
			assert: func(res *httpexpect.Response) {
				coffee := res.Status(http.StatusCreated).JSON().Object().Value("coffee").Object()
				info := coffee.Value("info").Object()

				info.Value("name").String().Equal(testutils.ValidCoffeeForm().Name)
				info.Value("img").String().NotEmpty()
				info.Value("roaster").String().Equal(testutils.ValidCoffeeForm().Roaster)
				info.Value("region").String().Equal(testutils.ValidCoffeeForm().Region)
				info.Value("process").String().Equal(testutils.ValidCoffeeForm().Process)
				info.Value("description").String().Equal(testutils.ValidCoffeeForm().Description)
				info.Value("origin_type").String().Equal(testutils.ValidCoffeeForm().OriginType)
				info.Value("rating").Number().Equal(testutils.ValidCoffeeForm().Rating)
				info.Value("cost").Number().Equal(testutils.ValidCoffeeForm().Cost)
				info.Value("roast_level").String().Equal(testutils.ValidCoffeeForm().RoastLevel)
				info.Value("decaf").Boolean().Equal(testutils.ValidCoffeeForm().Decaf)
				info.Value("variety").String().Equal(testutils.ValidCoffeeForm().Variety)

				coffee.Value("id").Number().Gt(0)
				coffee.Value("user_id").Number().Equal(user.ID)
				coffee.Value("created_at").String().NotEmpty()
				coffee.Value("version").Number().Gt(0)
			},
		},
		{
			name: "Coffee name too long returns an error",
			mutate: func(form *testutils.CoffeeForm) {
				form.Name = longName
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path("$.error.name").String().Equal("must not be more than 500 bytes long")
			},
		},
		{
			name: "Coffee roaster too long return an error",
			mutate: func(form *testutils.CoffeeForm) {
				form.Roaster = longRoaster
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path("$.error.roaster").String().Equal("must not be more than 200 bytes long")
			},
		},
		{
			name: "Coffee description too long returns an error",
			mutate: func(form *testutils.CoffeeForm) {
				form.Description = longDescription
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path("$.error.description").String().Equal("must not be more than 1000 bytes long")
			},
		},
		{
			name: "Coffee region too long returns an error",
			mutate: func(form *testutils.CoffeeForm) {
				form.Region = longRegion
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path("$.error.region").String().Equal("must not be more than 100 bytes long")
			},
		},
		{
			name: "Coffee process name too long returns an error",
			mutate: func(form *testutils.CoffeeForm) {
				form.Process = longProcess
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path("$.error.process").String().Equal("must not be more than 200 bytes long")
			},
		},
		{
			name: "Coffee Origin Type too long returns an error",
			mutate: func(form *testutils.CoffeeForm) {
				form.OriginType = longOriginType
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path("$.error.origin_type").String().Equal("must not be more than 100 bytes long")
			},
		},
		{
			name: "Coffee Tasting Notes amount too long returns an error",
			mutate: func(form *testutils.CoffeeForm) {
				form.TastingNotes = longTastingNotes
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path("$.error.tasting_notes").String().Equal("must not contain more than 50 entries")
			},
		},
		{
			name: "Coffee Testing Note length too long returns an error",
			mutate: func(form *testutils.CoffeeForm) {
				form.TastingNotes = longTastingNote
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path(`$.error["tasting_notes[0]"]`).String().
					Equal("must not be more than 100 bytes long")
			},
		},
		{
			name: "Coffee rating more than 5 returns an error",
			mutate: func(form *testutils.CoffeeForm) {
				form.Rating = 6
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path("$.error.rating").String().Equal("must be between 0 and 5")
			},
		},
		{
			name: "Coffee cost less than zero returns an error",
			mutate: func(form *testutils.CoffeeForm) {
				form.Cost = -1
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path("$.error.cost").String().Equal("must not be less than 0")
			},
		},
		{
			name: "Coffee cost more than 1,000,000 returns an error",
			mutate: func(form *testutils.CoffeeForm) {
				form.Cost = 1000001
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path("$.error.cost").String().Equal("must not be more than 1,000,000")
			},
		},
		{
			name: "Cofee varieties name too long returns an error",
			mutate: func(form *testutils.CoffeeForm) {
				form.Variety = longVariety
			},
			assert: func(res *httpexpect.Response) {
				res.Status(http.StatusUnprocessableEntity).
					JSON().Object().
					Path("$.error.variety").String().Equal("must not be more than 200 bytes long")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := testutils.ValidCoffeeForm()
			tt.mutate(&form)

			res := app.Client(token).
				POSTMultipart("/v1/coffees", form).Expect(t)

			tt.assert(res)
		})
	}

	t.Run("Creating minimal coffee omits all appropriate fields", func(t *testing.T) {
		res := app.Client(token).
			POSTMultipart("/v1/coffees", testutils.MinimalCoffeeForm("Test Name")).Expect(t)

		coffee := res.Status(http.StatusCreated).JSON().Object().Value("coffee").Object()
		coffee.Value("id").Number().Gt(0)
		coffee.Value("user_id").Number().Equal(user.ID)
		coffee.Value("created_at").String().NotEmpty()
		coffee.Value("version").Number().Equal(1)

		info := coffee.Value("info").Object()

		info.Value("name").String().Equal("Test Name")
		info.Value("decaf").Boolean().Equal(false)
	})

	t.Run("Empty params returns an error", func(t *testing.T) {
		app.Client(token).POSTMultipart("/v1/coffees", testutils.EmptyCoffeeForm()).
			Expect(t).Status(http.StatusUnprocessableEntity).
			JSON().Object().Path("$.error.name").String().Equal("must be provided")
	})

	t.Run("Unauthenticated call response when an error", func(t *testing.T) {
		form := testutils.ValidCoffeeForm()
		app.Client("").POSTMultipart("/v1/coffees", form).
			Expect(t).Status(http.StatusUnauthorized)
	})
}

func TestUpdateCoffee(t *testing.T) {
	app := testutils.NewTestApp(t)

	user := app.Factory.CreateUser(t)
	token := app.Factory.Login(t, user.Email, user.Password)

	tests := []struct {
		name   string
		mutate func(*testutils.CoffeeForm)
		assert func(*httpexpect.Response)
	}{
		{
			name: "Successfully updates a coffee",
			mutate: func(form *testutils.CoffeeForm) {
				form.Name = "Updated Name"
				form.Roaster = "Updated Roaster"
				form.Region = "Updated Region"
				form.Process = "Updated Process"
				form.Description = "Updated Description"
				form.OriginType = "Updated Origin Type"
				form.Cost = 100
				form.RoastLevel = "Updated Roast Level"
				form.Rating = 2
				form.Decaf = true
				form.TastingNotes = []string{"chocolate", "caramel"}
				form.Img = []byte("Updated Image")
				form.Variety = "Updated Variety"
			},
			assert: func(res *httpexpect.Response) {
				info := res.Status(http.StatusOK).
					JSON().Object().
					Value("coffee").Object().
					Value("info").Object()

				info.Value("name").String().Equal("Updated Name")
				info.Value("roaster").String().Equal("Updated Roaster")
				info.Value("region").String().Equal("Updated Region")
				info.Value("process").String().Equal("Updated Process")
				info.Value("description").String().Equal("Updated Description")
				info.Value("origin_type").String().Equal("Updated Origin Type")
				info.Value("cost").Number().Equal(100)
				info.Value("roast_level").String().Equal("Updated Roast Level")
				info.Value("rating").Number().Equal(2)
				info.Value("decaf").Boolean().True()
				info.Value("variety").String().Equal("Updated Variety")
				info.Value("img").String().NotEmpty()
			},
		},
		{
			name: "Successfully Partially updates a coffee",
			mutate: func(form *testutils.CoffeeForm) {
				*form = testutils.EmptyCoffeeForm()
				form.Name = "Updated Name"
			},
			assert: func(res *httpexpect.Response) {
				info := res.Status(http.StatusOK).
					JSON().Object().
					Value("coffee").Object().
					Value("info").Object()

				info.Value("name").String().Equal("Updated Name")
				info.Value("roaster").String().Equal(testutils.ValidCoffeeForm().Roaster)
				info.Value("region").String().Equal(testutils.ValidCoffeeForm().Region)
				info.Value("process").String().Equal(testutils.ValidCoffeeForm().Process)
				info.Value("description").String().Equal(testutils.ValidCoffeeForm().Description)
				info.Value("origin_type").String().Equal(testutils.ValidCoffeeForm().OriginType)
				info.Value("cost").Number().Equal(testutils.ValidCoffeeForm().Cost)
				info.Value("roast_level").String().Equal(testutils.ValidCoffeeForm().RoastLevel)
				info.Value("rating").Number().Equal(float64(testutils.ValidCoffeeForm().Rating))
				info.Value("decaf").Boolean().Equal(testutils.ValidCoffeeForm().Decaf)
				info.Value("variety").String().Equal(testutils.ValidCoffeeForm().Variety)
			},
		},
		{
			name: "Updating with no fields is still successful",
			mutate: func(form *testutils.CoffeeForm) {
				*form = testutils.EmptyCoffeeForm()
			},
			assert: func(res *httpexpect.Response) {
				info := res.Status(http.StatusOK).
					JSON().Object().
					Value("coffee").Object().
					Value("info").Object()

				info.Value("name").String().Equal(testutils.ValidCoffeeForm().Name)
				info.Value("roaster").String().Equal(testutils.ValidCoffeeForm().Roaster)
				info.Value("region").String().Equal(testutils.ValidCoffeeForm().Region)
				info.Value("process").String().Equal(testutils.ValidCoffeeForm().Process)
				info.Value("description").String().Equal(testutils.ValidCoffeeForm().Description)
				info.Value("origin_type").String().Equal(testutils.ValidCoffeeForm().OriginType)
				info.Value("cost").Number().Equal(testutils.ValidCoffeeForm().Cost)
				info.Value("roast_level").String().Equal(testutils.ValidCoffeeForm().RoastLevel)
				info.Value("rating").Number().Equal(float64(testutils.ValidCoffeeForm().Rating))
				info.Value("decaf").Boolean().Equal(testutils.ValidCoffeeForm().Decaf)
				info.Value("variety").String().Equal(testutils.ValidCoffeeForm().Variety)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			posted := app.Factory.CreateCoffee(t, token, testutils.ValidCoffeeForm())
			coffeeID := posted.Coffee.ID

			form := testutils.ValidCoffeeForm()
			tt.mutate(&form)

			res := app.Client(token).
				PATCHMultipart(fmt.Sprintf("/v1/coffees/%d", coffeeID), form).
				Expect(t)

			tt.assert(res)
		})
	}

	t.Run("Update with an unknown field returns an error", func(t *testing.T) {
		posted := app.Factory.CreateCoffee(t, token, testutils.ValidCoffeeForm())

		app.Client(token).
			PATCHMultipartWithExtraFields(
				fmt.Sprintf("/v1/coffees/%d", posted.Coffee.ID),
				testutils.ValidCoffeeForm(),
				map[string]string{"unknown_field": "unknown_value"},
			).
			Expect(t).Status(http.StatusBadRequest).
			JSON().Object().ValueEqual("error", `body contains unknown key "unknown_field"`)
	})

	t.Run("Update with a known AND unknown field returns an error", func(t *testing.T) {
		posted := app.Factory.CreateCoffee(t, token, testutils.ValidCoffeeForm())

		form := testutils.EmptyCoffeeForm()
		form.Name = "Updated Name"
		form.Region = "Updated Region"
		form.Process = "Updated Process"
		form.Description = "Updated Description"

		app.Client(token).
			PATCHMultipartWithExtraFields(
				fmt.Sprintf("/v1/coffees/%d", posted.Coffee.ID),
				form,
				map[string]string{"random_field": "unknown"},
			).
			Expect(t).
			Status(http.StatusBadRequest).
			JSON().Object().
			ValueEqual("error", `body contains unknown key "random_field"`)
	})

	t.Run("Updating an item that does not exist returns an error", func(t *testing.T) {
		app.Client(token).
			PATCHMultipart("/v1/coffees/47834957", testutils.EmptyCoffeeForm()).
			Expect(t).
			Status(http.StatusNotFound)
	})

	t.Run("Updating an item that the user does not own returns an error", func(t *testing.T) {
		otherUser := app.Factory.CreateUser(t)
		otherToken := app.Factory.Login(t, otherUser.Email, otherUser.Password)

		otherCoffee := app.Factory.CreateCoffee(t, otherToken, testutils.ValidCoffeeForm())

		app.Client(token).
			PATCHMultipart(fmt.Sprintf("/v1/coffees/%d", otherCoffee.Coffee.ID), testutils.EmptyCoffeeForm()).
			Expect(t).
			Status(http.StatusNotFound)
	})

	t.Run("Unauthenticated user updating a coffee returns an error", func(t *testing.T) {
		app.Client("").
			PATCHMultipart("/v1/coffees/1", testutils.EmptyCoffeeForm()).
			Expect(t).
			Status(http.StatusUnauthorized)
	})

	t.Run("Sending a patch request that is not a multi part form returns an error", func(t *testing.T) {
		posted := app.Factory.CreateCoffee(t, token, testutils.ValidCoffeeForm())

		app.Client(token).
			PATCHJSON(fmt.Sprintf("/v1/coffees/%d", posted.Coffee.ID), map[string]any{}).
			Expect(t).
			Status(http.StatusBadRequest).
			JSON().Object().
			ValueEqual("error", "content type must be multipart/form-data")
	})
}

func TestDeleteCoffee(t *testing.T) {
	app := testutils.NewTestApp(t)

	user := app.Factory.CreateUser(t)
	token := app.Factory.Login(t, user.Email, user.Password)

	t.Run("Successfully deletes a coffee", func(t *testing.T) {
		posted := app.Factory.CreateCoffee(t, token, testutils.ValidCoffeeForm())
		coffeeID := posted.Coffee.ID

		app.Client(token).
			GET(fmt.Sprintf("/v1/coffees/%d", coffeeID)).
			Expect(t).
			Status(http.StatusOK)

		app.Client(token).
			DELETE(fmt.Sprintf("/v1/coffees/%d", coffeeID)).
			Expect(t).
			Status(http.StatusOK).
			JSON().Object().
			Value("message").String().NotEmpty()

		app.Client(token).
			GET(fmt.Sprintf("/v1/coffees/%d", coffeeID)).
			Expect(t).
			Status(http.StatusNotFound)
	})

	t.Run("Deleting a coffee that does not exist returns an error", func(t *testing.T) {
		app.Client(token).
			DELETE("/v1/coffees/10000").
			Expect(t).
			Status(http.StatusNotFound)
	})

	t.Run("Deleting a coffee that the user does not own returns an error", func(t *testing.T) {
		otherUser := app.Factory.CreateUser(t)
		otherToken := app.Factory.Login(t, otherUser.Email, otherUser.Password)

		otherCoffee := app.Factory.CreateCoffee(t, otherToken, testutils.ValidCoffeeForm())

		app.Client(token).
			DELETE(fmt.Sprintf("/v1/coffees/%d", otherCoffee.Coffee.ID)).
			Expect(t).
			Status(http.StatusNotFound)
	})

	t.Run("Deleting a coffee when the user is not authenticated returns an error", func(t *testing.T) {
		app.Client("").
			DELETE("/v1/coffees/1").
			Expect(t).
			Status(http.StatusUnauthorized)
	})
}
