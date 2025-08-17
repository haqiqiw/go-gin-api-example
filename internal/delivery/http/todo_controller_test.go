package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"go-api-example/internal/config"
	"go-api-example/internal/delivery/http"
	"go-api-example/internal/mocks"
	"go-api-example/internal/model"
	"go-api-example/test"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type TodoControllerSuite struct {
	suite.Suite
	log      *zap.Logger
	validate *validator.Validate
	env      *config.Env
}

func (s *TodoControllerSuite) SetupTest() {
	s.log = zap.NewNop()
	s.validate = validator.New()
	s.env = &config.Env{
		AppName:         "api-example",
		AppReadTimeout:  60,
		AppWriteTimeout: 60,
	}
}

func (s *TodoControllerSuite) TestTodoController_Create() {
	tests := []struct {
		name       string
		body       any
		mockFunc   func(a *mocks.TodoUsecase)
		wantStatus int
		wantRes    string
	}{
		{
			name:       "empty body",
			body:       nil,
			mockFunc:   func(a *mocks.TodoUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "error on validate body",
			body: map[string]interface{}{
				"title":       "",
				"description": "",
			},
			mockFunc:   func(a *mocks.TodoUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "error on create",
			body: map[string]interface{}{
				"title":       "dummy title",
				"description": "dummy description",
			},
			mockFunc: func(a *mocks.TodoUsecase) {
				a.On("Create", mock.Anything, mock.Anything).
					Return(nil, errors.New("something error"))
			},
			wantStatus: fiber.StatusInternalServerError,
			wantRes:    `{"errors":[{"code":1500,"message":"Internal server error"}],"meta":{"http_status":500}}`,
		},
		{
			name: "success",
			body: map[string]interface{}{
				"title":       "dummy title",
				"description": "dummy description",
			},
			mockFunc: func(a *mocks.TodoUsecase) {
				now := time.Date(2025, 10, 27, 13, 7, 31, 000, time.UTC)
				a.On("Create", mock.Anything, mock.Anything).Return(&model.TodoResponse{
					ID:          1,
					UserID:      1,
					Title:       "dummy title",
					Description: "dummy description",
					Status:      "pending",
					CreatedAt:   now.Format(time.RFC3339),
					UpdatedAt:   now.Format(time.RFC3339),
				}, nil)
			},
			wantStatus: fiber.StatusCreated,
			wantRes: `{"data":{"id":1,"user_id":1,"title":"dummy title","description":"dummy description",` +
				`"status":"pending","created_at":"2025-10-27T13:07:31Z","updated_at":"2025-10-27T13:07:31Z"},` +
				`"meta":{"http_status":201}}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tu := mocks.NewTodoUsecase(s.T())
			tt.mockFunc(tu)

			tc := http.NewTodoController(s.log, s.validate, tu)

			app := config.NewFiber(s.env, s.log)
			app.Use(test.NewAuthMiddleware(1))
			app.Post("/api/todos", tc.Create)

			reqBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/todos", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			s.Nil(err)

			body, _ := io.ReadAll(resp.Body)
			s.Equal(tt.wantStatus, resp.StatusCode)
			s.Equal(tt.wantRes, strings.TrimSpace(string(body)))
		})
	}
}

func (s *TodoControllerSuite) TestTodoController_Search() {
	tests := []struct {
		name       string
		mockFunc   func(a *mocks.TodoUsecase)
		wantStatus int
		wantRes    string
	}{
		{
			name: "error on list",
			mockFunc: func(a *mocks.TodoUsecase) {
				a.On("List", mock.Anything, mock.Anything).
					Return([]model.TodoResponse{}, 0, errors.New("something error"))
			},
			wantStatus: fiber.StatusInternalServerError,
			wantRes:    `{"errors":[{"code":1500,"message":"Internal server error"}],"meta":{"http_status":500}}`,
		},
		{
			name: "success",
			mockFunc: func(a *mocks.TodoUsecase) {
				now := time.Date(2025, 10, 27, 13, 7, 31, 000, time.UTC)
				a.On("List", mock.Anything, mock.Anything).
					Return([]model.TodoResponse{
						{
							ID:          1,
							UserID:      1,
							Title:       "dummy title",
							Description: "dummy description",
							Status:      "pending",
							CreatedAt:   now.Format(time.RFC3339),
							UpdatedAt:   now.Format(time.RFC3339),
						},
					}, 1, nil)
			},
			wantStatus: fiber.StatusOK,
			wantRes: `{"data":[{"id":1,"user_id":1,"title":"dummy title","description":"dummy description",` +
				`"status":"pending","created_at":"2025-10-27T13:07:31Z","updated_at":"2025-10-27T13:07:31Z"}],` +
				`"meta":{"limit":10,"offset":0,"total":1,"http_status":200}}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tu := mocks.NewTodoUsecase(s.T())
			tt.mockFunc(tu)

			tc := http.NewTodoController(s.log, s.validate, tu)

			app := config.NewFiber(s.env, s.log)
			app.Use(test.NewAuthMiddleware(1))
			app.Get("/api/todos", tc.Search)

			req := httptest.NewRequest("GET", "/api/todos", nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			s.Nil(err)

			body, _ := io.ReadAll(resp.Body)
			s.Equal(tt.wantStatus, resp.StatusCode)
			s.Equal(tt.wantRes, strings.TrimSpace(string(body)))
		})
	}
}

func (s *TodoControllerSuite) TestTodoController_Get() {
	tests := []struct {
		name       string
		mockFunc   func(a *mocks.TodoUsecase)
		wantStatus int
		wantRes    string
	}{
		{
			name: "error on get",
			mockFunc: func(a *mocks.TodoUsecase) {
				a.On("FindByID", mock.Anything, mock.Anything).
					Return(nil, errors.New("something error"))
			},
			wantStatus: fiber.StatusInternalServerError,
			wantRes:    `{"errors":[{"code":1500,"message":"Internal server error"}],"meta":{"http_status":500}}`,
		},
		{
			name: "success",
			mockFunc: func(a *mocks.TodoUsecase) {
				now := time.Date(2025, 10, 27, 13, 7, 31, 000, time.UTC)
				a.On("FindByID", mock.Anything, mock.Anything).Return(&model.TodoResponse{
					ID:          1,
					UserID:      1,
					Title:       "dummy title",
					Description: "dummy description",
					Status:      "pending",
					CreatedAt:   now.Format(time.RFC3339),
					UpdatedAt:   now.Format(time.RFC3339),
				}, nil)
			},
			wantStatus: fiber.StatusOK,
			wantRes: `{"data":{"id":1,"user_id":1,"title":"dummy title","description":"dummy description",` +
				`"status":"pending","created_at":"2025-10-27T13:07:31Z","updated_at":"2025-10-27T13:07:31Z"},` +
				`"meta":{"http_status":200}}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tu := mocks.NewTodoUsecase(s.T())
			tt.mockFunc(tu)

			tc := http.NewTodoController(s.log, s.validate, tu)

			app := config.NewFiber(s.env, s.log)
			app.Use(test.NewAuthMiddleware(1))
			app.Get("/api/todos/:id", tc.Get)

			req := httptest.NewRequest("GET", "/api/todos/1", nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			s.Nil(err)

			body, _ := io.ReadAll(resp.Body)
			s.Equal(tt.wantStatus, resp.StatusCode)
			s.Equal(tt.wantRes, strings.TrimSpace(string(body)))
		})
	}
}

func (s *TodoControllerSuite) TestTodoController_Update() {
	tests := []struct {
		name       string
		body       any
		mockFunc   func(a *mocks.TodoUsecase)
		wantStatus int
		wantRes    string
	}{
		{
			name:       "empty body",
			body:       nil,
			mockFunc:   func(a *mocks.TodoUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "error on validate body",
			body: map[string]interface{}{
				"title":       "",
				"description": "",
				"status":      "",
			},
			mockFunc:   func(a *mocks.TodoUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "error on update",
			body: map[string]interface{}{
				"title":       "dummy title",
				"description": "dummy description",
				"status":      "completed",
			},
			mockFunc: func(a *mocks.TodoUsecase) {
				a.On("UpdateByID", mock.Anything, mock.Anything).
					Return(errors.New("something error"))
			},
			wantStatus: fiber.StatusInternalServerError,
			wantRes:    `{"errors":[{"code":1500,"message":"Internal server error"}],"meta":{"http_status":500}}`,
		},
		{
			name: "success",
			body: map[string]interface{}{
				"title":       "dummy title",
				"description": "dummy description",
				"status":      "completed",
			},
			mockFunc: func(a *mocks.TodoUsecase) {
				a.On("UpdateByID", mock.Anything, mock.Anything).Return(nil)
			},
			wantStatus: fiber.StatusOK,
			wantRes:    `{"message":"Todo updated","meta":{"http_status":200}}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tu := mocks.NewTodoUsecase(s.T())
			tt.mockFunc(tu)

			tc := http.NewTodoController(s.log, s.validate, tu)

			app := config.NewFiber(s.env, s.log)
			app.Use(test.NewAuthMiddleware(1))
			app.Patch("/api/todos/:id", tc.Update)

			reqBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("PATCH", "/api/todos/1", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			s.Nil(err)

			body, _ := io.ReadAll(resp.Body)
			s.Equal(tt.wantStatus, resp.StatusCode)
			s.Equal(tt.wantRes, strings.TrimSpace(string(body)))
		})
	}
}

func TestTodoControllerSuite(t *testing.T) {
	suite.Run(t, new(TodoControllerSuite))
}
