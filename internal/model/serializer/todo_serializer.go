package serializer

import (
	"go-api-example/internal/entity"
	"go-api-example/internal/model"
	"time"
)

func TodoToResponse(t *entity.Todo) *model.TodoResponse {
	return &model.TodoResponse{
		ID:          t.ID,
		UserID:      t.UserID,
		Title:       t.Title,
		Description: t.GetDescription(),
		Status:      t.Status.String(),
		CreatedAt:   t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   t.UpdatedAt.Format(time.RFC3339),
	}
}

func ListTodoToResponse(todos []entity.Todo) []model.TodoResponse {
	res := make([]model.TodoResponse, len(todos))

	for i, t := range todos {
		res[i] = *TodoToResponse(&t)
	}

	return res
}
