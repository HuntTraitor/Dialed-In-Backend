package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Users UserModelInterface
}

type UserModel struct {
	DB *sql.DB
}

type UserModelInterface interface {
	Insert(user *User) error
}

// NewModels returns models associated with a real database
func NewModels(db *sql.DB) Models {
	return Models{
		Users: UserModel{DB: db},
	}
}
