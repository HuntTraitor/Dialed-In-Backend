package data

import (
	"database/sql"
	"errors"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
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

type mockS3Client struct {
	s3iface.S3API
}

// NewModels returns models associated with a real database
func NewModels(db *sql.DB, s3 *s3iface.S3API) Models {
	return Models{
		Users:   UserModel{DB: db},
		Tokens:  TokenModel{DB: db},
		Methods: MethodModel{DB: db, s3: s3},
		Coffees: CoffeeModel{DB: db, s3: s3},
	}
}
