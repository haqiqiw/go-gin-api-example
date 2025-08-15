package model_test

import (
	"go-api-example/internal/model"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCustomError_MessageFor(t *testing.T) {
	tests := []struct {
		name    string
		param   int
		wantMsg string
	}{
		{
			name:    "success",
			param:   model.ErrUsernameAlreadyExist,
			wantMsg: "Username already exist",
		},
		{
			name:    "not found",
			param:   9999,
			wantMsg: "Unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := model.MessageFor(tt.param)

			assert.Equal(t, tt.wantMsg, msg)
		})
	}
}

func TestCustomError_Error(t *testing.T) {
	tests := []struct {
		name      string
		customErr *model.CustomError
		wantMsg   string
	}{
		{
			name: "single error",
			customErr: model.NewCustomError(
				http.StatusBadRequest,
				model.ErrUsernameAlreadyExist,
			),
			wantMsg: "Username already exist",
		},
		{
			name: "multiple errors",
			customErr: model.NewCustomError(
				http.StatusBadRequest,
				model.ErrInvalidUserID,
				model.ErrInternalServerError,
			),
			wantMsg: "Invalid user id",
		},
		{
			name: "empty errors",
			customErr: &model.CustomError{
				HTTPStatus: http.StatusBadRequest,
				Errors:     []model.ErrorItem{},
			},
			wantMsg: "No errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.customErr.Error()

			assert.Equal(t, tt.wantMsg, msg)
		})
	}
}
