package errors

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
)

const (
	UniqueViolationErr = "23505"
)

func IsPgErrorCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == code
	}
	return false
}
