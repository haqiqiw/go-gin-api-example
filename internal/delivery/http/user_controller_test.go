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

type UserControllerSuite struct {
	suite.Suite
	log      *zap.Logger
	validate *validator.Validate
	env      *config.Env
}

func (s *UserControllerSuite) SetupTest() {
	s.log = zap.NewNop()
	s.validate = validator.New()
	s.env = &config.Env{
		AppName:         "api-example",
		AppReadTimeout:  60,
		AppWriteTimeout: 60,
	}
}

func (s *UserControllerSuite) TestUserController_Create() {
	tests := []struct {
		name       string
		body       any
		mockFunc   func(a *mocks.UserUsecase)
		wantStatus int
		wantRes    string
	}{
		{
			name:       "empty body",
			body:       nil,
			mockFunc:   func(a *mocks.UserUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "error on validate body",
			body: map[string]interface{}{
				"username": "",
				"password": "",
			},
			mockFunc:   func(a *mocks.UserUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "error on create",
			body: map[string]interface{}{
				"username": "johndoe",
				"password": "password",
			},
			mockFunc: func(a *mocks.UserUsecase) {
				a.On("Create", mock.Anything, mock.Anything).
					Return(nil, errors.New("something error"))
			},
			wantStatus: fiber.StatusInternalServerError,
			wantRes:    `{"errors":[{"code":1500,"message":"Internal server error"}],"meta":{"http_status":500}}`,
		},
		{
			name: "success",
			body: map[string]interface{}{
				"username": "johndoe",
				"password": "password",
			},
			mockFunc: func(a *mocks.UserUsecase) {
				now := time.Date(2025, 10, 27, 13, 7, 31, 000, time.UTC)
				a.On("Create", mock.Anything, mock.Anything).Return(&model.UserResponse{
					ID:        1,
					Username:  "johndoe",
					CreatedAt: now.Format(time.RFC3339),
					UpdatedAt: now.Format(time.RFC3339),
				}, nil)
			},
			wantStatus: fiber.StatusCreated,
			wantRes: `{"data":{"id":1,"username":"johndoe","created_at":"2025-10-27T13:07:31Z",` +
				`"updated_at":"2025-10-27T13:07:31Z"},"meta":{"http_status":201}}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			uu := mocks.NewUserUsecase(s.T())
			tt.mockFunc(uu)

			uc := http.NewUserController(s.log, s.validate, uu)

			app := config.NewFiber(s.env, s.log)
			app.Post("/api/users", uc.Register)

			reqBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/users", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			s.Nil(err)

			body, _ := io.ReadAll(resp.Body)
			s.Equal(tt.wantStatus, resp.StatusCode)
			s.Equal(tt.wantRes, strings.TrimSpace(string(body)))
		})
	}
}

func (s *UserControllerSuite) TestUserController_Search() {
	tests := []struct {
		name       string
		mockFunc   func(a *mocks.UserUsecase)
		wantStatus int
		wantRes    string
	}{
		{
			name: "error on list",
			mockFunc: func(a *mocks.UserUsecase) {
				a.On("List", mock.Anything, mock.Anything).
					Return([]model.UserResponse{}, 0, errors.New("something error"))
			},
			wantStatus: fiber.StatusInternalServerError,
			wantRes:    `{"errors":[{"code":1500,"message":"Internal server error"}],"meta":{"http_status":500}}`,
		},
		{
			name: "success",
			mockFunc: func(a *mocks.UserUsecase) {
				now := time.Date(2025, 10, 27, 13, 7, 31, 000, time.UTC)
				a.On("List", mock.Anything, mock.Anything).
					Return([]model.UserResponse{
						{
							ID:        1,
							Username:  "johndoe",
							CreatedAt: now.Format(time.RFC3339),
							UpdatedAt: now.Format(time.RFC3339),
						},
					}, 1, nil)
			},
			wantStatus: fiber.StatusOK,
			wantRes: `{"data":[{"id":1,"username":"johndoe","created_at":"2025-10-27T13:07:31Z",` +
				`"updated_at":"2025-10-27T13:07:31Z"}],"meta":{"limit":10,"offset":0,"total":1,"http_status":200}}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			uu := mocks.NewUserUsecase(s.T())
			tt.mockFunc(uu)

			uc := http.NewUserController(s.log, s.validate, uu)

			app := config.NewFiber(s.env, s.log)
			app.Use(test.NewAuthMiddleware(1))
			app.Get("/api/users", uc.Search)

			req := httptest.NewRequest("GET", "/api/users", nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			s.Nil(err)

			body, _ := io.ReadAll(resp.Body)
			s.Equal(tt.wantStatus, resp.StatusCode)
			s.Equal(tt.wantRes, strings.TrimSpace(string(body)))
		})
	}
}

func (s *UserControllerSuite) TestUserController_Get() {
	tests := []struct {
		name       string
		mockFunc   func(a *mocks.UserUsecase)
		wantStatus int
		wantRes    string
	}{
		{
			name: "error on get",
			mockFunc: func(a *mocks.UserUsecase) {
				a.On("FindByID", mock.Anything, mock.Anything).
					Return(nil, errors.New("something error"))
			},
			wantStatus: fiber.StatusInternalServerError,
			wantRes:    `{"errors":[{"code":1500,"message":"Internal server error"}],"meta":{"http_status":500}}`,
		},
		{
			name: "success",
			mockFunc: func(a *mocks.UserUsecase) {
				now := time.Date(2025, 10, 27, 13, 7, 31, 000, time.UTC)
				a.On("FindByID", mock.Anything, mock.Anything).Return(&model.UserResponse{
					ID:        1,
					Username:  "johndoe",
					CreatedAt: now.Format(time.RFC3339),
					UpdatedAt: now.Format(time.RFC3339),
				}, nil)
			},
			wantStatus: fiber.StatusOK,
			wantRes: `{"data":{"id":1,"username":"johndoe","created_at":"2025-10-27T13:07:31Z",` +
				`"updated_at":"2025-10-27T13:07:31Z"},"meta":{"http_status":200}}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			uu := mocks.NewUserUsecase(s.T())
			tt.mockFunc(uu)

			uc := http.NewUserController(s.log, s.validate, uu)

			app := config.NewFiber(s.env, s.log)
			app.Use(test.NewAuthMiddleware(1))
			app.Get("/api/users/me", uc.Me)

			req := httptest.NewRequest("GET", "/api/users/me", nil)
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			s.Nil(err)

			body, _ := io.ReadAll(resp.Body)
			s.Equal(tt.wantStatus, resp.StatusCode)
			s.Equal(tt.wantRes, strings.TrimSpace(string(body)))
		})
	}
}

func (s *UserControllerSuite) TestUserController_Update() {
	tests := []struct {
		name       string
		body       any
		mockFunc   func(a *mocks.UserUsecase)
		wantStatus int
		wantRes    string
	}{
		{
			name:       "empty body",
			body:       nil,
			mockFunc:   func(a *mocks.UserUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "error on validate body",
			body: map[string]interface{}{
				"old_password": "",
				"new_password": "",
			},
			mockFunc:   func(a *mocks.UserUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "error on update",
			body: map[string]interface{}{
				"old_password": "old_password",
				"new_password": "new_password",
			},
			mockFunc: func(a *mocks.UserUsecase) {
				a.On("UpdateByID", mock.Anything, mock.Anything).
					Return(errors.New("something error"))
			},
			wantStatus: fiber.StatusInternalServerError,
			wantRes:    `{"errors":[{"code":1500,"message":"Internal server error"}],"meta":{"http_status":500}}`,
		},
		{
			name: "success",
			body: map[string]interface{}{
				"old_password": "old_password",
				"new_password": "new_password",
			},
			mockFunc: func(a *mocks.UserUsecase) {
				a.On("UpdateByID", mock.Anything, mock.Anything).Return(nil)
			},
			wantStatus: fiber.StatusOK,
			wantRes:    `{"message":"User updated","meta":{"http_status":200}}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			uu := mocks.NewUserUsecase(s.T())
			tt.mockFunc(uu)

			uc := http.NewUserController(s.log, s.validate, uu)

			app := config.NewFiber(s.env, s.log)
			app.Use(test.NewAuthMiddleware(1))
			app.Patch("/api/users/me", uc.Update)

			reqBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("PATCH", "/api/users/me", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			s.Nil(err)

			body, _ := io.ReadAll(resp.Body)
			s.Equal(tt.wantStatus, resp.StatusCode)
			s.Equal(tt.wantRes, strings.TrimSpace(string(body)))
		})
	}
}

func TestUserControllerSuite(t *testing.T) {
	suite.Run(t, new(UserControllerSuite))
}
