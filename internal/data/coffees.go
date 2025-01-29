package data

import (
	"context"
	"database/sql"
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
