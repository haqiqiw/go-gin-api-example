package usecase

import (
	"context"
	"fmt"
	"go-api-example/internal/db"
	"go-api-example/internal/entity"
	"go-api-example/internal/messaging"
	"go-api-example/internal/model"
	"go-api-example/internal/model/serializer"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	Log            *zap.Logger
	TX             db.Transactioner
	UserProducer   *messaging.UserProducer
	UserRepository UserRepository
}

func NewUserUsecase(log *zap.Logger, tx db.Transactioner, userProducer *messaging.UserProducer, userRepository UserRepository) UserUsecase {
	return &userUsecase{
		Log:            log,
		TX:             tx,
		UserProducer:   userProducer,
		UserRepository: userRepository,
	}
}

func (c *userUsecase) Create(ctx context.Context, req *model.CreateUserRequest) (*model.UserResponse, error) {
	total, err := c.UserRepository.CountByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to count by username: %w", err)
	}

	if total > 0 {
		return nil, model.NewCustomError(fiber.StatusBadRequest, model.ErrUsernameAlreadyExist)
	}

	password, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	user := &entity.User{
		Username: req.Username,
		Password: string(password),
	}

	err = c.TX.Do(ctx, func(exec db.Executor) error {
		txErr := c.UserRepository.Create(ctx, exec, user)
		if txErr != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		event := serializer.UserToEvent(user)
		txErr = c.UserProducer.Send(event)
		if txErr != nil {
			return fmt.Errorf("failed to send user event: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return serializer.UserToResponse(user), nil
}

func (c *userUsecase) List(ctx context.Context, req *model.SearchUserRequest) ([]model.UserResponse, int, error) {
	users, total, err := c.UserRepository.List(ctx, req)
	if err != nil {
		return []model.UserResponse{}, 0, fmt.Errorf("failed to get users: %w", err)
	}

	if len(users) == 0 {
		return []model.UserResponse{}, 0, nil
	}

	return serializer.ListUserToResponse(users), total, nil
}

func (c *userUsecase) FindByID(ctx context.Context, req *model.GetUserRequest) (*model.UserResponse, error) {
	user, err := c.UserRepository.FindByID(ctx, req.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	return serializer.UserToResponse(user), nil
}

func (c *userUsecase) UpdateByID(ctx context.Context, req *model.UpdateUserRequest) error {
	user, err := c.UserRepository.FindByID(ctx, req.ID)
	if err != nil {
		return fmt.Errorf("failed to find user by id: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		return model.NewCustomError(fiber.StatusBadRequest, model.ErrInvalidOldPassword)
	}

	newPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to generate password: %w", err)
	}

	req.NewPassword = string(newPassword)
	err = c.UserRepository.UpdateByID(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update user by id: %w", err)
	}

	return nil
}
