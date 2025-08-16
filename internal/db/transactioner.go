package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type Transactioner interface {
	Do(ctx context.Context, fn func(Executor) error) error
}

type transactioner struct {
	DB *sql.DB
}

func NewTransactioner(db *sql.DB) Transactioner {
	return &transactioner{
		DB: db,
	}
}

func (h *transactioner) Do(ctx context.Context, fn func(Executor) error) (err error) {
	tx, err := h.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction error: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("panic occured and rollback error: %w", rbErr)
			}
			panic(p)
		}

		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("rollback error: %w", rbErr)
			}
			return
		}

		if cmErr := tx.Commit(); cmErr != nil {
			err = fmt.Errorf("commit error: %w", cmErr)
		}
	}()

	err = fn(tx)

	return err
}
