package entity

import (
	"fmt"
	"time"
)

type TodoStatus int

const (
	TodoStatusPending TodoStatus = iota + 1
	TodoStatusInProgress
	TodoStatusCompleted
)

type Todo struct {
	ID          uint64     `db:"id"`
	UserID      uint64     `db:"user_id"`
	Title       string     `db:"title"`
	Description *string    `db:"description"`
	Status      TodoStatus `db:"status"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

func (t *Todo) GetDescription() string {
	if t != nil && t.Description != nil {
		return *t.Description
	}

	return ""
}

func (ts TodoStatus) String() string {
	switch ts {
	case TodoStatusPending:
		return "pending"
	case TodoStatusInProgress:
		return "in_progress"
	case TodoStatusCompleted:
		return "completed"
	default:
		return "unknown"
	}
}

func ParseTodoStatus(str string) (TodoStatus, error) {
	switch str {
	case "pending":
		return TodoStatusPending, nil
	case "in_progress":
		return TodoStatusInProgress, nil
	case "completed":
		return TodoStatusCompleted, nil
	default:
		return 0, fmt.Errorf("invalid status: %s", str)
	}
}
