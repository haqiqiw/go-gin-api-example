package model

import (
	"go-api-example/internal/entity"
)

type CreateTodoRequest struct {
	UserID      uint64  `json:"user_id"`
	Title       string  `json:"title" validate:"required"`
	Description *string `json:"description"`
}

type TodoResponse struct {
	ID          uint64 `json:"id"`
	UserID      uint64 `json:"user_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type SearchTodoRequest struct {
	UserID uint64             `json:"user_id"`
	Status *entity.TodoStatus `json:"status"`
	Limit  int                `json:"limit" validate:"min=1,max=20"`
	Offset int                `json:"offset" validate:"min=0"`
}

type GetTodoRequest struct {
	ID     uint64 `json:"id"`
	UserID uint64 `json:"user_id"`
}

type UpdateTodoRequest struct {
	ID          uint64            `json:"id"`
	UserID      uint64            `json:"user_id"`
	Title       string            `json:"title" validate:"required"`
	Description string            `json:"description"`
	Status      string            `json:"status" validate:"required"`
	IntStatus   entity.TodoStatus `json:"int_status"`
}
