package entity

import "time"

type User struct {
	ID        uint64    `db:"id"`
	Username  string    `db:"username"`
	Password  string    `db:"password"`
	CreatedAt time.Time `db:"id"`
	UpdatedAt time.Time `db:"id"`
}
