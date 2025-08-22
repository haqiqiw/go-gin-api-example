package middleware_test

import (
	"context"
	"errors"
	"go-api-example/internal/auth"
	"go-api-example/internal/config"
	"go-api-example/internal/delivery/http/middleware"
	"go-api-example/internal/mocks"
	"go-api-example/internal/model"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type AuthMiddlewareSuite struct {
	suite.Suite
	log *zap.Logger
}

func (s *AuthMiddlewareSuite) SetupTest() {
	s.log = zap.NewNop()
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
			wantStatus: http.StatusUnauthorized,
			wantRes:    `{"errors":[{"code":104,"message":"missing or invalid auth header"}],"meta":{"http_status":401}}`,
		},
		{
			name:      "invalid auth token",
			authToken: "Bearer dummy-token",
			mockFunc: func(rc *mocks.RedisClient, j *mocks.JWTToken) {
				j.On("Parse", "dummy-token").
					Return(nil, errors.New("something error"))
			},
			wantStatus: http.StatusUnauthorized,
			wantRes:    `{"errors":[{"code":105,"message":"invalid auth token"}],"meta":{"http_status":401}}`,
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
			wantStatus: http.StatusUnauthorized,
			wantRes:    `{"errors":[{"code":105,"message":"invalid auth token"}],"meta":{"http_status":401}}`,
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
			wantStatus: http.StatusUnauthorized,
			wantRes:    `{"errors":[{"code":106,"message":"token revoked"}],"meta":{"http_status":401}}`,
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
			wantStatus: http.StatusOK,
			wantRes:    `{"data":{"user_id":"1"},"meta":{"http_status":200}}`,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			rc := mocks.NewRedisClient(s.T())
			jwt := mocks.NewJWTToken(s.T())
			tt.mockFunc(rc, jwt)

			authMw := middleware.NewAuthMiddleware(s.log, rc, jwt)

			app := config.NewGin(s.log)
			app.Use(authMw)
			app.GET("/", func(ctx *gin.Context) {
				claims, err := middleware.GetJWTClaims(ctx)
				if err != nil {
					ctx.Error(err)
					return
				}

				data := struct {
					UserID string `json:"user_id"`
				}{
					UserID: claims.UserID,
				}

				ctx.JSON(
					http.StatusOK,
					model.NewSuccessResponse(data, http.StatusOK),
				)
			})

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", tt.authToken)

			rec := httptest.NewRecorder()
			app.ServeHTTP(rec, req)

			s.Equal(tt.wantStatus, rec.Code)
			s.Equal(tt.wantRes, strings.TrimSpace(rec.Body.String()))
		})
	}
}

func TestAuthMiddlewareSuite(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareSuite))
}
