package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"go-api-example/internal/entity"
	"go-api-example/internal/model"
	"go-api-example/internal/repository"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
)

type TodoRepositorySuite struct {
	suite.Suite
	db   *sql.DB
	mock sqlmock.Sqlmock
	repo *repository.TodoRepository
	ctx  context.Context
	now  time.Time
}

func (s *TodoRepositorySuite) SetupTest() {
	db, mock, _ := sqlmock.New()
	s.db = db
	s.mock = mock
	s.repo = repository.NewTodoRepository(s.db)
	s.ctx = context.Background()
	s.now = time.Now()
}

func (s *TodoRepositorySuite) TearDownTest() {
	s.db.Close()
}

func (s *TodoRepositorySuite) TestTodoRepository_Create() {
	description := "dummy description"

	tests := []struct {
		name     string
		mockFunc func(sqlmock.Sqlmock)
		param    *entity.Todo
		wantErr  error
	}{
		{
			name: "success",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectExec(regexp.QuoteMeta(
					`INSERT INTO todos (user_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
				)).
					WithArgs(1, "dummy title", "dummy description", 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			param: &entity.Todo{
				UserID:      1,
				Title:       "dummy title",
				Description: &description,
				Status:      entity.TodoStatusPending,
			},
			wantErr: nil,
		},
		{
			name: "unexpected error",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectExec(regexp.QuoteMeta(
					`INSERT INTO todos (user_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
				)).
					WithArgs(1, "dummy title", "dummy description", 1, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("something error"))
			},
			param: &entity.Todo{
				UserID:      1,
				Title:       "dummy title",
				Description: &description,
				Status:      entity.TodoStatusPending,
			},
			wantErr: errors.New("something error"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.mockFunc(s.mock)

			err := s.repo.Create(s.ctx, tt.param)
			s.Equal(tt.wantErr, err)
		})
	}
}

func (s *TodoRepositorySuite) TestTodoRepository_List() {
	description := "dummy description"
	status := entity.TodoStatusCompleted

	tests := []struct {
		name      string
		mockFunc  func(sqlmock.Sqlmock)
		param     *model.SearchTodoRequest
		wantTodos []entity.Todo
		wantTotal int
		wantErr   error
	}{
		{
			name: "success with default param",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM todos WHERE user_id = ?`,
				)).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

				rows := sqlmock.NewRows([]string{"id", "user_id", "title", "description", "status", "created_at", "updated_at"}).
					AddRow(1, 1, "dummy title 1", description, 1, s.now, s.now).
					AddRow(2, 1, "dummy title 2", description, 2, s.now, s.now)
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, user_id, title, description, status, created_at, updated_at FROM todos
					WHERE user_id = ? ORDER BY id ASC LIMIT ? OFFSET ?`,
				)).
					WithArgs(1, 10, 0).
					WillReturnRows(rows)
			},
			param: &model.SearchTodoRequest{
				UserID: 1,
				Limit:  10,
				Offset: 0,
			},
			wantTodos: []entity.Todo{
				{
					ID:          1,
					UserID:      1,
					Title:       "dummy title 1",
					Description: &description,
					Status:      entity.TodoStatusPending,
					CreatedAt:   s.now,
					UpdatedAt:   s.now,
				},
				{
					ID:          2,
					UserID:      1,
					Title:       "dummy title 2",
					Description: &description,
					Status:      entity.TodoStatusInProgress,
					CreatedAt:   s.now,
					UpdatedAt:   s.now,
				},
			},
			wantTotal: 2,
			wantErr:   nil,
		},
		{
			name: "success with status param",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM todos WHERE user_id = ? AND status = ?`,
				)).
					WithArgs(1, 3).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

				rows := sqlmock.NewRows([]string{"id", "user_id", "title", "description", "status", "created_at", "updated_at"}).
					AddRow(1, 1, "dummy title 1", description, 3, s.now, s.now).
					AddRow(2, 1, "dummy title 2", description, 3, s.now, s.now)
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, user_id, title, description, status, created_at, updated_at FROM todos
					WHERE user_id = ? AND status = ? ORDER BY id ASC LIMIT ? OFFSET ?`,
				)).
					WithArgs(1, 3, 10, 0).
					WillReturnRows(rows)
			},
			param: &model.SearchTodoRequest{
				UserID: 1,
				Status: &status,
				Limit:  10,
				Offset: 0,
			},
			wantTodos: []entity.Todo{
				{
					ID:          1,
					UserID:      1,
					Title:       "dummy title 1",
					Description: &description,
					Status:      entity.TodoStatusCompleted,
					CreatedAt:   s.now,
					UpdatedAt:   s.now,
				},
				{
					ID:          2,
					UserID:      1,
					Title:       "dummy title 2",
					Description: &description,
					Status:      entity.TodoStatusCompleted,
					CreatedAt:   s.now,
					UpdatedAt:   s.now,
				},
			},
			wantTotal: 2,
			wantErr:   nil,
		},
		{
			name: "not found",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM todos WHERE user_id = ?`,
				)).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				rows := sqlmock.NewRows([]string{"id", "user_id", "title", "description", "status", "created_at", "updated_at"})
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, user_id, title, description, status, created_at, updated_at FROM todos
					WHERE user_id = ? ORDER BY id ASC LIMIT ? OFFSET ?`,
				)).
					WithArgs(1, 10, 0).
					WillReturnRows(rows)
			},
			param: &model.SearchTodoRequest{
				UserID: 1,
				Limit:  10,
				Offset: 0,
			},
			wantTodos: nil,
			wantTotal: 0,
			wantErr:   nil,
		},
		{
			name: "unexpected error when count rows",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM todos WHERE user_id = ?`,
				)).
					WithArgs(1).
					WillReturnError(errors.New("something error"))
			},
			param: &model.SearchTodoRequest{
				UserID: 1,
				Limit:  10,
				Offset: 0,
			},
			wantTodos: nil,
			wantTotal: 0,
			wantErr:   errors.New("something error"),
		},
		{
			name: "unexpected error when select rows",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM todos WHERE user_id = ?`,
				)).
					WithArgs(1).
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, user_id, title, description, status, created_at, updated_at FROM todos
					WHERE user_id = ? ORDER BY id ASC LIMIT ? OFFSET ?`,
				)).
					WithArgs(1, 10, 0).
					WillReturnError(errors.New("something error"))
			},
			param: &model.SearchTodoRequest{
				UserID: 1,
				Limit:  10,
				Offset: 0,
			},
			wantTodos: nil,
			wantTotal: 0,
			wantErr:   errors.New("something error"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.mockFunc(s.mock)

			res, total, err := s.repo.List(s.ctx, tt.param)
			s.Equal(tt.wantTodos, res)
			s.Equal(tt.wantTotal, total)
			s.Equal(tt.wantErr, err)
		})
	}
}

func (s *TodoRepositorySuite) TestTodoRepository_FindByID() {
	description := "dummy description"

	tests := []struct {
		name     string
		mockFunc func(sqlmock.Sqlmock)
		paramID  uint64
		wantTodo *entity.Todo
		wantErr  error
	}{
		{
			name: "success",
			mockFunc: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "user_id", "title", "description", "status", "created_at", "updated_at"}).
					AddRow(1, 1, "dummy title", description, 1, s.now, s.now)
				m.ExpectQuery(`SELECT id, user_id, title, description, status, created_at, updated_at FROM todos WHERE id = \? LIMIT 1`).
					WithArgs(1).
					WillReturnRows(rows)
			},
			paramID: 1,
			wantTodo: &entity.Todo{
				ID:          1,
				UserID:      1,
				Title:       "dummy title",
				Description: &description,
				Status:      entity.TodoStatusPending,
				CreatedAt:   s.now,
				UpdatedAt:   s.now,
			},
			wantErr: nil,
		},
		{
			name: "not found",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id, user_id, title, description, status, created_at, updated_at FROM todos WHERE id = \? LIMIT 1`).
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			paramID:  1,
			wantTodo: nil,
			wantErr:  nil,
		},
		{
			name: "unexpected error",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(`SELECT id, user_id, title, description, status, created_at, updated_at FROM todos WHERE id = \? LIMIT 1`).
					WithArgs(1).
					WillReturnError(errors.New("something error"))
			},
			paramID:  1,
			wantTodo: nil,
			wantErr:  errors.New("something error"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.mockFunc(s.mock)

			res, err := s.repo.FindByID(s.ctx, tt.paramID)
			s.Equal(tt.wantTodo, res)
			s.Equal(tt.wantErr, err)
		})
	}
}

func (s *TodoRepositorySuite) TestTodoRepository_UpdateByID() {
	tests := []struct {
		name     string
		mockFunc func(sqlmock.Sqlmock)
		param    *model.UpdateTodoRequest
		wantErr  error
	}{
		{
			name: "success",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE todos SET title = \?, description = \?, status = \?, updated_at = \? WHERE id = \?`).
					WithArgs("new title", "new description", 2, sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			param: &model.UpdateTodoRequest{
				ID:          1,
				UserID:      1,
				Title:       "new title",
				Description: "new description",
				Status:      "in_progress",
				IntStatus:   entity.TodoStatusInProgress,
			},
			wantErr: nil,
		},
		{
			name: "unexpected error",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectExec(`UPDATE todos SET title = \?, description = \?, status = \?, updated_at = \? WHERE id = \?`).
					WithArgs("new title", "new description", 2, sqlmock.AnyArg(), 1).
					WillReturnError(errors.New("something error"))
			},
			param: &model.UpdateTodoRequest{
				ID:          1,
				UserID:      1,
				Title:       "new title",
				Description: "new description",
				Status:      "in_progress",
				IntStatus:   entity.TodoStatusInProgress,
			},
			wantErr: errors.New("something error"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.mockFunc(s.mock)

			err := s.repo.UpdateByID(s.ctx, tt.param)
			s.Equal(tt.wantErr, err)
		})
	}
}

func TestTodoRepositorySuite(t *testing.T) {
	suite.Run(t, new(TodoRepositorySuite))
}
