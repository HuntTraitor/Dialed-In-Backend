package data

import (
	"context"
	"database/sql"
	"github.com/aws/aws-sdk-go/service/s3"
	"time"
)

type Method struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Img       string `json:"img"`
	CreatedAt string `json:"created_at"`
}

type MethodModel struct {
	DB *sql.DB
	s3 *s3.S3
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

	methods := []*Method{}

	for rows.Next() {
		var method Method

		err = rows.Scan(
			&method.ID,
			&method.CreatedAt,
			&method.Name,
			&method.Img,
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
