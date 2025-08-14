package repository_test

import (
	"context"
	"database/sql"
	"errors"
	"go-api-example/internal/db"
	"go-api-example/internal/entity"
	"go-api-example/internal/model"
	"go-api-example/internal/repository"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
)

type UserRepositorySuite struct {
	suite.Suite
	db   *sql.DB
	mock sqlmock.Sqlmock
	exec db.Executor
	repo *repository.UserRepository
	ctx  context.Context
	now  time.Time
}

func (s *UserRepositorySuite) SetupTest() {
	db, mock, _ := sqlmock.New()
	s.db = db
	s.mock = mock
	s.exec = db
	s.repo = repository.NewUserRepository(s.db)
	s.ctx = context.Background()
	s.now = time.Now()
}

func (s *UserRepositorySuite) TearDownTest() {
	s.db.Close()
}

func (s *UserRepositorySuite) TestUserRepository_Create() {
	tests := []struct {
		name     string
		mockFunc func(sqlmock.Sqlmock)
		param    *entity.User
		wantErr  error
	}{
		{
			name: "success",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectExec(regexp.QuoteMeta(
					`INSERT INTO users (username, password, created_at, updated_at) VALUES (?, ?, ?, ?)`,
				)).
					WithArgs("johndoe", "password", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(1, 0))
			},
			param: &entity.User{
				Username: "johndoe",
				Password: "password",
			},
			wantErr: nil,
		},
		{
			name: "unexpected error",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectExec(regexp.QuoteMeta(
					`INSERT INTO users (username, password, created_at, updated_at) VALUES (?, ?, ?, ?)`,
				)).
					WithArgs("johndoe", "password", sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("something error"))
			},
			param: &entity.User{
				Username: "johndoe",
				Password: "password",
			},
			wantErr: errors.New("something error"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.mockFunc(s.mock)

			err := s.repo.Create(s.ctx, s.exec, tt.param)
			s.Equal(tt.wantErr, err)
		})
	}
}

func (s *UserRepositorySuite) TestUserRepository_List() {
	id := uint64(1)
	username := "johndoe"

	tests := []struct {
		name      string
		mockFunc  func(sqlmock.Sqlmock)
		param     *model.SearchUserRequest
		wantUsers []entity.User
		wantTotal int
		wantErr   error
	}{
		{
			name: "success with default param",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM users`,
				)).WithoutArgs().WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

				rows := sqlmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
					AddRow(1, "johndoe", "password", s.now, s.now).
					AddRow(2, "chyntia", "password", s.now, s.now)
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, username, password, created_at, updated_at FROM users
					ORDER BY id ASC LIMIT ? OFFSET ?`,
				)).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			param: &model.SearchUserRequest{
				Limit:  10,
				Offset: 0,
			},
			wantUsers: []entity.User{
				{
					ID:        1,
					Username:  "johndoe",
					Password:  "password",
					CreatedAt: s.now,
					UpdatedAt: s.now,
				},
				{
					ID:        2,
					Username:  "chyntia",
					Password:  "password",
					CreatedAt: s.now,
					UpdatedAt: s.now,
				},
			},
			wantTotal: 2,
			wantErr:   nil,
		},
		{
			name: "success with custom param",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM users WHERE id = ? AND username = ?`,
				)).
					WithArgs(1, "johndoe").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				rows := sqlmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
					AddRow(1, "johndoe", "password", s.now, s.now)
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, username, password, created_at, updated_at FROM users
					 WHERE id = ? AND username = ? ORDER BY id ASC LIMIT ? OFFSET ?`,
				)).
					WithArgs(1, "johndoe", 10, 0).
					WillReturnRows(rows)
			},
			param: &model.SearchUserRequest{
				ID:       &id,
				Username: &username,
				Limit:    10,
				Offset:   0,
			},
			wantUsers: []entity.User{
				{
					ID:        1,
					Username:  "johndoe",
					Password:  "password",
					CreatedAt: s.now,
					UpdatedAt: s.now,
				},
			},
			wantTotal: 1,
			wantErr:   nil,
		},
		{
			name: "not found",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM users`,
				)).WithoutArgs().WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

				rows := sqlmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"})
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, username, password, created_at, updated_at FROM users
					ORDER BY id ASC LIMIT ? OFFSET ?`,
				)).
					WithArgs(10, 0).
					WillReturnRows(rows)
			},
			param: &model.SearchUserRequest{
				Limit:  10,
				Offset: 0,
			},
			wantUsers: nil,
			wantTotal: 0,
			wantErr:   nil,
		},
		{
			name: "unexpected error when count rows",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM users`,
				)).WithoutArgs().WillReturnError(errors.New("something error"))
			},
			param: &model.SearchUserRequest{
				Limit:  10,
				Offset: 0,
			},
			wantUsers: nil,
			wantTotal: 0,
			wantErr:   errors.New("something error"),
		},
		{
			name: "unexpected error when select rows",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM users`,
				)).WithoutArgs().WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, username, password, created_at, updated_at FROM users
					ORDER BY id ASC LIMIT ? OFFSET ?`,
				)).
					WithArgs(10, 0).
					WillReturnError(errors.New("something error"))
			},
			param: &model.SearchUserRequest{
				Limit:  10,
				Offset: 0,
			},
			wantUsers: nil,
			wantTotal: 0,
			wantErr:   errors.New("something error"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.mockFunc(s.mock)

			res, total, err := s.repo.List(s.ctx, tt.param)
			s.Equal(tt.wantUsers, res)
			s.Equal(tt.wantTotal, total)
			s.Equal(tt.wantErr, err)
		})
	}
}

func (s *UserRepositorySuite) TestUserRepository_FindByID() {
	tests := []struct {
		name     string
		mockFunc func(sqlmock.Sqlmock)
		paramID  uint64
		wantUser *entity.User
		wantErr  error
	}{
		{
			name: "success",
			mockFunc: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
					AddRow(1, "johndoe", "password", s.now, s.now)
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, username, password, created_at, updated_at FROM users WHERE id = ? LIMIT 1`,
				)).
					WithArgs(1).
					WillReturnRows(rows)
			},
			paramID: 1,
			wantUser: &entity.User{
				ID:        1,
				Username:  "johndoe",
				Password:  "password",
				CreatedAt: s.now,
				UpdatedAt: s.now,
			},
			wantErr: nil,
		},
		{
			name: "not found",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, username, password, created_at, updated_at FROM users WHERE id = ? LIMIT 1`,
				)).
					WithArgs(1).
					WillReturnError(sql.ErrNoRows)
			},
			paramID:  1,
			wantUser: nil,
			wantErr:  nil,
		},
		{
			name: "unexpected error",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, username, password, created_at, updated_at FROM users WHERE id = ? LIMIT 1`,
				)).
					WithArgs(1).
					WillReturnError(errors.New("something error"))
			},
			paramID:  1,
			wantUser: nil,
			wantErr:  errors.New("something error"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.mockFunc(s.mock)

			res, err := s.repo.FindByID(s.ctx, tt.paramID)
			s.Equal(tt.wantUser, res)
			s.Equal(tt.wantErr, err)
		})
	}
}

func (s *UserRepositorySuite) TestUserRepository_FindByUsername() {
	tests := []struct {
		name          string
		mockFunc      func(sqlmock.Sqlmock)
		paramUsername string
		wantUser      *entity.User
		wantErr       error
	}{
		{
			name: "success",
			mockFunc: func(m sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "password", "created_at", "updated_at"}).
					AddRow(1, "johndoe", "password", s.now, s.now)
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, username, password, created_at, updated_at FROM users WHERE username = ? LIMIT 1`,
				)).
					WithArgs("johndoe").
					WillReturnRows(rows)
			},
			paramUsername: "johndoe",
			wantUser: &entity.User{
				ID:        1,
				Username:  "johndoe",
				Password:  "password",
				CreatedAt: s.now,
				UpdatedAt: s.now,
			},
			wantErr: nil,
		},
		{
			name: "not found",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, username, password, created_at, updated_at FROM users WHERE username = ? LIMIT 1`,
				)).
					WithArgs("johndoe").
					WillReturnError(sql.ErrNoRows)
			},
			paramUsername: "johndoe",
			wantUser:      nil,
			wantErr:       nil,
		},
		{
			name: "unexpected error",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT id, username, password, created_at, updated_at FROM users WHERE username = ? LIMIT 1`,
				)).
					WithArgs("johndoe").
					WillReturnError(errors.New("something error"))
			},
			paramUsername: "johndoe",
			wantUser:      nil,
			wantErr:       errors.New("something error"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.mockFunc(s.mock)

			res, err := s.repo.FindByUsername(s.ctx, tt.paramUsername)
			s.Equal(tt.wantUser, res)
			s.Equal(tt.wantErr, err)
		})
	}
}

func (s *UserRepositorySuite) TestUserRepository_UpdateByID() {
	tests := []struct {
		name     string
		mockFunc func(sqlmock.Sqlmock)
		param    *model.UpdateUserRequest
		wantErr  error
	}{
		{
			name: "success",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectExec(regexp.QuoteMeta(
					`UPDATE users SET password = ?, updated_at = ? WHERE id = ?`,
				)).
					WithArgs("newpassword", sqlmock.AnyArg(), 1).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
			param: &model.UpdateUserRequest{
				ID:          1,
				OldPassword: "oldpassword",
				NewPassword: "newpassword",
			},
			wantErr: nil,
		},
		{
			name: "unexpected error",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectExec(regexp.QuoteMeta(
					`UPDATE users SET password = ?, updated_at = ? WHERE id = ?`,
				)).
					WithArgs("newpassword", sqlmock.AnyArg(), 1).
					WillReturnError(errors.New("something error"))
			},
			param: &model.UpdateUserRequest{
				ID:          1,
				OldPassword: "oldpassword",
				NewPassword: "newpassword",
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

func (s *UserRepositorySuite) TestUserRepository_CountByUsername() {
	tests := []struct {
		name          string
		mockFunc      func(sqlmock.Sqlmock)
		paramUsername string
		wantTotal     int
		wantErr       error
	}{
		{
			name: "success",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM users WHERE username = ?`,
				)).
					WithArgs("johndoe").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
			},
			paramUsername: "johndoe",
			wantTotal:     1,
			wantErr:       nil,
		},
		{
			name: "not found",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM users WHERE username = ?`,
				)).
					WithArgs("johndoe").
					WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
			},
			paramUsername: "johndoe",
			wantTotal:     0,
			wantErr:       nil,
		},
		{
			name: "unexpected error",
			mockFunc: func(m sqlmock.Sqlmock) {
				m.ExpectQuery(regexp.QuoteMeta(
					`SELECT COUNT(id) FROM users WHERE username = ?`,
				)).
					WithArgs("johndoe").
					WillReturnError(errors.New("something error"))
			},
			paramUsername: "johndoe",
			wantTotal:     0,
			wantErr:       errors.New("something error"),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.mockFunc(s.mock)

			res, err := s.repo.CountByUsername(s.ctx, tt.paramUsername)
			s.Equal(tt.wantTotal, res)
			s.Equal(tt.wantErr, err)
		})
	}
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(t, new(UserRepositorySuite))
}
