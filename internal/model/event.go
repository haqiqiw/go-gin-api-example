package model

import "fmt"

type Event interface {
	GetId() string
}

type UserEvent struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (u *UserEvent) GetId() string {
	return fmt.Sprintf("%d-%s", u.ID, u.Username)
}
