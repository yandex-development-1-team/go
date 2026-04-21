package postgres

import (
	"context"

	"github.com/jmoiron/sqlx"

	"github.com/yandex-development-1-team/go/internal/ctxutil"
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
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	txCtx := ctxutil.WithTx(ctx, tx)

	if err = fn(txCtx); err != nil {
		return err
	}

	return tx.Commit()
}

func (u *TxRepo) BeginTx(ctx context.Context) (*sqlx.Tx, error) {
	return u.db.BeginTxx(ctx, nil)
}
