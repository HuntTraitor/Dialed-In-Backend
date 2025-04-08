package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Method struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type MethodModel struct {
	DB *sql.DB
}

type MethodModelInterface interface {
	GetAll() ([]*Method, error)
	GetOne(id int64) (*Method, error)
}

func (m MethodModel) GetAll() ([]*Method, error) {
	query := `SELECT * FROM methods`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	methods := []*Method{}

	for rows.Next() {
		var method Method

		err = rows.Scan(
			&method.ID,
			&method.CreatedAt,
			&method.Name,
		)
		if err != nil {
			return nil, err
		}
		methods = append(methods, &method)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return methods, nil
}

func (m MethodModel) GetOne(id int64) (*Method, error) {
	query := `SELECT * FROM methods WHERE id = $1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	row := m.DB.QueryRowContext(ctx, query, id)
	var method Method
	err := row.Scan(
		&method.ID,
		&method.CreatedAt,
		&method.Name,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &method, nil
}
