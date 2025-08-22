package model_test

import (
	"go-api-example/internal/model"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomError_Error(t *testing.T) {
	tests := []struct {
		name      string
		customErr *model.CustomError
		wantMsg   string
	}{
		{
			name:      "single error",
			customErr: model.ErrUsernameAlreadyExist,
			wantMsg:   "username already exist",
		},
		{
			name: "multiple errors",
			customErr: func() *model.CustomError {
				err := model.ErrInvalidUserID
				err.Append(model.ErrorItem{
					Code:    9999,
					Message: "another error",
				})
				return err
			}(),
			wantMsg: "invalid user id",
		},
		{
			name: "empty errors",
			customErr: &model.CustomError{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []model.ErrorItem{},
			},
			wantMsg: "empty error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.customErr.Error()

			assert.Equal(t, tt.wantMsg, msg)
		})
	}
}
