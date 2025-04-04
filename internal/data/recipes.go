package data

import (
	"context"
	"database/sql"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"time"
)

type Recipe struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	MethodID  int        `json:"method_id"`
	CoffeeID  int        `json:"coffee_id"`
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
	v.Check(phase.Amount > 0, "amount", "must be greater than zero")
}

type RecipeModelInterface interface {
	Insert(recipe *Recipe) error
}

func (m RecipeModel) Insert(recipe *Recipe) error {
	query := `INSERT INTO recipes (user_id, method_id, coffee_id, info) VALUES ($1, $2, $3, $4)`

	args := []any{recipe.UserID, recipe.MethodID, recipe.CoffeeID, recipe.Info}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&recipe.ID, &recipe.CreatedAt, &recipe.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}
