package serializer_test

import (
	"go-api-example/internal/entity"
	"go-api-example/internal/model"
	"go-api-example/internal/model/serializer"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUserSerializer_UserToResponse(t *testing.T) {
	now := time.Date(2025, 8, 13, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		param   *entity.User
		wantRes *model.UserResponse
	}{
		{
			name: "success",
			param: &entity.User{
				ID:        1,
				Username:  "johndoe",
				Password:  "password",
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantRes: &model.UserResponse{
				ID:        1,
				Username:  "johndoe",
				CreatedAt: now.Format(time.RFC3339),
				UpdatedAt: now.Format(time.RFC3339),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := serializer.UserToResponse(tt.param)

			assert.Equal(t, tt.wantRes, res)
		})
	}
}

func TestUserSerializer_ListUserToResponse(t *testing.T) {
	now := time.Date(2025, 8, 13, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		param   []entity.User
		wantRes []model.UserResponse
	}{
		{
			name: "success",
			param: []entity.User{
				{
					ID:        1,
					Username:  "johndoe",
					Password:  "password",
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        2,
					Username:  "chyntia",
					Password:  "password",
					CreatedAt: now,
					UpdatedAt: now,
				},
			},
			wantRes: []model.UserResponse{
				{

					ID:        1,
					Username:  "johndoe",
					CreatedAt: now.Format(time.RFC3339),
					UpdatedAt: now.Format(time.RFC3339),
				},
				{

					ID:        2,
					Username:  "chyntia",
					CreatedAt: now.Format(time.RFC3339),
					UpdatedAt: now.Format(time.RFC3339),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := serializer.ListUserToResponse(tt.param)

			assert.Equal(t, tt.wantRes, res)
		})
	}
}
