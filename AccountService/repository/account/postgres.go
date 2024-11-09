package account

import (
	"context"
	"fmt"
	"github.com/hunttraitor/dialed-in-backend/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Db struct {
	Pool *pgxpool.Pool
}

func (db *Db) Insert(ctx context.Context, account model.Account) (model.ReturnedAccount, error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return model.ReturnedAccount{}, fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	query := `
		INSERT INTO accounts (name, email, password) 
		VALUES ($1, $2, crypt($3, '87'))
		RETURNING id, name, email, created_at, updated_at
	`

	var newAccount model.ReturnedAccount
	err = tx.QueryRow(ctx, query, account.Name, account.Email, account.Password).Scan(
		&newAccount.AccountId,
		&newAccount.Name,
		&newAccount.Email,
		&newAccount.CreatedAt,
		&newAccount.UpdatedAt,
	)
	if err != nil {
		return model.ReturnedAccount{}, fmt.Errorf("could not execute insert query: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return model.ReturnedAccount{}, fmt.Errorf("could not commit transaction: %w", err)
	}
	return newAccount, nil
}
