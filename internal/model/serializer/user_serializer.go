package serializer

import (
	"go-api-example/internal/entity"
	"go-api-example/internal/model"
)

func UserToResponse(u *entity.User) *model.UserResponse {
	return &model.UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func ListUserToResponse(users []entity.User) []model.UserResponse {
	res := make([]model.UserResponse, len(users))

	for i, u := range users {
		res[i] = model.UserResponse{
			ID:        u.ID,
			Username:  u.Username,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		}
	}

	return res
}
