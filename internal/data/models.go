package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Users   UserModelInterface
	Tokens  TokenModelInterface
	Methods MethodModelInterface
	Coffees CoffeeModelInterface
	Recipes RecipeModelInterface
}

// NewModels returns models associated with a real database
func NewModels(db *sql.DB) Models {
	return Models{
		Users:   UserModel{DB: db},
		Tokens:  TokenModel{DB: db},
		Methods: MethodModel{DB: db},
		Coffees: CoffeeModel{DB: db},
		Recipes: RecipeModel{DB: db},
	}
}
