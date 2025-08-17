package usecase

import (
	"context"
	"go-api-example/internal/db"
	"go-api-example/internal/entity"
	"go-api-example/internal/model"
)

//go:generate mockery --name=UserRepository --structname UserRepository --outpkg=mocks --output=./../mocks
type UserRepository interface {
	Create(ctx context.Context, exec db.Executor, user *entity.User) error
	List(ctx context.Context, req *model.SearchUserRequest) ([]entity.User, int, error)
	FindByID(ctx context.Context, id uint64) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	UpdateByID(ctx context.Context, req *model.UpdateUserRequest) error
	CountByUsername(ctx context.Context, username string) (int, error)
}

//go:generate mockery --name=TodoRepository --structname TodoRepository --outpkg=mocks --output=./../mocks
type TodoRepository interface {
	Create(ctx context.Context, user *entity.Todo) error
	List(ctx context.Context, req *model.SearchTodoRequest) ([]entity.Todo, int, error)
	FindByID(ctx context.Context, id uint64) (*entity.Todo, error)
	UpdateByID(ctx context.Context, req *model.UpdateTodoRequest) error
}
