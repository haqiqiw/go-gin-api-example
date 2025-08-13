package serializer

import (
	"go-api-example/internal/entity"
	"go-api-example/internal/model"
	"time"
)

func UserToResponse(u *entity.User) *model.UserResponse {
	return &model.UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
}

func ListUserToResponse(users []entity.User) []model.UserResponse {
	res := make([]model.UserResponse, len(users))

	for i, u := range users {
		res[i] = *UserToResponse(&u)
	}

	return res
}

func UserToEvent(u *entity.User) *model.UserEvent {
	return &model.UserEvent{
		ID:        u.ID,
		Username:  u.Username,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
	}
}
