package usecase

import (
	"context"
	"go-api-example/internal/entity"
	"go-api-example/internal/model"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	List(ctx context.Context, req *model.SearchUserRequest) ([]entity.User, int, error)
	FindByID(ctx context.Context, id uint64) (*entity.User, error)
	FindByUsername(ctx context.Context, username string) (*entity.User, error)
	UpdateByID(ctx context.Context, req *model.UpdateUserRequest) error
	CountByUsername(ctx context.Context, username string) (int, error)
}
