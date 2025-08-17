package usecase

import (
	"context"
	"go-api-example/internal/model"
)

//go:generate mockery --name=AuthUsecase --structname AuthUsecase --outpkg=mocks --output=./../mocks
type AuthUsecase interface {
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
	Logout(ctx context.Context, req *model.LogoutRequest) error
	Refresh(ctx context.Context, req *model.RefreshRequest) (*model.RefreshResponse, error)
}

//go:generate mockery --name=UserUsecase --structname UserUsecase --outpkg=mocks --output=./../mocks
type UserUsecase interface {
	Create(ctx context.Context, req *model.CreateUserRequest) (*model.UserResponse, error)
	List(ctx context.Context, req *model.SearchUserRequest) ([]model.UserResponse, int, error)
	FindByID(ctx context.Context, req *model.GetUserRequest) (*model.UserResponse, error)
	UpdateByID(ctx context.Context, req *model.UpdateUserRequest) error
}

//go:generate mockery --name=TodoUsecase --structname TodoUsecase --outpkg=mocks --output=./../mocks
type TodoUsecase interface {
	Create(ctx context.Context, req *model.CreateTodoRequest) (*model.TodoResponse, error)
	List(ctx context.Context, req *model.SearchTodoRequest) ([]model.TodoResponse, int, error)
	FindByID(ctx context.Context, req *model.GetTodoRequest) (*model.TodoResponse, error)
	UpdateByID(ctx context.Context, req *model.UpdateTodoRequest) error
}
