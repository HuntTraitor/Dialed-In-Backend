package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"time"
)

type Recipe struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	MethodID  int64      `json:"method_id"`
	CoffeeID  int64      `json:"coffee_id"`
	Info      RecipeInfo `json:"info"`
	CreatedAt string     `json:"created_at"`
	Version   int        `json:"version"`
}

type FullRecipe struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	Method    Method     `json:"method"`
	Coffee    Coffee     `json:"coffee"`
	Info      RecipeInfo `json:"info"`
	CreatedAt string     `json:"created_at"`
	Version   int        `json:"version"`
}

type RecipeInfo struct {
	Name   string  `json:"name"`
	GramIn int     `json:"grams_in"`
	MlOut  int     `json:"ml_out"`
	Phases []Phase `json:"phases"`
}

type Phase struct {
	Open   bool `json:"open"`
	Time   int  `json:"time"`
	Amount int  `json:"amount"`
}

type RecipeModel struct {
	DB *sql.DB
}

// ValidateRecipe validates a specific recipe is correct
func ValidateRecipe(v *validator.Validator, recipe *Recipe) {
	v.Check(recipe.Info.Name != "", "name", "must be provided")
	v.Check(len(recipe.Info.Name) <= 100, "name", "must not be more than 100 bytes")
	v.Check(recipe.Info.MlOut > 0, "ml_out", "must be greater than zero")
	v.Check(recipe.Info.GramIn > 0, "grams_in", "must be greater than zero")
	v.Check(recipe.Info.MlOut < 1000, "ml_out", "must be less than a thousand")
	v.Check(recipe.Info.GramIn < 10000, "grams_in", "must be less than ten thousand")
	for _, phase := range recipe.Info.Phases {
		ValidatePhase(v, &phase)
	}
}

// ValidatePhase validates a phase is the correct format
func ValidatePhase(v *validator.Validator, phase *Phase) {
	v.Check(phase.Time > 0, "time", "must be greater than zero")
	v.Check(phase.Amount >= 0, "amount", "must be greater than or equal to zero")
}

type RecipeModelInterface interface {
	Insert(recipe *Recipe) error
	GetAllForUser(userID int64, methodID int64, coffeeID int64) ([]*Recipe, error)
}

func (m RecipeModel) Insert(recipe *Recipe) error {
	query := `INSERT INTO recipes (user_id, method_id, coffee_id, info) VALUES ($1, $2, $3, $4) 
            RETURNING id, created_at, version`

	infoJSON, err := json.Marshal(recipe.Info)
	if err != nil {
		return err
	}

	args := []any{recipe.UserID, recipe.MethodID, recipe.CoffeeID, infoJSON}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&recipe.ID, &recipe.CreatedAt, &recipe.Version)
	if err != nil {
		return err
	}
	return nil
}

func (m RecipeModel) GetAllForUser(userID int64, methodID int64, coffeeID int64) ([]*Recipe, error) {
	query := `SELECT * FROM recipes 
         		WHERE user_id = $1
         		AND (method_id = $2 OR $2 = 0)
         		AND (coffee_id = $3 OR $3 = 0)
         		ORDER BY id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userID, methodID, coffeeID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	recipes := []*Recipe{}
	var infoBytes []byte

	for rows.Next() {
		var recipe Recipe
		err = rows.Scan(
			&recipe.ID,
			&recipe.UserID,
			&recipe.CoffeeID,
			&recipe.MethodID,
			&infoBytes,
			&recipe.Version,
			&recipe.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(infoBytes, &recipe.Info)
		if err != nil {
			return nil, err
		}
		recipes = append(recipes, &recipe)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return recipes, nil
}
