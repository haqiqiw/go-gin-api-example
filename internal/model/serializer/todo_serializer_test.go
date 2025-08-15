package serializer_test

import (
	"go-api-example/internal/entity"
	"go-api-example/internal/model"
	"go-api-example/internal/model/serializer"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTodoSerializer_TodoToResponse(t *testing.T) {
	now := time.Date(2025, 8, 13, 10, 0, 0, 0, time.UTC)
	description := "dummy description"

	tests := []struct {
		name    string
		param   *entity.Todo
		wantRes *model.TodoResponse
	}{
		{
			name: "success with description",
			param: &entity.Todo{
				ID:          1,
				UserID:      1,
				Title:       "dummy title",
				Description: &description,
				Status:      entity.TodoStatusPending,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			wantRes: &model.TodoResponse{
				ID:          1,
				UserID:      1,
				Title:       "dummy title",
				Description: description,
				Status:      "pending",
				CreatedAt:   now.Format(time.RFC3339),
				UpdatedAt:   now.Format(time.RFC3339),
			},
		},
		{
			name: "success without description",
			param: &entity.Todo{
				ID:          2,
				UserID:      1,
				Title:       "dummy title",
				Description: nil,
				Status:      entity.TodoStatusCompleted,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			wantRes: &model.TodoResponse{
				ID:          2,
				UserID:      1,
				Title:       "dummy title",
				Description: "",
				Status:      "completed",
				CreatedAt:   now.Format(time.RFC3339),
				UpdatedAt:   now.Format(time.RFC3339),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := serializer.TodoToResponse(tt.param)

			assert.Equal(t, tt.wantRes, res)
		})
	}
}

func TestTodoSerializer_ListTodoToResponse(t *testing.T) {
	now := time.Date(2025, 8, 13, 10, 0, 0, 0, time.UTC)
	description := "dummy description"

	tests := []struct {
		name    string
		param   []entity.Todo
		wantRes []model.TodoResponse
	}{
		{
			name: "success",
			param: []entity.Todo{
				{
					ID:          1,
					UserID:      1,
					Title:       "dummy title",
					Description: &description,
					Status:      entity.TodoStatusPending,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
				{
					ID:          2,
					UserID:      1,
					Title:       "dummy title",
					Description: nil,
					Status:      entity.TodoStatusCompleted,
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			},
			wantRes: []model.TodoResponse{
				{
					ID:          1,
					UserID:      1,
					Title:       "dummy title",
					Description: description,
					Status:      "pending",
					CreatedAt:   now.Format(time.RFC3339),
					UpdatedAt:   now.Format(time.RFC3339),
				},
				{
					ID:          2,
					UserID:      1,
					Title:       "dummy title",
					Description: "",
					Status:      "completed",
					CreatedAt:   now.Format(time.RFC3339),
					UpdatedAt:   now.Format(time.RFC3339),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := serializer.ListTodoToResponse(tt.param)

			assert.Equal(t, tt.wantRes, res)
		})
	}
}
