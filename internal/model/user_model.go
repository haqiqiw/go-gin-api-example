package model

type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=4,max=64"`
	Password string `json:"password" validate:"required,min=4,max=64"`
}

type SearchUserRequest struct {
	ID       *uint64 `json:"id"`
	Username *string `json:"username"`
	Limit    int     `json:"limit" validate:"min=1,max=20"`
	Offset   int     `json:"offset" validate:"min=0"`
}

type GetUserRequest struct {
	ID uint64 `json:"id"`
}

type UpdateUserRequest struct {
	ID          uint64 `json:"id"`
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

type UserResponse struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
