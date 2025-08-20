package usecase_test

import (
	"context"
	"errors"
	"go-api-example/internal/auth"
	"go-api-example/internal/entity"
	"go-api-example/internal/mocks"
	"go-api-example/internal/model"
	"go-api-example/internal/usecase"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthUsecaseSuite struct {
	suite.Suite
	log *zap.Logger
	ctx context.Context
}

type MockFunc func(
	c context.Context,
	rc *mocks.RedisClient,
	jwt *mocks.JWTToken,
	rt *mocks.RefreshToken,
	ur *mocks.UserRepository,
)

func (s *AuthUsecaseSuite) SetupTest() {
	s.log, _ = zap.NewDevelopment()
	s.ctx = context.Background()
}

func (s *AuthUsecaseSuite) TestAuthUsecase_Login() {
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	now := time.Now()

	tests := []struct {
		name       string
		request    *model.LoginRequest
		mockFunc   MockFunc
		wantRes    *model.LoginResponse
		wantErrMsg string
	}{
		{
			name: "error on find by username",
			request: &model.LoginRequest{
				Username: "johndoe",
				Password: "password",
			},
			mockFunc: func(
				c context.Context,
				rc *mocks.RedisClient,
				jwt *mocks.JWTToken,
				rt *mocks.RefreshToken,
				ur *mocks.UserRepository,
			) {
				ur.On("FindByUsername", mock.Anything, "johndoe").
					Return(nil, errors.New("something error"))
			},
			wantRes:    nil,
			wantErrMsg: "failed to find user by username: something error",
		},
		{
			name: "error user not found",
			request: &model.LoginRequest{
				Username: "johndoe",
				Password: "password",
			},
			mockFunc: func(
				c context.Context,
				rc *mocks.RedisClient,
				jwt *mocks.JWTToken,
				rt *mocks.RefreshToken,
				ur *mocks.UserRepository,
			) {
				ur.On("FindByUsername", mock.Anything, "johndoe").
					Return(nil, nil)
			},
			wantRes:    nil,
			wantErrMsg: "Username not found",
		},
		{
			name: "error invalid password",
			request: &model.LoginRequest{
				Username: "johndoe",
				Password: "invalid_password",
			},
			mockFunc: func(
				c context.Context,
				rc *mocks.RedisClient,
				jwt *mocks.JWTToken,
				rt *mocks.RefreshToken,
				ur *mocks.UserRepository,
			) {
				ur.On("FindByUsername", mock.Anything, "johndoe").Return(&entity.User{
					ID:        1,
					Username:  "johndoe",
					Password:  string(passwordHash),
					CreatedAt: now,
					UpdatedAt: now,
				}, nil)
			},
			wantRes:    nil,
			wantErrMsg: "Invalid password",
		},
		{
			name: "error on create jwt token",
			request: &model.LoginRequest{
				Username: "johndoe",
				Password: "password",
			},
			mockFunc: func(
				c context.Context,
				rc *mocks.RedisClient,
				jwt *mocks.JWTToken,
				rt *mocks.RefreshToken,
				ur *mocks.UserRepository,
			) {
				ur.On("FindByUsername", mock.Anything, "johndoe").Return(&entity.User{
					ID:        1,
					Username:  "johndoe",
					Password:  string(passwordHash),
					CreatedAt: now,
					UpdatedAt: now,
				}, nil)
				jwt.On("Create", "1").Return("", errors.New("something error"))
			},
			wantRes:    nil,
			wantErrMsg: "failed to create access token: something error",
		},
		{
			name: "error on set refresh token cache",
			request: &model.LoginRequest{
				Username: "johndoe",
				Password: "password",
			},
			mockFunc: func(
				c context.Context,
				rc *mocks.RedisClient,
				jwt *mocks.JWTToken,
				rt *mocks.RefreshToken,
				ur *mocks.UserRepository,
			) {
				ur.On("FindByUsername", mock.Anything, "johndoe").Return(&entity.User{
					ID:        1,
					Username:  "johndoe",
					Password:  string(passwordHash),
					CreatedAt: now,
					UpdatedAt: now,
				}, nil)
				jwt.On("Create", "1").Return("qwerty-12345", nil)
				rt.On("Create").Return("zxc-123")
				setCmd := redis.NewStatusCmd(s.ctx)
				setCmd.SetErr(errors.New("something error"))
				rc.On("SetEx", mock.Anything, "refresh-token:zxc-123", "1", mock.Anything).
					Return(setCmd)
			},
			wantRes:    nil,
			wantErrMsg: "failed to store refresh token: something error",
		},
		{
			name: "success",
			request: &model.LoginRequest{
				Username: "johndoe",
				Password: "password",
			},
			mockFunc: func(
				c context.Context,
				rc *mocks.RedisClient,
				jwt *mocks.JWTToken,
				rt *mocks.RefreshToken,
				ur *mocks.UserRepository,
			) {
				ur.On("FindByUsername", mock.Anything, "johndoe").Return(&entity.User{
					ID:        1,
					Username:  "johndoe",
					Password:  string(passwordHash),
					CreatedAt: now,
					UpdatedAt: now,
				}, nil)
				jwt.On("Create", "1").Return("qwerty-12345", nil)
				rt.On("Create").Return("zxc-123")
				setCmd := redis.NewStatusCmd(s.ctx)
				rc.On("SetEx", mock.Anything, "refresh-token:zxc-123", "1", mock.Anything).
					Return(setCmd)
			},
			wantRes: &model.LoginResponse{
				AccessToken:  "qwerty-12345",
				RefreshToken: "zxc-123",
			},
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			rc := mocks.NewRedisClient(s.T())
			jwt := mocks.NewJWTToken(s.T())
			rt := mocks.NewRefreshToken(s.T())
			ur := mocks.NewUserRepository(s.T())
			usecase := usecase.NewAuthUsecase(s.log, rc, jwt, rt, ur)
			tt.mockFunc(s.ctx, rc, jwt, rt, ur)

			res, err := usecase.Login(s.ctx, tt.request)

			if tt.wantErrMsg != "" {
				s.Nil(res)
				s.Equal(tt.wantErrMsg, err.Error())
			} else {
				s.Equal(*tt.wantRes, *res)
				s.Nil(err)
			}
		})
	}
}

func (s *AuthUsecaseSuite) TestAuthUsecase_Logout() {
	now := time.Now()

	tests := []struct {
		name       string
		request    *model.LogoutRequest
		mockFunc   func(ctx context.Context, rc *mocks.RedisClient)
		wantErrMsg string
	}{
		{
			name: "refresh token not found",
			request: &model.LogoutRequest{
				Claims: &auth.JWTClaims{
					UserID: "1",
				},
				RefreshToken: "zxc-123",
			},
			mockFunc: func(ctx context.Context, rc *mocks.RedisClient) {
				getCmd := redis.NewStringCmd(s.ctx)
				getCmd.SetErr(redis.Nil)
				rc.On("Get", mock.Anything, "refresh-token:zxc-123").Return(getCmd)
			},
			wantErrMsg: "Invalid refresh token",
		},
		{
			name: "error on get refresh token cache",
			request: &model.LogoutRequest{
				Claims: &auth.JWTClaims{
					UserID: "1",
				},
				RefreshToken: "zxc-123",
			},
			mockFunc: func(ctx context.Context, rc *mocks.RedisClient) {
				getCmd := redis.NewStringCmd(s.ctx)
				getCmd.SetErr(errors.New("something error"))
				rc.On("Get", mock.Anything, "refresh-token:zxc-123").Return(getCmd)
			},
			wantErrMsg: "failed to get refresh token: something error",
		},
		{
			name: "error on invalid user id",
			request: &model.LogoutRequest{
				Claims: &auth.JWTClaims{
					UserID: "1",
				},
				RefreshToken: "zxc-123",
			},
			mockFunc: func(ctx context.Context, rc *mocks.RedisClient) {
				getCmd := redis.NewStringCmd(s.ctx)
				getCmd.SetVal("2")
				rc.On("Get", mock.Anything, "refresh-token:zxc-123").Return(getCmd)
			},
			wantErrMsg: "Invalid logout session",
		},
		{
			name: "error on set revoke token cache",
			request: &model.LogoutRequest{
				Claims: &auth.JWTClaims{
					UserID: "1",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Minute)),
						ID:        "asd-789",
					},
				},
				RefreshToken: "zxc-123",
			},
			mockFunc: func(ctx context.Context, rc *mocks.RedisClient) {
				getCmd := redis.NewStringCmd(s.ctx)
				getCmd.SetVal("1")
				rc.On("Get", mock.Anything, "refresh-token:zxc-123").Return(getCmd)

				setCmd := redis.NewStatusCmd(s.ctx)
				setCmd.SetErr(errors.New("something error"))
				rc.On("SetEx", mock.Anything, "revoke-jwt-token:asd-789", "true", mock.Anything).
					Return(setCmd)
			},
			wantErrMsg: "failed to set revoke token",
		},
		{
			name: "success",
			request: &model.LogoutRequest{
				Claims: &auth.JWTClaims{
					UserID: "1",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Minute)),
						ID:        "asd-789",
					},
				},
				RefreshToken: "zxc-123",
			},
			mockFunc: func(ctx context.Context, rc *mocks.RedisClient) {
				getCmd := redis.NewStringCmd(s.ctx)
				getCmd.SetVal("1")
				rc.On("Get", mock.Anything, "refresh-token:zxc-123").Return(getCmd)

				setCmd := redis.NewStatusCmd(s.ctx)
				rc.On("SetEx", mock.Anything, "revoke-jwt-token:asd-789", "true", mock.Anything).
					Return(setCmd)

				delCmd := redis.NewIntCmd(s.ctx)
				rc.On("Del", mock.Anything, "refresh-token:zxc-123").
					Return(delCmd)
			},
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			rc := mocks.NewRedisClient(s.T())
			jwt := mocks.NewJWTToken(s.T())
			rt := mocks.NewRefreshToken(s.T())
			ur := mocks.NewUserRepository(s.T())
			usecase := usecase.NewAuthUsecase(s.log, rc, jwt, rt, ur)
			tt.mockFunc(s.ctx, rc)

			err := usecase.Logout(s.ctx, tt.request)

			if tt.wantErrMsg != "" {
				s.Equal(tt.wantErrMsg, err.Error())
			} else {
				s.Nil(err)
			}
		})
	}
}

func (s *AuthUsecaseSuite) TestAuthUsecase_Refresh() {
	tests := []struct {
		name       string
		request    *model.RefreshRequest
		mockFunc   MockFunc
		wantRes    *model.RefreshResponse
		wantErrMsg string
	}{
		{
			name: "refresh token cache not found",
			request: &model.RefreshRequest{
				RefreshToken: "zxc-123",
			},
			mockFunc: func(
				c context.Context,
				rc *mocks.RedisClient,
				jwt *mocks.JWTToken,
				rt *mocks.RefreshToken,
				ur *mocks.UserRepository,
			) {
				getCmd := redis.NewStringCmd(s.ctx)
				getCmd.SetErr(redis.Nil)
				rc.On("Get", mock.Anything, "refresh-token:zxc-123").
					Return(getCmd)
			},
			wantRes:    nil,
			wantErrMsg: "Invalid refresh token",
		},
		{
			name: "error on get refresh token cache",
			request: &model.RefreshRequest{
				RefreshToken: "zxc-123",
			},
			mockFunc: func(
				c context.Context,
				rc *mocks.RedisClient,
				jwt *mocks.JWTToken,
				rt *mocks.RefreshToken,
				ur *mocks.UserRepository,
			) {
				getCmd := redis.NewStringCmd(s.ctx)
				getCmd.SetErr(errors.New("something error"))
				rc.On("Get", mock.Anything, "refresh-token:zxc-123").
					Return(getCmd)
			},
			wantRes:    nil,
			wantErrMsg: "failed to get refresh token: something error",
		},
		{
			name: "refresh token has invalid user id",
			request: &model.RefreshRequest{
				RefreshToken: "zxc-123",
			},
			mockFunc: func(
				c context.Context,
				rc *mocks.RedisClient,
				jwt *mocks.JWTToken,
				rt *mocks.RefreshToken,
				ur *mocks.UserRepository,
			) {
				getCmd := redis.NewStringCmd(s.ctx)
				getCmd.SetVal("")
				rc.On("Get", mock.Anything, "refresh-token:zxc-123").
					Return(getCmd)
			},
			wantRes:    nil,
			wantErrMsg: "Invalid user id",
		},
		{
			name: "error on create jwt token",
			request: &model.RefreshRequest{
				RefreshToken: "zxc-123",
			},
			mockFunc: func(
				c context.Context,
				rc *mocks.RedisClient,
				jwt *mocks.JWTToken,
				rt *mocks.RefreshToken,
				ur *mocks.UserRepository,
			) {
				getCmd := redis.NewStringCmd(s.ctx)
				getCmd.SetVal("1")
				rc.On("Get", mock.Anything, "refresh-token:zxc-123").
					Return(getCmd)
				jwt.On("Create", "1").Return("", errors.New("something error"))
			},
			wantRes:    nil,
			wantErrMsg: "failed to create access token: something error",
		},
		{
			name: "error on set new refresh token cache",
			request: &model.RefreshRequest{
				RefreshToken: "zxc-123",
			},
			mockFunc: func(
				c context.Context,
				rc *mocks.RedisClient,
				jwt *mocks.JWTToken,
				rt *mocks.RefreshToken,
				ur *mocks.UserRepository,
			) {
				getCmd := redis.NewStringCmd(s.ctx)
				getCmd.SetVal("1")
				rc.On("Get", mock.Anything, "refresh-token:zxc-123").
					Return(getCmd)
				jwt.On("Create", "1").Return("", nil)
				rt.On("Create").Return("asd-123")
				setCmd := redis.NewStatusCmd(s.ctx)
				setCmd.SetErr(errors.New("something error"))
				rc.On("SetEx", mock.Anything, "refresh-token:asd-123", "1", mock.Anything).
					Return(setCmd)
			},
			wantRes:    nil,
			wantErrMsg: "failed to store refresh token: something error",
		},
		{
			name: "success",
			request: &model.RefreshRequest{
				RefreshToken: "zxc-123",
			},
			mockFunc: func(
				c context.Context,
				rc *mocks.RedisClient,
				jwt *mocks.JWTToken,
				rt *mocks.RefreshToken,
				ur *mocks.UserRepository,
			) {
				getCmd := redis.NewStringCmd(s.ctx)
				getCmd.SetVal("1")
				rc.On("Get", mock.Anything, "refresh-token:zxc-123").
					Return(getCmd)
				jwt.On("Create", "1").Return("tyuip-12345", nil)
				rt.On("Create").Return("asd-123")
				setCmd := redis.NewStatusCmd(s.ctx)
				setCmd.SetErr(nil)
				rc.On("SetEx", mock.Anything, "refresh-token:asd-123", "1", mock.Anything).
					Return(setCmd)
				delCmd := redis.NewIntCmd(s.ctx)
				rc.On("Del", mock.Anything, "refresh-token:zxc-123").
					Return(delCmd)
			},
			wantRes: &model.RefreshResponse{
				AccessToken:  "tyuip-12345",
				RefreshToken: "asd-123",
			},
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			rc := mocks.NewRedisClient(s.T())
			jwt := mocks.NewJWTToken(s.T())
			rt := mocks.NewRefreshToken(s.T())
			ur := mocks.NewUserRepository(s.T())
			usecase := usecase.NewAuthUsecase(s.log, rc, jwt, rt, ur)
			tt.mockFunc(s.ctx, rc, jwt, rt, ur)

			res, err := usecase.Refresh(s.ctx, tt.request)

			if tt.wantErrMsg != "" {
				s.Nil(res)
				s.Equal(tt.wantErrMsg, err.Error())
			} else {
				s.Equal(*tt.wantRes, *res)
				s.Nil(err)
			}
		})
	}
}

func TestAuthUsecaseSuite(t *testing.T) {
	suite.Run(t, new(AuthUsecaseSuite))
}
