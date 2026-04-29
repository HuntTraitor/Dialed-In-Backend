package data

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hunttraitor/dialed-in-backend/internal/validator"
)

type Recipe struct {
	ID        int64           `json:"id,omitempty"`
	UserID    int64           `json:"user_id,omitempty"`
	MethodID  int64           `json:"method_id"`
	CoffeeID  *int64          `json:"coffee_id,omitempty"`
	GrinderID *int64          `json:"grinder_id,omitempty"`
	Info      json.RawMessage `json:"info"`
	CreatedAt string          `json:"created_at,omitempty"`
	Version   int             `json:"version,omitempty"`
}

type FullRecipe struct {
	ID        int64           `json:"id"`
	UserID    int64           `json:"user_id"`
	Method    Method          `json:"method"`
	Coffee    *Coffee         `json:"coffee,omitempty"`
	Grinder   *Grinder        `json:"grinder,omitempty"`
	Info      json.RawMessage `json:"info"`
	CreatedAt string          `json:"created_at"`
	Version   int             `json:"version"`
}

type SwitchRecipeInfo struct {
	Name      string        `json:"name"`
	GramIn    int           `json:"grams_in"`
	MlOut     int           `json:"ml_out"`
	WaterTemp string        `json:"water_temp"`
	GrindSize string        `json:"grind_size,omitempty"`
	Phases    []SwitchPhase `json:"phases"`
}

type SwitchPhase struct {
	Open   *bool `json:"open"`
	Time   *int  `json:"time"`
	Amount *int  `json:"amount"`
}

type V60RecipeInfo struct {
	Name      string     `json:"name"`
	GramIn    int        `json:"grams_in"`
	MlOut     int        `json:"ml_out"`
	WaterTemp string     `json:"water_temp"`
	GrindSize string     `json:"grind_size,omitempty"`
	Phases    []V60Phase `json:"phases"`
}

type V60Phase struct {
	Time   *int `json:"time"`
	Amount *int `json:"amount"`
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
func ValidateRecipe(v *validator.Validator, recipe *Recipe) {
	v.Check(recipe.MethodID > 0, "method_id", "must be provided")
	v.Check(recipe.Info != nil, "info", "must be provided")

	switch recipe.MethodID {
	case 1:
		info, ok := decodeInfoStrict[V60RecipeInfo](v, recipe.Info)
		if !ok {
			return
		}
		v.Check(info.Name != "", "name", "must be provided")
		v.Check(len(info.Name) <= 100, "name", "must not be more than 100 bytes")
		v.Check(info.MlOut > 0, "ml_out", "must be greater than zero")
		v.Check(info.GramIn > 0, "grams_in", "must be greater than zero")
		v.Check(info.MlOut < 1000, "ml_out", "must be less than 1000ml")
		v.Check(info.GramIn < 10000, "grams_in", "must be less than 10000g")
		v.Check(len(info.Phases) > 0, "phases", "must be greater than zero")
		v.Check(len(info.GrindSize) <= 50, "grind_size", "must not be more than 50 bytes")
		v.Check(validator.Matches(info.WaterTemp, validator.TempRX), "water_temp", "must be an int ending in either °C or °F")

		for _, phase := range info.Phases {
			ValidateV60Phase(v, &phase)
		}

	case 2:
		info, ok := decodeInfoStrict[SwitchRecipeInfo](v, recipe.Info)
		if !ok {
			return
		}
		v.Check(info.Name != "", "name", "must be provided")
		v.Check(len(info.Name) <= 100, "name", "must not be more than 100 bytes")
		v.Check(info.MlOut > 0, "ml_out", "must be greater than zero")
		v.Check(info.GramIn > 0, "grams_in", "must be greater than zero")
		v.Check(info.MlOut < 1000, "ml_out", "must be less than 1000ml")
		v.Check(info.GramIn < 10000, "grams_in", "must be less than 10000g")
		v.Check(len(info.Phases) > 0, "phases", "must be greater than zero")
		v.Check(len(info.GrindSize) <= 50, "grind_size", "must not be more than 50 bytes")
		v.Check(validator.Matches(info.WaterTemp, validator.TempRX), "water_temp", "must be an int ending in either °C or °F")

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
	v.Check(phase.Time != nil, "time", "must be provided")
	v.Check(phase.Amount != nil, "amount", "must be provided")
	if phase.Time != nil {
		v.Check(*phase.Time > 0, "time", "must be greater than zero")
		v.Check(*phase.Time < 1000, "time", "must be less than 10000")
	}
	if phase.Amount != nil {
		v.Check(*phase.Amount >= 0, "amount", "must be greater than or equal to zero")
		v.Check(*phase.Amount < 1000, "amount", "must be less than 10000")
	}
}

// ValidateV60Phase validates a v60 phase is the correct format
func ValidateV60Phase(v *validator.Validator, phase *V60Phase) {
	v.Check(phase.Time != nil, "time", "must be provided")
	v.Check(phase.Amount != nil, "amount", "must be provided")
	if phase.Time != nil {
		v.Check(*phase.Time > 0, "time", "must be greater than zero")
		v.Check(*phase.Time < 1000, "time", "must be less than 10000")
	}
	if phase.Amount != nil {
		v.Check(*phase.Amount >= 0, "amount", "must be greater than or equal to zero")
		v.Check(*phase.Amount < 1000, "amount", "must be less than 10000")
	}
}

type RecipeModelInterface interface {
	Insert(recipe *Recipe) error
	GetAllForUser(userID int64, filters RecipeFilters) ([]*Recipe, MetaData, error)
	Get(id int64, userId int64) (*Recipe, error)
	Update(recipe *Recipe) error
	Delete(id int64, userID int64) error
}

func (m RecipeModel) Insert(recipe *Recipe) error {
	query := `INSERT INTO recipes (user_id, method_id, coffee_id, grinder_id, info) VALUES ($1, $2, $3, $4, $5) 
            RETURNING id, created_at, version`

	infoJSON, err := json.Marshal(recipe.Info)
	if err != nil {
		return err
	}

	args := []any{recipe.UserID, recipe.MethodID, recipe.CoffeeID, recipe.GrinderID, infoJSON}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&recipe.ID, &recipe.CreatedAt, &recipe.Version)
	if err != nil {
		return err
	}
	return nil
}

func (m RecipeModel) GetAllForUser(userID int64, filters RecipeFilters) ([]*Recipe, MetaData, error) {
	query := fmt.Sprintf(`
		SELECT count(*) OVER(),
			id,
			user_id,
			coffee_id,
			method_id,
			grinder_id,
			info,
			version,
			created_at
		FROM recipes 
        WHERE user_id = $1
        
        -- ID
		AND (method_id::int = $2 OR $2 = 0)
		AND (coffee_id::int = $3 OR $3 = 0)
		AND (grinder_id::int = $4 OR $4 = 0)
        
    	-- General Search
		AND (
		  $5 = ''
		  OR info->>'name'    ILIKE '%%' || $5 || '%%'
		)
		
		-- Text search
		AND (info->>'name'    ILIKE '%%' || $6 || '%%' OR $6 = '')
		AND (info->>'water_temp' ILIKE '%%' || $7 || '%%' OR $7 = '')
		AND (info->>'grind_size'  ILIKE '%%' || $8 || '%%' OR $8 = '')
        
        -- JSONB int filters
		AND ((info->>'grams_in')::int = $9 OR $9 = 0)
		AND ((info->>'ml_out')::int = $10 OR $10 = 0)
		
		ORDER BY %s %s, id ASC
		LIMIT $11 OFFSET $12;`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	fmt.Println(filters)

	args := []any{
		userID,            // $1
		filters.MethodID,  // $2
		filters.CoffeeID,  // $3
		filters.GrinderID, // $4
		filters.Search,    // $5
		filters.Name,      // $6
		filters.WaterTemp, // $7
		filters.GrindSize, // $8
		filters.GramsIn,   // $9
		filters.MlOut,     // $10
		filters.limit(),   // $11
		filters.offset(),  // $12
	}

	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, MetaData{}, err
	}

	defer rows.Close()

	totalRecords := 0
	recipes := []*Recipe{}
	var infoBytes []byte

	for rows.Next() {
		var recipe Recipe
		err = rows.Scan(
			&totalRecords,
			&recipe.ID,
			&recipe.UserID,
			&recipe.CoffeeID,
			&recipe.MethodID,
			&recipe.GrinderID,
			&infoBytes,
			&recipe.Version,
			&recipe.CreatedAt,
		)
		if err != nil {
			return nil, MetaData{}, err
		}
		err = json.Unmarshal(infoBytes, &recipe.Info)
		if err != nil {
			return nil, MetaData{}, err
		}
		recipes = append(recipes, &recipe)
	}

	if err = rows.Err(); err != nil {
		return nil, MetaData{}, err
	}
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return recipes, metadata, nil
}

func (m RecipeModel) Update(recipe *Recipe) error {
	query := `
		UPDATE recipes
		SET method_id = $1, 
		    coffee_id = $2, 
		    grinder_id = $3,
		    info = $4,
		    version = version + 1
		WHERE id = $5 and version = $6
		returning version
	`

	infoJSON, err := json.Marshal(recipe.Info)
	if err != nil {
		return err
	}

	args := []any{
		recipe.MethodID,
		recipe.CoffeeID,
		recipe.GrinderID,
		infoJSON,
		recipe.ID,
		recipe.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&recipe.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m RecipeModel) Get(id int64, userId int64) (*Recipe, error) {
	if id < 1 || userId < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT
				id,
				user_id,
				coffee_id,
				method_id,
				grinder_id,
				info,
				version,
				created_at
			FROM recipes
			WHERE id = $1 and user_id = $2
					`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var recipe Recipe
	var infoBytes []byte

	err := m.DB.QueryRowContext(ctx, query, id, userId).Scan(
		&recipe.ID,
		&recipe.UserID,
		&recipe.CoffeeID,
		&recipe.MethodID,
		&recipe.GrinderID,
		&infoBytes,
		&recipe.Version,
		&recipe.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	err = json.Unmarshal(infoBytes, &recipe.Info)
	if err != nil {
		return nil, err
	}
	return &recipe, nil
}

func (m RecipeModel) Delete(id int64, userID int64) error {
	if id < 1 || userID < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM recipes WHERE id = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
