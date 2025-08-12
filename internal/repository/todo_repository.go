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

type TodoRepository struct {
	DB *sql.DB
}

func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{
		DB: db,
	}
}

func (r *TodoRepository) Create(ctx context.Context, todo *entity.Todo) error {
	now := time.Now()
	query := `INSERT INTO todos (user_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`

	res, err := r.DB.ExecContext(ctx, query, todo.UserID, todo.Title, todo.Description, todo.Status, now, now)
	if err != nil {
		return err
	}

	id, _ := res.LastInsertId()
	todo.ID = uint64(id)
	todo.CreatedAt = now
	todo.UpdatedAt = now

	return nil
}

func (r *TodoRepository) List(ctx context.Context, req *model.SearchTodoRequest) ([]entity.Todo, int, error) {
	conditions := []string{"user_id = ?"}
	args := []any{req.UserID}

	if req.Status != nil {
		conditions = append(conditions, "status = ?")
		args = append(args, *req.Status)
	}

	var countSb strings.Builder
	countSb.WriteString("SELECT COUNT(id) FROM todos")

	if len(conditions) > 0 {
		countSb.WriteString(" WHERE ")
		countSb.WriteString(strings.Join(conditions, " AND "))
	}

	var total int
	if err := r.DB.QueryRowContext(ctx, countSb.String(), args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	var sb strings.Builder
	sb.WriteString(`SELECT id, user_id, title, description, status, created_at, updated_at FROM todos`)

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

	var todos []entity.Todo
	for rows.Next() {
		var t entity.Todo
		err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt)
		if err != nil {
			return nil, 0, err
		}
		todos = append(todos, t)
	}

	return todos, total, nil
}

func (r *TodoRepository) FindByID(ctx context.Context, id uint64) (*entity.Todo, error) {
	query := `SELECT id, user_id, title, description, status, created_at, updated_at FROM todos WHERE id = ? LIMIT 1`

	var t entity.Todo
	err := r.DB.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.UserID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return &t, nil
}

func (r *TodoRepository) UpdateByID(ctx context.Context, req *model.UpdateTodoRequest) error {
	now := time.Now()
	query := `UPDATE todos SET title = ?, description = ?, status = ?, updated_at = ? WHERE id = ?`

	_, err := r.DB.ExecContext(ctx, query, req.Title, req.Description, req.IntStatus, now, req.ID)
	if err != nil {
		return err
	}

	return nil
}
