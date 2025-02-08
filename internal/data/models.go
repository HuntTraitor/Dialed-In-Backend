package data

import (
	"database/sql"
	"errors"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Users   UserModelInterface
	Tokens  TokenModelInterface
	Methods MethodModelInterface
	Coffees CoffeeModelInterface
}

// NewModels returns models associated with a real database
func NewModels(db *sql.DB, s3 *s3.S3) Models {
	return Models{
		Users:   UserModel{DB: db},
		Tokens:  TokenModel{DB: db},
		Methods: MethodModel{DB: db, s3: s3},
		Coffees: CoffeeModel{DB: db, s3: s3},
	}
}
