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

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type AuthControllerSuite struct {
	suite.Suite
	log      *zap.Logger
	validate *validator.Validate
	env      *config.Env
}

func (s *AuthControllerSuite) SetupTest() {
	s.log = zap.NewNop()
	s.validate = validator.New()
	s.env = &config.Env{
		AppName:         "api-example",
		AppReadTimeout:  60,
		AppWriteTimeout: 60,
	}
}

func (s *AuthControllerSuite) TestAuthController_Login() {
	tests := []struct {
		name       string
		body       any
		mockFunc   func(a *mocks.AuthUsecase)
		wantStatus int
		wantRes    string
	}{
		{
			name:       "empty body",
			body:       nil,
			mockFunc:   func(a *mocks.AuthUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "invalid body",
			body: map[string]interface{}{
				"foo": 0,
				"bar": "123",
			},
			mockFunc:   func(a *mocks.AuthUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "error on validate body",
			body: map[string]interface{}{
				"username": "",
				"password": "",
			},
			mockFunc:   func(a *mocks.AuthUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "custom error on login",
			body: map[string]interface{}{
				"username": "johndoe",
				"password": "password",
			},
			mockFunc: func(a *mocks.AuthUsecase) {
				a.On("Login", mock.Anything, mock.Anything).
					Return(nil, model.NewCustomError(fiber.StatusNotFound, model.ErrUserNotFound))
			},
			wantStatus: fiber.StatusNotFound,
			wantRes:    `{"errors":[{"code":1002,"message":"Username not found"}],"meta":{"http_status":404}}`,
		},
		{
			name: "unexpected error on login",
			body: map[string]interface{}{
				"username": "johndoe",
				"password": "password",
			},
			mockFunc: func(a *mocks.AuthUsecase) {
				a.On("Login", mock.Anything, mock.Anything).
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
			mockFunc: func(a *mocks.AuthUsecase) {
				a.On("Login", mock.Anything, mock.Anything).Return(&model.LoginResponse{
					AccessToken:  "qwerty-12345",
					RefreshToken: "zxc-123",
				}, nil)
			},
			wantStatus: fiber.StatusOK,
			wantRes:    `{"data":{"access_token":"qwerty-12345","refresh_token":"zxc-123"},"meta":{"http_status":200}}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			au := mocks.NewAuthUsecase(s.T())
			tt.mockFunc(au)

			ac := http.NewAuthController(s.log, s.validate, au)

			app := config.NewFiber(s.env, s.log)
			app.Post("/api/login", ac.Login)

			reqBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			s.Nil(err)

			body, _ := io.ReadAll(resp.Body)
			s.Equal(tt.wantStatus, resp.StatusCode)
			s.Equal(tt.wantRes, strings.TrimSpace(string(body)))
		})
	}
}

func (s *AuthControllerSuite) TestAuthController_Logout() {
	tests := []struct {
		name       string
		body       any
		mockFunc   func(a *mocks.AuthUsecase)
		wantStatus int
		wantRes    string
	}{
		{
			name:       "empty body",
			body:       nil,
			mockFunc:   func(a *mocks.AuthUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "error on validate body",
			body: map[string]interface{}{
				"refresh_token": "",
			},
			mockFunc:   func(a *mocks.AuthUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "error on logout",
			body: map[string]interface{}{
				"refresh_token": "zxc-123",
			},
			mockFunc: func(a *mocks.AuthUsecase) {
				a.On("Logout", mock.Anything, mock.Anything).
					Return(errors.New("something error"))
			},
			wantStatus: fiber.StatusInternalServerError,
			wantRes:    `{"errors":[{"code":1500,"message":"Internal server error"}],"meta":{"http_status":500}}`,
		},
		{
			name: "success",
			body: map[string]interface{}{
				"refresh_token": "zxc-123",
			},
			mockFunc: func(a *mocks.AuthUsecase) {
				a.On("Logout", mock.Anything, mock.Anything).Return(nil)
			},
			wantStatus: fiber.StatusOK,
			wantRes:    `{"message":"Logged out","meta":{"http_status":200}}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			au := mocks.NewAuthUsecase(s.T())
			tt.mockFunc(au)

			ac := http.NewAuthController(s.log, s.validate, au)

			app := config.NewFiber(s.env, s.log)
			app.Use(test.NewAuthMiddleware(1))
			app.Post("/api/logout", ac.Logout)

			reqBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/logout", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			s.Nil(err)

			body, _ := io.ReadAll(resp.Body)
			s.Equal(tt.wantStatus, resp.StatusCode)
			s.Equal(tt.wantRes, strings.TrimSpace(string(body)))
		})
	}
}

func (s *AuthControllerSuite) TestAuthController_RefreshToken() {
	tests := []struct {
		name       string
		body       any
		mockFunc   func(a *mocks.AuthUsecase)
		wantStatus int
		wantRes    string
	}{
		{
			name:       "empty body",
			body:       nil,
			mockFunc:   func(a *mocks.AuthUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "error on validate body",
			body: map[string]interface{}{
				"refresh_token": "",
			},
			mockFunc:   func(a *mocks.AuthUsecase) {},
			wantStatus: fiber.StatusBadRequest,
			wantRes:    `{"errors":[{"code":400,"message":"Bad Request"}],"meta":{"http_status":400}}`,
		},
		{
			name: "error on refresh",
			body: map[string]interface{}{
				"refresh_token": "zxc-123",
			},
			mockFunc: func(a *mocks.AuthUsecase) {
				a.On("Refresh", mock.Anything, mock.Anything).
					Return(nil, errors.New("something error"))
			},
			wantStatus: fiber.StatusInternalServerError,
			wantRes:    `{"errors":[{"code":1500,"message":"Internal server error"}],"meta":{"http_status":500}}`,
		},
		{
			name: "success",
			body: map[string]interface{}{
				"refresh_token": "zxc-123",
			},
			mockFunc: func(a *mocks.AuthUsecase) {
				a.On("Refresh", mock.Anything, mock.Anything).Return(&model.RefreshResponse{
					AccessToken:  "qwerty-12345",
					RefreshToken: "zxc-123",
				}, nil)
			},
			wantStatus: fiber.StatusOK,
			wantRes:    `{"data":{"access_token":"qwerty-12345","refresh_token":"zxc-123"},"meta":{"http_status":200}}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			au := mocks.NewAuthUsecase(s.T())
			tt.mockFunc(au)

			ac := http.NewAuthController(s.log, s.validate, au)

			app := config.NewFiber(s.env, s.log)
			app.Post("/api/refresh-token", ac.RefreshToken)

			reqBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/refresh-token", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			resp, err := app.Test(req)
			s.Nil(err)

			body, _ := io.ReadAll(resp.Body)
			s.Equal(tt.wantStatus, resp.StatusCode)
			s.Equal(tt.wantRes, strings.TrimSpace(string(body)))
		})
	}
}

func TestAuthControllerSuite(t *testing.T) {
	suite.Run(t, new(AuthControllerSuite))
}
