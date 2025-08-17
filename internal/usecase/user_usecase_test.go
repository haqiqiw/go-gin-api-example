package usecase_test

import (
	"context"
	"errors"
	"go-api-example/internal/entity"
	"go-api-example/internal/messaging"
	"go-api-example/internal/mocks"
	"go-api-example/internal/model"
	"go-api-example/internal/usecase"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type UserUsecaseSuite struct {
	suite.Suite
	log          *zap.Logger
	userProducer *messaging.UserProducer
	ctx          context.Context
}

func (s *UserUsecaseSuite) SetupTest() {
	s.log, _ = zap.NewDevelopment()
	producer := mocks.NewKafkaProducer(s.T())
	s.userProducer = messaging.NewUserProducer(
		s.log, producer, "user-registered",
	)
	s.ctx = context.Background()
}

func (s *UserUsecaseSuite) TestUserRepository_Create() {
	now := time.Now()

	tests := []struct {
		name       string
		request    *model.CreateUserRequest
		mockFunc   func(tx *mocks.Transactioner, r *mocks.UserRepository)
		wantUser   *model.UserResponse
		wantErrMsg string
	}{
		{
			name: "error on count",
			request: &model.CreateUserRequest{
				Username: "johndoe",
				Password: "password",
			},
			mockFunc: func(tx *mocks.Transactioner, r *mocks.UserRepository) {
				r.On("CountByUsername", mock.Anything, "johndoe").
					Return(0, errors.New("something error"))
			},
			wantUser:   nil,
			wantErrMsg: "failed to count by username: something error",
		},
		{
			name: "error on duplicate username",
			request: &model.CreateUserRequest{
				Username: "johndoe",
				Password: "password",
			},
			mockFunc: func(tx *mocks.Transactioner, r *mocks.UserRepository) {
				r.On("CountByUsername", mock.Anything, "johndoe").
					Return(1, nil)
			},
			wantUser:   nil,
			wantErrMsg: "Username already exist",
		},
		{
			name: "error on create",
			request: &model.CreateUserRequest{
				Username: "johndoe",
				Password: "password",
			},
			mockFunc: func(tx *mocks.Transactioner, r *mocks.UserRepository) {
				r.On("CountByUsername", mock.Anything, "johndoe").
					Return(0, nil)
				tx.On("Do", mock.Anything, mock.Anything).
					Return(errors.New("something error"))
			},
			wantUser:   nil,
			wantErrMsg: "something error",
		},
		{
			name: "success",
			request: &model.CreateUserRequest{
				Username: "johndoe",
				Password: "password",
			},
			mockFunc: func(tx *mocks.Transactioner, r *mocks.UserRepository) {
				r.On("CountByUsername", mock.Anything, "johndoe").
					Return(0, nil)
				tx.On("Do", mock.Anything, mock.Anything).
					Return(nil)
			},
			wantUser: &model.UserResponse{
				ID:        1,
				Username:  "johndoe",
				CreatedAt: now.Format(time.RFC3339),
				UpdatedAt: now.Format(time.RFC3339),
			},
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tx := mocks.NewTransactioner(s.T())
			userRepository := mocks.NewUserRepository(s.T())
			usecase := usecase.NewUserUsecase(s.log, tx, s.userProducer, userRepository)
			tt.mockFunc(tx, userRepository)

			_, err := usecase.Create(s.ctx, tt.request)

			if tt.wantErrMsg != "" {
				s.Equal(tt.wantErrMsg, err.Error())
			} else {
				s.Nil(err)
			}
		})
	}
}

func (s *UserUsecaseSuite) TestUserRepository_List() {
	now := time.Now()

	tests := []struct {
		name       string
		request    *model.SearchUserRequest
		mockFunc   func(r *mocks.UserRepository)
		wantUsers  []model.UserResponse
		wantTotal  int
		wantErrMsg string
	}{
		{
			name: "error on find",
			request: &model.SearchUserRequest{
				Limit:  10,
				Offset: 0,
			},
			mockFunc: func(r *mocks.UserRepository) {
				r.On("List", mock.Anything, mock.Anything).
					Return(nil, 0, errors.New("something error"))
			},
			wantUsers:  []model.UserResponse{},
			wantTotal:  0,
			wantErrMsg: "failed to get users: something error",
		},
		{
			name: "success",
			request: &model.SearchUserRequest{
				Limit:  10,
				Offset: 0,
			},
			mockFunc: func(r *mocks.UserRepository) {
				r.On("List", mock.Anything, mock.Anything).Return([]entity.User{
					{
						ID:        1,
						Username:  "johndoe",
						Password:  "password",
						CreatedAt: now,
						UpdatedAt: now,
					},
				}, 1, nil)
			},
			wantUsers: []model.UserResponse{
				{
					ID:        1,
					Username:  "johndoe",
					CreatedAt: now.Format(time.RFC3339),
					UpdatedAt: now.Format(time.RFC3339),
				},
			},
			wantTotal:  1,
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tx := mocks.NewTransactioner(s.T())
			userRepository := mocks.NewUserRepository(s.T())
			usecase := usecase.NewUserUsecase(s.log, tx, s.userProducer, userRepository)
			tt.mockFunc(userRepository)

			res, total, err := usecase.List(s.ctx, tt.request)

			if tt.wantErrMsg != "" {
				s.Empty(res)
				s.Equal(tt.wantErrMsg, err.Error())
			} else {
				s.Len(res, 1)
				s.Equal(tt.wantUsers[0], res[0])
				s.Nil(err)
			}
			s.Equal(tt.wantTotal, total)
		})
	}
}

func (s *UserUsecaseSuite) TestUserRepository_FindByID() {
	now := time.Now()

	tests := []struct {
		name       string
		request    *model.GetUserRequest
		mockFunc   func(r *mocks.UserRepository)
		wantUser   *model.UserResponse
		wantErrMsg string
	}{
		{
			name: "error on find",
			request: &model.GetUserRequest{
				ID: 1,
			},
			mockFunc: func(r *mocks.UserRepository) {
				r.On("FindByID", mock.Anything, uint64(1)).
					Return(nil, errors.New("something error"))
			},
			wantUser:   nil,
			wantErrMsg: "failed to find user by id: something error",
		},
		{
			name: "success",
			request: &model.GetUserRequest{
				ID: 1,
			},
			mockFunc: func(r *mocks.UserRepository) {
				r.On("FindByID", mock.Anything, uint64(1)).Return(&entity.User{
					ID:        1,
					Username:  "johndoe",
					Password:  "password",
					CreatedAt: now,
					UpdatedAt: now,
				}, nil)
			},
			wantUser: &model.UserResponse{
				ID:        1,
				Username:  "johndoe",
				CreatedAt: now.Format(time.RFC3339),
				UpdatedAt: now.Format(time.RFC3339),
			},
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tx := mocks.NewTransactioner(s.T())
			userRepository := mocks.NewUserRepository(s.T())
			usecase := usecase.NewUserUsecase(s.log, tx, s.userProducer, userRepository)
			tt.mockFunc(userRepository)

			res, err := usecase.FindByID(s.ctx, tt.request)

			if tt.wantErrMsg != "" {
				s.Nil(res)
				s.Equal(tt.wantErrMsg, err.Error())
			} else {
				s.Equal(*tt.wantUser, *res)
				s.Nil(err)
			}
		})
	}
}

func (s *UserUsecaseSuite) TestUserRepository_UpdateByID() {
	oldPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("old_password"), bcrypt.DefaultCost)

	now := time.Now()

	tests := []struct {
		name       string
		request    *model.UpdateUserRequest
		mockFunc   func(r *mocks.UserRepository)
		wantErrMsg string
	}{
		{
			name: "error on find",
			request: &model.UpdateUserRequest{
				ID:          1,
				OldPassword: "old_password",
				NewPassword: "new_password",
			},
			mockFunc: func(r *mocks.UserRepository) {
				r.On("FindByID", mock.Anything, uint64(1)).
					Return(nil, errors.New("something error"))
			},
			wantErrMsg: "failed to find user by id: something error",
		},
		{
			name: "error on compare hash and password",
			request: &model.UpdateUserRequest{
				ID:          1,
				OldPassword: "old_password",
				NewPassword: "new_password",
			},
			mockFunc: func(r *mocks.UserRepository) {
				r.On("FindByID", mock.Anything, uint64(1)).Return(&entity.User{
					ID:        1,
					Username:  "johndoe",
					Password:  "invalid_hash",
					CreatedAt: now,
					UpdatedAt: now,
				}, nil)
			},
			wantErrMsg: "Invalid old password",
		},
		{
			name: "error on update",
			request: &model.UpdateUserRequest{
				ID:          1,
				OldPassword: "old_password",
				NewPassword: "new_password",
			},
			mockFunc: func(r *mocks.UserRepository) {
				r.On("FindByID", mock.Anything, uint64(1)).Return(&entity.User{
					ID:        1,
					Username:  "johndoe",
					Password:  string(oldPasswordHash),
					CreatedAt: now,
					UpdatedAt: now,
				}, nil)
				r.On("UpdateByID", mock.Anything, mock.Anything).
					Return(errors.New("something error"))
			},
			wantErrMsg: "failed to update user by id: something error",
		},
		{
			name: "success",
			request: &model.UpdateUserRequest{
				ID:          1,
				OldPassword: "old_password",
				NewPassword: "new_password",
			},
			mockFunc: func(r *mocks.UserRepository) {
				r.On("FindByID", mock.Anything, uint64(1)).Return(&entity.User{
					ID:        1,
					Username:  "johndoe",
					Password:  string(oldPasswordHash),
					CreatedAt: now,
					UpdatedAt: now,
				}, nil)
				r.On("UpdateByID", mock.Anything, mock.Anything).Return(nil)
			},
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tx := mocks.NewTransactioner(s.T())
			userRepository := mocks.NewUserRepository(s.T())
			usecase := usecase.NewUserUsecase(s.log, tx, s.userProducer, userRepository)
			tt.mockFunc(userRepository)

			err := usecase.UpdateByID(s.ctx, tt.request)

			if tt.wantErrMsg != "" {
				s.Equal(tt.wantErrMsg, err.Error())
			} else {
				s.Nil(err)
			}
		})
	}
}

func TestUserUsecaseSuite(t *testing.T) {
	suite.Run(t, new(UserUsecaseSuite))
}
