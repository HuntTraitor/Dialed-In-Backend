package handler

import (
	"encoding/json"
	"fmt"
	"github.com/hunttraitor/dialed-in-backend/errors"
	"github.com/hunttraitor/dialed-in-backend/model"
	"github.com/hunttraitor/dialed-in-backend/repository/account"
	"net/http"
)

type Account struct {
	Db *account.Db
}

func (h *Account) Create(w http.ResponseWriter, r *http.Request) error {
	var body struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return errors.InvalidJSON()
	}

	account := model.Account{
		Name:     body.Name,
		Email:    body.Email,
		Password: body.Password,
	}

	newAccount, err := h.Db.Insert(r.Context(), account)
	if err != nil && errors.IsPgErrorCode(err, errors.UniqueViolationErr) {
		return errors.DuplicateEmail()
	} else if err != nil {
		return err
	}

	res, err := json.Marshal(newAccount)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(res)
	if err != nil {
		return err
	}
	return nil
}

func (h *Account) List(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Account lists...")
}

func (h *Account) GetById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Returns an item by its id")
}

func (h *Account) Update(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Updates an account")
}

func (h *Account) Delete(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Deletes an account")
}

func (h *Account) Login(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Login an account")
}
