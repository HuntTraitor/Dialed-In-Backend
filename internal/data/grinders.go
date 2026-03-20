package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/hunttraitor/dialed-in-backend/internal/validator"
)

type Grinder struct {
	ID        int64  `json:"id,omitempty"`
	UserId    int64  `json:"user_id,omitempty"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at,omitempty"`
	Version   int64  `json:"version,omitempty"`
}

type GrinderModel struct {
	DB *sql.DB
}

type GrinderModelInterface interface {
	GetAllForUser(userId int64) ([]*Grinder, error)
	Insert(grinder *Grinder) error
	GetOne(id int64, userId int64) (*Grinder, error)
	Update(grinder *Grinder) error
	Delete(id int64, userId int64) error
}

func ValidateGrinder(v *validator.Validator, grinder *Grinder) {
	v.Check(grinder.Name != "", "name", "must be provided")
	v.Check(len(grinder.Name) <= 500, "name", "must not be more than 500 bytes long")
}

func (m GrinderModel) GetOne(id int64, userId int64) (*Grinder, error) {
	if id < 1 || userId < 1 {
		return nil, ErrRecordNotFound
	}
	query := `SELECT * FROM grinders WHERE id = $1 AND user_id = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var grinder Grinder

	row := m.DB.QueryRowContext(ctx, query, id, userId)
	err := row.Scan(
		&grinder.ID,
		&grinder.UserId,
		&grinder.Name,
		&grinder.Version,
		&grinder.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &grinder, nil
}

func (m GrinderModel) GetAllForUser(userId int64) ([]*Grinder, error) {
	query := `SELECT * FROM grinders WHERE user_id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}

	grinders := []*Grinder{}

	for rows.Next() {
		var grinder Grinder
		err = rows.Scan(
			&grinder.ID,
			&grinder.UserId,
			&grinder.Name,
			&grinder.Version,
			&grinder.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		grinders = append(grinders, &grinder)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return grinders, nil
}

func (m GrinderModel) Insert(grinder *Grinder) error {
	query := `INSERT INTO grinders (user_id, name) VALUES ($1, $2) RETURNING id, created_at, version`

	args := []any{grinder.UserId, grinder.Name}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&grinder.ID, &grinder.CreatedAt, &grinder.Version)
	if err != nil {
		return err
	}
	return nil
}

func (m GrinderModel) Update(grinder *Grinder) error {
	query := `    UPDATE grinders
    SET name = $1, version = version + 1
    WHERE id = $2 AND version = $3
    RETURNING version`

	args := []any{grinder.Name, grinder.ID, grinder.Version}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&grinder.Version)
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

func (m GrinderModel) Delete(id int64, userId int64) error {
	if id < 1 || userId < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM grinders WHERE id = $1 AND user_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.DB.ExecContext(ctx, query, id, userId)
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
