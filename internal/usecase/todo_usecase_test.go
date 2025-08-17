package usecase_test

import (
	"context"
	"errors"
	"go-api-example/internal/entity"
	"go-api-example/internal/mocks"
	"go-api-example/internal/model"
	"go-api-example/internal/usecase"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type TodoUsecaseSuite struct {
	suite.Suite
	log *zap.Logger
	ctx context.Context
}

func (s *TodoUsecaseSuite) SetupTest() {
	s.log, _ = zap.NewDevelopment()
	s.ctx = context.Background()
}

func (s *TodoUsecaseSuite) TestTodoRepository_Create() {
	description := "description"
	now := time.Now()

	tests := []struct {
		name       string
		request    *model.CreateTodoRequest
		mockFunc   func(r *mocks.TodoRepository)
		wantTodo   *model.TodoResponse
		wantErrMsg string
	}{
		{
			name: "error on create",
			request: &model.CreateTodoRequest{
				UserID:      1,
				Title:       "title",
				Description: &description,
			},
			mockFunc: func(r *mocks.TodoRepository) {
				r.On("Create", mock.Anything, mock.Anything).
					Return(errors.New("something error"))
			},
			wantTodo:   nil,
			wantErrMsg: "failed to create todo: something error",
		},
		{
			name: "success",
			request: &model.CreateTodoRequest{
				UserID:      1,
				Title:       "title",
				Description: &description,
			},
			mockFunc: func(r *mocks.TodoRepository) {
				r.On("Create", mock.Anything, mock.Anything).Return(nil).
					Run(func(args mock.Arguments) {
						t := args.Get(1).(*entity.Todo)
						t.ID = 1
						t.CreatedAt = now
						t.UpdatedAt = now
					})
			},
			wantTodo: &model.TodoResponse{
				ID:          1,
				UserID:      1,
				Title:       "title",
				Description: description,
				Status:      entity.TodoStatusPending.String(),
				CreatedAt:   now.Format(time.RFC3339),
				UpdatedAt:   now.Format(time.RFC3339),
			},
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			todoRepository := mocks.NewTodoRepository(s.T())
			usecase := usecase.NewTodoUsecase(s.log, todoRepository)
			tt.mockFunc(todoRepository)

			res, err := usecase.Create(s.ctx, tt.request)

			if tt.wantErrMsg != "" {
				s.Nil(res)
				s.Equal(tt.wantErrMsg, err.Error())
			} else {
				s.Equal(*tt.wantTodo, *res)
				s.Nil(err)
			}
		})
	}
}

func (s *TodoUsecaseSuite) TestTodoRepository_List() {
	description := "description"
	now := time.Now()

	tests := []struct {
		name       string
		request    *model.SearchTodoRequest
		mockFunc   func(r *mocks.TodoRepository)
		wantTodos  []model.TodoResponse
		wantTotal  int
		wantErrMsg string
	}{
		{
			name: "error on find",
			request: &model.SearchTodoRequest{
				Limit:  10,
				Offset: 0,
			},
			mockFunc: func(r *mocks.TodoRepository) {
				r.On("List", mock.Anything, mock.Anything).
					Return(nil, 0, errors.New("something error"))
			},
			wantTodos:  []model.TodoResponse{},
			wantTotal:  0,
			wantErrMsg: "failed to get todos: something error",
		},
		{
			name: "success",
			request: &model.SearchTodoRequest{
				Limit:  10,
				Offset: 0,
			},
			mockFunc: func(r *mocks.TodoRepository) {
				r.On("List", mock.Anything, mock.Anything).Return([]entity.Todo{
					{
						ID:          1,
						UserID:      1,
						Title:       "title",
						Description: &description,
						Status:      entity.TodoStatusPending,
						CreatedAt:   now,
						UpdatedAt:   now,
					},
				}, 1, nil)
			},
			wantTodos: []model.TodoResponse{
				{
					ID:          1,
					UserID:      1,
					Title:       "title",
					Description: description,
					Status:      entity.TodoStatusPending.String(),
					CreatedAt:   now.Format(time.RFC3339),
					UpdatedAt:   now.Format(time.RFC3339),
				},
			},
			wantTotal:  1,
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			todoRepository := mocks.NewTodoRepository(s.T())
			usecase := usecase.NewTodoUsecase(s.log, todoRepository)
			tt.mockFunc(todoRepository)

			res, total, err := usecase.List(s.ctx, tt.request)

			if tt.wantErrMsg != "" {
				s.Empty(res)
				s.Equal(tt.wantErrMsg, err.Error())
			} else {
				s.Len(res, 1)
				s.Equal(tt.wantTodos[0], res[0])
				s.Nil(err)
			}
			s.Equal(tt.wantTotal, total)
		})
	}
}

func (s *TodoUsecaseSuite) TestTodoRepository_FindByID() {
	description := "description"
	now := time.Now()

	tests := []struct {
		name       string
		request    *model.GetTodoRequest
		mockFunc   func(r *mocks.TodoRepository)
		wantTodo   *model.TodoResponse
		wantErrMsg string
	}{
		{
			name: "error on find",
			request: &model.GetTodoRequest{
				ID:     1,
				UserID: 1,
			},
			mockFunc: func(r *mocks.TodoRepository) {
				r.On("FindByID", mock.Anything, uint64(1)).
					Return(nil, errors.New("something error"))
			},
			wantTodo:   nil,
			wantErrMsg: "failed to find todo by id: something error",
		},
		{
			name: "error forbidden",
			request: &model.GetTodoRequest{
				ID:     1,
				UserID: 1,
			},
			mockFunc: func(r *mocks.TodoRepository) {
				r.On("FindByID", mock.Anything, uint64(1)).Return(&entity.Todo{
					ID:          1,
					UserID:      2,
					Title:       "title",
					Description: &description,
					Status:      entity.TodoStatusPending,
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil)
			},
			wantTodo:   nil,
			wantErrMsg: "Forbidden",
		},
		{
			name: "success",
			request: &model.GetTodoRequest{
				ID:     1,
				UserID: 1,
			},
			mockFunc: func(r *mocks.TodoRepository) {
				r.On("FindByID", mock.Anything, uint64(1)).Return(&entity.Todo{
					ID:          1,
					UserID:      1,
					Title:       "title",
					Description: &description,
					Status:      entity.TodoStatusPending,
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil)
			},
			wantTodo: &model.TodoResponse{
				ID:          1,
				UserID:      1,
				Title:       "title",
				Description: description,
				Status:      entity.TodoStatusPending.String(),
				CreatedAt:   now.Format(time.RFC3339),
				UpdatedAt:   now.Format(time.RFC3339),
			},
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			todoRepository := mocks.NewTodoRepository(s.T())
			usecase := usecase.NewTodoUsecase(s.log, todoRepository)
			tt.mockFunc(todoRepository)

			res, err := usecase.FindByID(s.ctx, tt.request)

			if tt.wantErrMsg != "" {
				s.Nil(res)
				s.Equal(tt.wantErrMsg, err.Error())
			} else {
				s.Equal(*tt.wantTodo, *res)
				s.Nil(err)
			}
		})
	}
}

func (s *TodoUsecaseSuite) TestTodoRepository_UpdateByID() {
	description := "description"
	now := time.Now()

	tests := []struct {
		name       string
		request    *model.UpdateTodoRequest
		mockFunc   func(r *mocks.TodoRepository)
		wantErrMsg string
	}{
		{
			name: "error on find",
			request: &model.UpdateTodoRequest{
				ID:          1,
				UserID:      1,
				Title:       "new title",
				Description: "new description",
				Status:      "in_progress",
				IntStatus:   entity.TodoStatusInProgress,
			},
			mockFunc: func(r *mocks.TodoRepository) {
				r.On("FindByID", mock.Anything, uint64(1)).
					Return(nil, errors.New("something error"))
			},
			wantErrMsg: "failed to find todo by id: something error",
		},
		{
			name: "error forbidden",
			request: &model.UpdateTodoRequest{
				ID:          1,
				UserID:      1,
				Title:       "new title",
				Description: "new description",
				Status:      "in_progress",
				IntStatus:   entity.TodoStatusInProgress,
			},
			mockFunc: func(r *mocks.TodoRepository) {
				r.On("FindByID", mock.Anything, uint64(1)).Return(&entity.Todo{
					ID:          1,
					UserID:      2,
					Title:       "title",
					Description: &description,
					Status:      entity.TodoStatusPending,
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil)
			},
			wantErrMsg: "Forbidden",
		},
		{
			name: "error on update",
			request: &model.UpdateTodoRequest{
				ID:          1,
				UserID:      1,
				Title:       "new title",
				Description: "new description",
				Status:      "in_progress",
				IntStatus:   entity.TodoStatusInProgress,
			},
			mockFunc: func(r *mocks.TodoRepository) {
				r.On("FindByID", mock.Anything, uint64(1)).Return(&entity.Todo{
					ID:          1,
					UserID:      1,
					Title:       "title",
					Description: &description,
					Status:      entity.TodoStatusPending,
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil)
				r.On("UpdateByID", mock.Anything, mock.Anything).
					Return(errors.New("something error"))
			},
			wantErrMsg: "failed to update todo by id: something error",
		},
		{
			name: "success",
			request: &model.UpdateTodoRequest{
				ID:          1,
				UserID:      1,
				Title:       "new title",
				Description: "new description",
				Status:      "in_progress",
				IntStatus:   entity.TodoStatusInProgress,
			},
			mockFunc: func(r *mocks.TodoRepository) {
				r.On("FindByID", mock.Anything, uint64(1)).Return(&entity.Todo{
					ID:          1,
					UserID:      1,
					Title:       "title",
					Description: &description,
					Status:      entity.TodoStatusPending,
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil)
				r.On("UpdateByID", mock.Anything, mock.Anything).Return(nil)
			},
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			todoRepository := mocks.NewTodoRepository(s.T())
			usecase := usecase.NewTodoUsecase(s.log, todoRepository)
			tt.mockFunc(todoRepository)

			err := usecase.UpdateByID(s.ctx, tt.request)

			if tt.wantErrMsg != "" {
				s.Equal(tt.wantErrMsg, err.Error())
			} else {
				s.Nil(err)
			}
		})
	}
}

func TestTodoUsecaseSuite(t *testing.T) {
	suite.Run(t, new(TodoUsecaseSuite))
}
