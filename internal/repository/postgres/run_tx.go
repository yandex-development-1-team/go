package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type TxRepo struct {
	db *sqlx.DB
}

func NewTxRepo(db *sqlx.DB) *TxRepo {
	return &TxRepo{
		db: db,
	}
}

func (u *TxRepo) RunToTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	if err := fn(ctx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
