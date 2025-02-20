package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"time"
)

type Coffee struct {
	ID          int    `json:"id"`
	UserID      int    `json:"user_id"`
	Name        string `json:"name"`
	Region      string `json:"region"`
	Process     string `json:"process"`
	Img         string `json:"img"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
	Version     int    `json:"version"`
}

type CoffeeModel struct {
	DB *sql.DB
}

type CoffeeModelInterface interface {
	GetAllForUser(userID int64) ([]*Coffee, error)
	Insert(userID int64, coffee *Coffee) (*Coffee, error)
	GetOne(id int64, userId int64) (*Coffee, error)
	Update(coffee *Coffee) error
	Delete(id int64, userID int64) error
}

func ValidateCoffee(v *validator.Validator, coffee *Coffee) {
	v.Check(coffee.Name != "", "name", "must be provided")
	v.Check(coffee.Description != "", "description", "must be provided")
	v.Check(coffee.Region != "", "region", "must be provided")
	v.Check(coffee.Process != "", "process", "must be provided")
	v.Check(len(coffee.Name) <= 500, "name", "must not be more than 500 bytes long")
	v.Check(len(coffee.Description) <= 1000, "description", "must not be more than 1000 bytes long")
	v.Check(len(coffee.Region) <= 100, "region", "must not be more than 100 bytes long")
	v.Check(len(coffee.Process) <= 200, "process", "must not be more than 200 bytes long")
	v.Check(len(coffee.Img) <= 8192, "img", "must not be more than 8192 bytes long")
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
			&coffee.Version,
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
	query := `INSERT INTO coffees (user_id, name, region, process, img, description) VALUES ($1, $2, $3, $4, $5, $6) RETURNING *`

	args := []any{userID, coffee.Name, coffee.Region, coffee.Process, coffee.Img, coffee.Description}

	var returnedCoffee Coffee

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&returnedCoffee.ID,
		&returnedCoffee.UserID,
		&returnedCoffee.CreatedAt,
		&returnedCoffee.Name,
		&returnedCoffee.Region,
		&returnedCoffee.Process,
		&returnedCoffee.Img,
		&returnedCoffee.Description,
		&returnedCoffee.Version)

	if err != nil {
		return nil, err
	}
	return &returnedCoffee, nil
}

func (m CoffeeModel) GetOne(id int64, userId int64) (*Coffee, error) {

	if id < 1 || userId < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT * FROM coffees WHERE id = $1 AND user_id = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	var coffee Coffee
	err := m.DB.QueryRowContext(ctx, query, id, userId).Scan(
		&coffee.ID,
		&coffee.UserID,
		&coffee.CreatedAt,
		&coffee.Name,
		&coffee.Region,
		&coffee.Img,
		&coffee.Description,
		&coffee.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &coffee, nil
}

func (m CoffeeModel) Update(coffee *Coffee) error {
	query := `UPDATE coffees
						SET name = $1, region = $2, img = $3, description = $4, version = version + 1
						WHERE coffees.id = $5 AND version = $6
						RETURNING version`

	args := []any{
		coffee.Name,
		coffee.Region,
		coffee.Img,
		coffee.Description,
		coffee.ID,
		coffee.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&coffee.Version)
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

func (m CoffeeModel) Delete(id int64, userID int64) error {
	if id < 1 || userID < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM coffees WHERE id = $1 AND user_id = $2`

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
