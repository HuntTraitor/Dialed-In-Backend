package model

import (
	"github.com/google/uuid"
	"time"
)

type Account struct {
	AccountId uuid.UUID `json:"account_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ReturnedAccount struct {
	AccountId uuid.UUID `json:"account_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
