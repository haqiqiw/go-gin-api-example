package entity_test

import (
	"go-api-example/internal/entity"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTodo_GetDescription(t *testing.T) {
	description := "dummy description"

	tests := []struct {
		name    string
		model   *entity.Todo
		wantRes string
	}{
		{
			name:    "nil model",
			model:   nil,
			wantRes: "",
		},
		{
			name: "nil description",
			model: &entity.Todo{
				Description: nil,
			},
			wantRes: "",
		},
		{
			name: "success",
			model: &entity.Todo{
				Description: &description,
			},
			wantRes: description,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.model.GetDescription()

			assert.Equal(t, tt.wantRes, res)
		})
	}
}

func TestTodoStatus_String(t *testing.T) {
	tests := []struct {
		name    string
		status  entity.TodoStatus
		wantRes string
	}{
		{
			name:    "pending status",
			status:  entity.TodoStatusPending,
			wantRes: "pending",
		},
		{
			name:    "in progress status",
			status:  entity.TodoStatusInProgress,
			wantRes: "in_progress",
		},
		{
			name:    "completed status",
			status:  entity.TodoStatusCompleted,
			wantRes: "completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.status.String()

			assert.Equal(t, tt.wantRes, res)
		})
	}
}

func TestTodoStatus_ParseTodoStatus(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		wantRes    entity.TodoStatus
		wantErrMsg string
	}{
		{
			name:       "pending status",
			status:     "pending",
			wantRes:    entity.TodoStatusPending,
			wantErrMsg: "",
		},
		{
			name:       "in progress status",
			status:     "in_progress",
			wantRes:    entity.TodoStatusInProgress,
			wantErrMsg: "",
		},
		{
			name:       "completed status",
			status:     "completed",
			wantRes:    entity.TodoStatusCompleted,
			wantErrMsg: "",
		},
		{
			name:       "unknown status",
			status:     "unknown",
			wantRes:    0,
			wantErrMsg: "invalid status: unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := entity.ParseTodoStatus(tt.status)

			assert.Equal(t, tt.wantRes, res)
			if tt.wantErrMsg == "" {
				assert.Nil(t, err)
			} else {
				assert.Equal(t, tt.wantErrMsg, err.Error())
			}
		})
	}
}
