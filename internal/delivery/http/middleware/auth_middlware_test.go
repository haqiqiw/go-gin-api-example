package middleware_test

import (
	"context"
	"errors"
	"go-api-example/internal/auth"
	"go-api-example/internal/config"
	"go-api-example/internal/delivery/http/middleware"
	"go-api-example/internal/mocks"
	"go-api-example/internal/model"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type AuthMiddlewareSuite struct {
	suite.Suite
	log *zap.Logger
	env *config.Env
}

func (s *AuthMiddlewareSuite) SetupTest() {
	s.log = zap.NewNop()
	s.env = &config.Env{
		AppName:         "api-example",
		AppReadTimeout:  60,
		AppWriteTimeout: 60,
	}
}

func (s *AuthMiddlewareSuite) TestAuthMiddleware_Handler() {
	tests := []struct {
		name       string
		authToken  string
		mockFunc   func(rc *mocks.RedisClient, j *mocks.JWTToken)
		wantStatus int
		wantRes    string
	}{
		{
			name:       "empty auth token",
			authToken:  "",
			mockFunc:   func(rc *mocks.RedisClient, j *mocks.JWTToken) {},
			wantStatus: fiber.StatusUnauthorized,
			wantRes:    `{"errors":[{"code":401,"message":"missing or invalid auth header"}],"meta":{"http_status":401}}`,
		},
		{
			name:      "invalid auth token",
			authToken: "Bearer dummy-token",
			mockFunc: func(rc *mocks.RedisClient, j *mocks.JWTToken) {
				j.On("Parse", "dummy-token").
					Return(nil, errors.New("something error"))
			},
			wantStatus: fiber.StatusUnauthorized,
			wantRes:    `{"errors":[{"code":401,"message":"invalid auth token"}],"meta":{"http_status":401}}`,
		},
		{
			name:      "error on get revoked token cache",
			authToken: "Bearer dummy-token",
			mockFunc: func(rc *mocks.RedisClient, j *mocks.JWTToken) {
				j.On("Parse", "dummy-token").Return(&auth.JWTClaims{
					UserID: "1",
					RegisteredClaims: jwt.RegisteredClaims{
						ID: "zxc-123",
					},
				}, nil)
				existsCmd := redis.NewIntCmd(context.Background())
				existsCmd.SetErr(errors.New("something error"))
				rc.On("Exists", mock.Anything, "revoke-jwt-token:zxc-123").Return(existsCmd)
			},
			wantStatus: fiber.StatusUnauthorized,
			wantRes:    `{"errors":[{"code":401,"message":"invalid auth token"}],"meta":{"http_status":401}}`,
		},
		{
			name:      "revoked auth token",
			authToken: "Bearer dummy-token",
			mockFunc: func(rc *mocks.RedisClient, j *mocks.JWTToken) {
				j.On("Parse", "dummy-token").Return(&auth.JWTClaims{
					UserID: "1",
					RegisteredClaims: jwt.RegisteredClaims{
						ID: "zxc-123",
					},
				}, nil)
				existsCmd := redis.NewIntCmd(context.Background())
				existsCmd.SetVal(1)
				rc.On("Exists", mock.Anything, "revoke-jwt-token:zxc-123").Return(existsCmd)
			},
			wantStatus: fiber.StatusUnauthorized,
			wantRes:    `{"errors":[{"code":401,"message":"token revoked"}],"meta":{"http_status":401}}`,
		},
		{
			name:      "success",
			authToken: "Bearer dummy-token",
			mockFunc: func(rc *mocks.RedisClient, j *mocks.JWTToken) {
				j.On("Parse", "dummy-token").Return(&auth.JWTClaims{
					UserID: "1",
					RegisteredClaims: jwt.RegisteredClaims{
						ID: "zxc-123",
					},
				}, nil)
				existsCmd := redis.NewIntCmd(context.Background())
				existsCmd.SetVal(0)
				rc.On("Exists", mock.Anything, "revoke-jwt-token:zxc-123").Return(existsCmd)
			},
			wantStatus: fiber.StatusOK,
			wantRes:    `{"data":{"user_id":"1"},"meta":{"http_status":200}}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			rc := mocks.NewRedisClient(s.T())
			jwt := mocks.NewJWTToken(s.T())
			tt.mockFunc(rc, jwt)

			authMw := middleware.NewAuthMiddleware(s.log, rc, jwt)

			app := config.NewFiber(s.env, s.log)
			app.Use(authMw)
			app.Get("/", func(ctx *fiber.Ctx) error {
				claims, err := middleware.GetJWTClaims(ctx)
				if err != nil {
					return fiber.ErrUnauthorized
				}

				data := struct {
					UserID string `json:"user_id"`
				}{
					UserID: claims.UserID,
				}

				return ctx.
					Status(fiber.StatusOK).
					JSON(model.NewSuccessResponse(data, fiber.StatusOK))
			})

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", tt.authToken)

			resp, err := app.Test(req)
			s.Nil(err)

			body, _ := io.ReadAll(resp.Body)
			s.Equal(tt.wantStatus, resp.StatusCode)
			s.Equal(tt.wantRes, strings.TrimSpace(string(body)))
		})
	}
}

func TestAuthMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareSuite))
}
