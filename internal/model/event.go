package model

import "fmt"

type Event interface {
	GetID() string
}

type UserEvent struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (u *UserEvent) GetID() string {
	return fmt.Sprintf("%d-%s", u.ID, u.Username)
}
