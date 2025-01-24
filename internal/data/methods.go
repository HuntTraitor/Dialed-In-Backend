package data

import (
	"context"
	"database/sql"
	"time"
)

type Method struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Img       string    `json:"img"`
	CreatedAt time.Time `json:"created_at"`
}

type MethodModel struct {
	DB *sql.DB
}

type MethodModelInterface interface {
	GetAll() ([]*Method, error)
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

	var methods []*Method

	for rows.Next() {
		var method Method

		err = rows.Scan(
			&method.ID,
			&method.Name,
			&method.Img,
			&method.CreatedAt,
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
