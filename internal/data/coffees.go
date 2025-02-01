package data

import (
	"context"
	"database/sql"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"time"
)

type Coffee struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	Name        string `json:"name"`
	Region      string `json:"region"`
	Img         string `json:"img"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type CoffeeModel struct {
	DB *sql.DB
}

type CoffeeModelInterface interface {
	GetAllForUser(userID int64) ([]*Coffee, error)
	Insert(userID int64, coffee *Coffee) (*Coffee, error)
}

func ValidateCoffee(v *validator.Validator, coffee *Coffee) {
	v.Check(coffee.Name != "", "name", "must be provided")
	v.Check(coffee.Description != "", "description", "must be provided")
	v.Check(coffee.Region != "", "region", "must be provided")
	v.Check(len(coffee.Name) <= 500, "name", "must not be more than 500 bytes long")
	v.Check(len(coffee.Description) <= 1000, "description", "must not be more than 500 bytes long")
	v.Check(len(coffee.Region) <= 100, "region", "must not be more than 100 bytes long")
	if coffee.Img != "" {
		v.Check(validator.Matches(coffee.Img, validator.UrlRX), "img", "must be a valid image URL")
	}
}

func (m CoffeeModel) GetAllForUser(userID int64) ([]*Coffee, error) {
	query := `SELECT * FROM coffees WHERE user_id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	rows, err := m.DB.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	coffees := []*Coffee{}

	for rows.Next() {
		var coffee Coffee

		err = rows.Scan(
			&coffee.ID,
			&coffee.UserID,
			&coffee.CreatedAt,
			&coffee.Name,
			&coffee.Region,
			&coffee.Img,
			&coffee.Description,
		)
		if err != nil {
			return nil, err
		}
		coffees = append(coffees, &coffee)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return coffees, nil
}

func (m CoffeeModel) Insert(userID int64, coffee *Coffee) (*Coffee, error) {
	query := `INSERT INTO coffees (user_id, name, region, img, description) VALUES ($1, $2, $3, $4, $5) RETURNING *`

	args := []any{userID, coffee.Name, coffee.Region, coffee.Img, coffee.Description}

	var returnedCoffee Coffee

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&returnedCoffee.ID,
		&returnedCoffee.UserID,
		&returnedCoffee.CreatedAt,
		&returnedCoffee.Name,
		&returnedCoffee.Region,
		&returnedCoffee.Img,
		&returnedCoffee.Description)

	if err != nil {
		return nil, err
	}
	return &returnedCoffee, nil
}
