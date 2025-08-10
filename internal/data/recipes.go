package data

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"net/url"
	"strconv"
	"time"
)

type Recipe struct {
	ID        int64           `json:"id,omitempty"`
	UserID    int64           `json:"user_id,omitempty"`
	MethodID  int64           `json:"method_id"`
	CoffeeID  int64           `json:"coffee_id"`
	Info      json.RawMessage `json:"info"`
	CreatedAt string          `json:"created_at,omitempty"`
	Version   int             `json:"version,omitempty"`
}

type FullRecipe struct {
	ID        int64           `json:"id"`
	UserID    int64           `json:"user_id"`
	Method    Method          `json:"method"`
	Coffee    Coffee          `json:"coffee"`
	Info      json.RawMessage `json:"info"`
	CreatedAt string          `json:"created_at"`
	Version   int             `json:"version"`
}

type SwitchRecipeInfo struct {
	Name   string        `json:"name"`
	GramIn int           `json:"grams_in"`
	MlOut  int           `json:"ml_out"`
	Phases []SwitchPhase `json:"phases"`
}

type SwitchPhase struct {
	Open   *bool `json:"open"`
	Time   int   `json:"time"`
	Amount int   `json:"amount"`
}

type V60RecipeInfo struct {
	Name   string     `json:"name"`
	GramIn int        `json:"grams_in"`
	MlOut  int        `json:"ml_out"`
	Phases []V60Phase `json:"phases"`
}

type V60Phase struct {
	Time   int `json:"time"`
	Amount int `json:"amount"`
}

type RecipeModel struct {
	DB *sql.DB
}

// decodeInfoStrict strictly (allowing no extra values) decodes a json message into type T
func decodeInfoStrict[T any](v *validator.Validator, raw json.RawMessage) (T, bool) {
	var typed T
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&typed); err != nil {
		v.AddError("info", fmt.Sprintf("invalid format: %v", err))
		return typed, false
	}
	return typed, true
}

// ValidateRecipe validates a specific Recipe is correct based on the Method
func ValidateRecipe(v *validator.Validator, recipe *Recipe, method *Method) {
	switch method.ID {
	case 1:
		info, ok := decodeInfoStrict[V60RecipeInfo](v, recipe.Info)
		if !ok {
			return
		}
		v.Check(info.Name != "", "name", "must be provided")
		v.Check(len(info.Name) <= 100, "name", "must not be more than 100 bytes")
		v.Check(info.MlOut > 0, "ml_out", "must be greater than zero")
		v.Check(info.GramIn > 0, "grams_in", "must be greater than zero")
		v.Check(info.MlOut < 1000, "ml_out", "must be less than a thousand")
		v.Check(info.GramIn < 10000, "grams_in", "must be less than ten thousand")
		for _, phase := range info.Phases {
			ValidateV60Phase(v, &phase)
		}

	case 2:
		fmt.Println(recipe)
		info, ok := decodeInfoStrict[SwitchRecipeInfo](v, recipe.Info)
		if !ok {
			return
		}
		v.Check(info.Name != "", "name", "must be provided")
		v.Check(len(info.Name) <= 100, "name", "must not be more than 100 bytes")
		v.Check(info.MlOut > 0, "ml_out", "must be greater than zero")
		v.Check(info.GramIn > 0, "grams_in", "must be greater than zero")
		v.Check(info.MlOut < 1000, "ml_out", "must be less than a thousand")
		v.Check(info.GramIn < 10000, "grams_in", "must be less than ten thousand")
		for _, phase := range info.Phases {
			ValidateSwitchPhase(v, &phase)
		}

	default:
		v.AddError("info", "is in an known format")
	}
}

// ValidateSwitchPhase validates a switch phase is the correct format
func ValidateSwitchPhase(v *validator.Validator, phase *SwitchPhase) {
	v.Check(phase.Open != nil, "open", "must be provided")
	v.Check(phase.Time > 0, "time", "must be greater than zero")
	v.Check(phase.Amount >= 0, "amount", "must be greater than or equal to zero")
}

// ValidateV60Phase validates a v60 phase is the correct format
func ValidateV60Phase(v *validator.Validator, phase *V60Phase) {
	v.Check(phase.Time > 0, "time", "must be greater than zero")
	v.Check(phase.Amount >= 0, "amount", "must be greater than or equal to zero")
}

type RecipeModelInterface interface {
	Insert(recipe *Recipe) error
	GetAllForUser(userID int64, params url.Values) ([]*Recipe, error)
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

func (m RecipeModel) GetAllForUser(userID int64, params url.Values) ([]*Recipe, error) {

	// Edge case if 0 is passed into the query parameter
	if params.Get("method_id") == "0" || params.Get("coffee_id") == "0" {
		return []*Recipe{}, nil
	}

	query := `SELECT * FROM recipes WHERE user_id = $1`
	args := []any{userID}
	argIndex := 2

	// Safely parse optional parameters
	methodID, _ := strconv.ParseInt(params.Get("method_id"), 10, 64)
	coffeeID, _ := strconv.ParseInt(params.Get("coffee_id"), 10, 64)

	if methodID != 0 {
		query += fmt.Sprintf(" AND method_id = $%d", argIndex)
		args = append(args, methodID)
		argIndex++
	}

	if coffeeID != 0 {
		query += fmt.Sprintf(" AND coffee_id = $%d", argIndex)
		args = append(args, coffeeID)
		argIndex++
	}

	query += " ORDER BY id"

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, args...)
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
