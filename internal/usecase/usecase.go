package usecase

import (
	"context"
	"go-api-example/internal/model"
)

type AuthUsecase interface {
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
	Logout(ctx context.Context, req *model.LogoutRequest) error
	Refresh(ctx context.Context, req *model.RefreshRequest) (*model.RefreshResponse, error)
}

type UserUsecase interface {
	Create(ctx context.Context, req *model.CreateUserRequest) (*model.UserResponse, error)
	List(ctx context.Context, req *model.SearchUserRequest) ([]model.UserResponse, int, error)
	FindByID(ctx context.Context, req *model.GetUserRequest) (*model.UserResponse, error)
	UpdateByID(ctx context.Context, req *model.UpdateUserRequest) error
}
