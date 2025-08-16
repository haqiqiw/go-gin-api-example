package db_test

import (
	"context"
	"errors"
	"go-api-example/internal/db"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestTransactioner_Do(t *testing.T) {
	tests := []struct {
		name       string
		mockFunc   func(sqlmock.Sqlmock)
		panic      bool
		wantErrMsg string
	}{
		{
			name: "success",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM users`,
				)).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
				m.ExpectCommit()
			},
			panic:      false,
			wantErrMsg: "",
		},
		{
			name: "error",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM users`,
				)).
					WillReturnError(errors.New("something error"))
				m.ExpectRollback()
			},
			panic:      false,
			wantErrMsg: "something error",
		},
		{
			name: "panic",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectBegin()
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM users`,
				)).
					WillReturnError(errors.New("something error"))
				m.ExpectRollback()
			},
			panic:      true,
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbMock, mock, _ := sqlmock.New()
			defer dbMock.Close()

			ctx := context.Background()
			trx := db.NewTransactioner(dbMock)
			tt.mockFunc(mock)

			if tt.panic {
				assert.Panics(t, func() {
					_ = trx.Do(ctx, func(exec db.Executor) error {
						panic("something error")
					})
				})
			} else {
				err := trx.Do(ctx, func(exec db.Executor) error {
					var count int
					return exec.QueryRowContext(context.Background(),
						"SELECT COUNT(id) FROM users").Scan(&count)
				})

				if tt.wantErrMsg != "" {
					assert.Equal(t, tt.wantErrMsg, err.Error())
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}
