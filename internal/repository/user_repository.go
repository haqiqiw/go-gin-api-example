package repository

import (
	"context"
	"database/sql"
	"errors"
	"go-api-example/internal/entity"
	"go-api-example/internal/model"
	"strings"
	"time"
)

type UserRepository struct {
	DB *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	now := time.Now()
	query := `INSERT INTO users (username, password, created_at, updated_at) VALUES (?, ?, ?, ?)`

	res, err := r.DB.ExecContext(ctx, query, user.Username, user.Password, now, now)
	if err != nil {
		return err
	}

	id, _ := res.LastInsertId()
	user.ID = uint64(id)

	return nil
}

func (r *UserRepository) List(ctx context.Context, req *model.SearchUserRequest) ([]entity.User, int, error) {
	var conditions []string
	var args []any

	if req.ID != nil {
		conditions = append(conditions, "id = ?")
		args = append(args, *req.ID)
	}
	if req.Username != nil {
		conditions = append(conditions, "username = ?")
		args = append(args, *req.Username)
	}

	var countSb strings.Builder
	countSb.WriteString("SELECT COUNT(id) FROM users")

	if len(conditions) > 0 {
		countSb.WriteString(" WHERE ")
		countSb.WriteString(strings.Join(conditions, " AND "))
	}

	var total int
	if err := r.DB.QueryRowContext(ctx, countSb.String(), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	var sb strings.Builder
	sb.WriteString(`SELECT id, username, password, created_at, updated_at FROM users`)

	if len(conditions) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(conditions, " AND "))
	}

	sb.WriteString(" ORDER BY id ASC LIMIT ? OFFSET ?")
	args = append(args, req.Limit, req.Offset)

	rows, err := r.DB.QueryContext(ctx, sb.String(), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []entity.User
	for rows.Next() {
		var u entity.User
		err := rows.Scan(&u.ID, &u.Username, &u.Password, &u.CreatedAt, &u.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		users = append(users, u)
	}

	return users, total, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id uint64) (*entity.User, error) {
	query := `SELECT id, username, password, created_at, updated_at FROM users WHERE id = ? LIMIT 1`

	var u entity.User
	err := r.DB.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Username, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &u, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*entity.User, error) {
	query := `SELECT id, username, password, created_at, updated_at FROM users WHERE username = ? LIMIT 1`

	var u entity.User
	err := r.DB.QueryRowContext(ctx, query, username).Scan(&u.ID, &u.Username, &u.Password, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &u, nil
}

func (r *UserRepository) UpdateByID(ctx context.Context, req *model.UpdateUserRequest) error {
	now := time.Now()
	query := `UPDATE users SET password = ?, updated_at = ? WHERE id = ?`

	_, err := r.DB.ExecContext(ctx, query, req.NewPassword, now, req.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepository) CountByUsername(ctx context.Context, username string) (int, error) {
	query := `SELECT COUNT(id) FROM users WHERE username = ?`

	var count int
	err := r.DB.QueryRowContext(ctx, query, username).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
