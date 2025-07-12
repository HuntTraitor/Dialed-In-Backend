package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hunttraitor/dialed-in-backend/internal/validator"
	"time"
)

type Coffee struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	Info      CoffeeInfo `json:"info"`
	CreatedAt string     `json:"created_at"`
	Version   int        `json:"version"`
}

type CoffeeInfo struct {
	Name         string   `json:"name"`
	Roaster      string   `json:"roaster"`
	Region       string   `json:"region"`
	Process      string   `json:"process"`
	Decaf        bool     `json:"decaf"`
	OriginType   string   `json:"origin_type"`
	Rating       int      `json:"rating"`
	TastingNotes []string `json:"tasting_notes"`
	RoastLevel   string   `json:"roast_level"`
	Cost         float32  `json:"cost"`
	Img          string   `json:"img"`
	Description  string   `json:"description"`
}

type CoffeeModel struct {
	DB *sql.DB
}

type CoffeeModelInterface interface {
	GetAllForUser(userID int64) ([]*Coffee, error)
	Insert(coffee *Coffee) error
	GetOne(id int64, userId int64) (*Coffee, error)
	Update(coffee *Coffee) error
	Delete(id int64, userID int64) error
}

func ValidateCoffee(v *validator.Validator, coffee *Coffee) {
	v.Check(coffee.Info.Name != "", "name", "must be provided")
	v.Check(len(coffee.Info.Name) <= 500, "name", "must not be more than 500 bytes long")
	v.Check(len(coffee.Info.Roaster) <= 500, "roaster", "must not be more than 500 bytes long")
	v.Check(len(coffee.Info.Description) <= 1000, "description", "must not be more than 1000 bytes long")
	v.Check(len(coffee.Info.Region) <= 100, "region", "must not be more than 100 bytes long")
	v.Check(len(coffee.Info.Process) <= 200, "process", "must not be more than 200 bytes long")
	v.Check(len(coffee.Info.OriginType) <= 100, "origin_type", "must not be more than 100 bytes long")
	v.Check(coffee.Info.Rating >= 0 && coffee.Info.Rating <= 5, "rating", "must be between 0 and 5")
	v.Check(len(coffee.Info.TastingNotes) <= 50, "tasting_notes", "must not contain more than 50 entries")
	v.Check(len(coffee.Info.RoastLevel) <= 100, "roast_level", "must not be more than 100 bytes long")
	v.Check(coffee.Info.Cost <= 1_000_000, "cost", "must not be more than 1,000,000")
	for i, note := range coffee.Info.TastingNotes {
		field := fmt.Sprintf("tasting_notes[%d]", i)
		v.Check(len(note) <= 100, field, "must not be more than 100 bytes long")
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
	var infoBytes []byte

	for rows.Next() {
		var coffee Coffee
		err = rows.Scan(
			&coffee.ID,
			&coffee.UserID,
			&infoBytes,
			&coffee.Version,
			&coffee.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(infoBytes, &coffee.Info)
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

func (m CoffeeModel) Insert(coffee *Coffee) error {
	query := `INSERT INTO coffees (user_id, info) VALUES ($1, $2) RETURNING id, created_at, version`

	infoJSON, err := json.Marshal(coffee.Info)
	if err != nil {
		return err
	}

	args := []any{coffee.UserID, infoJSON}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&coffee.ID, &coffee.CreatedAt, &coffee.Version)
	if err != nil {
		return err
	}
	return nil
}

func (m CoffeeModel) GetOne(id int64, userId int64) (*Coffee, error) {
	if id < 1 || userId < 1 {
		return nil, ErrRecordNotFound
	}

	query := `SELECT * FROM coffees WHERE id = $1 AND user_id = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var coffee Coffee
	var infoBytes []byte

	err := m.DB.QueryRowContext(ctx, query, id, userId).Scan(
		&coffee.ID,
		&coffee.UserID,
		&infoBytes,
		&coffee.Version,
		&coffee.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	err = json.Unmarshal(infoBytes, &coffee.Info)
	if err != nil {
		return nil, err
	}

	return &coffee, nil
}

func (m CoffeeModel) Update(coffee *Coffee) error {
	query := `
    UPDATE coffees
    SET info = $1, version = version + 1
    WHERE id = $2 AND version = $3
    RETURNING version`

	infoJSON, err := json.Marshal(coffee.Info)
	if err != nil {
		return err
	}

	args := []any{
		infoJSON,
		coffee.ID,
		coffee.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = m.DB.QueryRowContext(ctx, query, args...).Scan(&coffee.Version)
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
