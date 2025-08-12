package usecase

import (
	"context"
	"fmt"
	"go-api-example/internal/entity"
	"go-api-example/internal/model"
	"go-api-example/internal/model/serializer"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type todoUsecase struct {
	Log            *zap.Logger
	TodoRepository TodoRepository
}

func NewTodoUsecase(log *zap.Logger, todoRepository TodoRepository) TodoUsecase {
	return &todoUsecase{
		Log:            log,
		TodoRepository: todoRepository,
	}
}

func (c *todoUsecase) Create(ctx context.Context, req *model.CreateTodoRequest) (*model.TodoResponse, error) {
	todo := &entity.Todo{
		UserID:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
		Status:      entity.TodoStatusPending,
	}

	err := c.TodoRepository.Create(ctx, todo)
	if err != nil {
		return nil, fmt.Errorf("failed to create todo: %w", err)
	}

	return serializer.TodoToResponse(todo), nil
}

func (c *todoUsecase) List(ctx context.Context, req *model.SearchTodoRequest) ([]model.TodoResponse, int, error) {
	todos, total, err := c.TodoRepository.List(ctx, req)
	if err != nil {
		return []model.TodoResponse{}, 0, fmt.Errorf("failed to get todos: %w", err)
	}

	if len(todos) == 0 {
		return []model.TodoResponse{}, 0, nil
	}

	return serializer.ListTodoToResponse(todos), total, nil
}

func (c *todoUsecase) FindByID(ctx context.Context, req *model.GetTodoRequest) (*model.TodoResponse, error) {
	todo, err := c.TodoRepository.FindByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find todo by id: %w", err)
	}

	if req.UserID != todo.UserID {
		return nil, fiber.ErrForbidden
	}

	return serializer.TodoToResponse(todo), nil
}

func (c *todoUsecase) UpdateByID(ctx context.Context, req *model.UpdateTodoRequest) error {
	todo, err := c.TodoRepository.FindByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("failed to find todo by id: %w", err)
	}

	if req.UserID != todo.UserID {
		return fiber.ErrForbidden
	}

	err = c.TodoRepository.UpdateByID(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update todo by id: %w", err)
	}

	return nil
}
