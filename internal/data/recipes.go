package data

import "database/sql"

type Recipe struct {
	ID       int        `json:"id"`
	UserID   int        `json:"user_id"`
	MethodID int        `json:"method_id"`
	CoffeeID int        `json:"coffee_id"`
	Info     RecipeInfo `json:"info"`
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

type RecipeModelInterface interface{}
