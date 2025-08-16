package usecase

import (
	"context"
	"errors"
	"fmt"
	"go-api-example/internal/auth"
	"go-api-example/internal/model"
	"go-api-example/internal/storage"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type authUsecase struct {
	Log            *zap.Logger
	RedisClient    storage.RedisClient
	JWTToken       auth.JWTToken
	RefreshToken   auth.RefreshToken
	UserRepository UserRepository
}

func NewAuthUsecase(log *zap.Logger, redisClient storage.RedisClient, jwtToken auth.JWTToken,
	refreshToken auth.RefreshToken, userRepository UserRepository) AuthUsecase {
	return &authUsecase{
		Log:            log,
		RedisClient:    redisClient,
		JWTToken:       jwtToken,
		RefreshToken:   refreshToken,
		UserRepository: userRepository,
	}
}

func (c *authUsecase) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := c.UserRepository.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}
	if user == nil {
		return nil, model.NewCustomError(fiber.StatusNotFound, model.ErrInvalidPassword)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return nil, model.NewCustomError(fiber.StatusUnauthorized, model.ErrInvalidPassword)
	}

	accessToken, err := c.JWTToken.Create(fmt.Sprint(user.ID))
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	refreshToken := c.RefreshToken.Create()
	refreshKey := fmt.Sprintf("%s:%s", auth.PrefixRefreshKey, refreshToken)

	err = c.RedisClient.SetEx(ctx, refreshKey, fmt.Sprint(user.ID), auth.RefreshTTL).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (c *authUsecase) Logout(ctx context.Context, req *model.LogoutRequest) error {
	refreshKey := fmt.Sprintf("%s:%s", auth.PrefixRefreshKey, req.RefreshToken)

	userID, err := c.RedisClient.Get(ctx, refreshKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return model.NewCustomError(fiber.StatusUnauthorized, model.ErrInvalidRefreshToken)
		} else {
			return fmt.Errorf("failed to get refresh token: %w", err)
		}
	}

	if userID != req.Claims.UserID {
		return model.NewCustomError(fiber.StatusUnauthorized, model.ErrInvalidLogoutSession)
	}

	revokeKey := fmt.Sprintf("%s:%s", auth.PrefixRevokeKey, req.Claims.ID)
	revokeTTL := time.Until(req.Claims.ExpiresAt.Time)
	err = c.RedisClient.SetEx(ctx, revokeKey, "true", revokeTTL).Err()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to set revoke token")
	}

	_ = c.RedisClient.Del(ctx, refreshKey)

	return nil
}

func (c *authUsecase) Refresh(ctx context.Context, req *model.RefreshRequest) (*model.RefreshResponse, error) {
	oldRefreshKey := fmt.Sprintf("%s:%s", auth.PrefixRefreshKey, req.RefreshToken)

	userID, err := c.RedisClient.Get(ctx, oldRefreshKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, model.NewCustomError(fiber.StatusUnauthorized, model.ErrInvalidRefreshToken)
		} else {
			return nil, fmt.Errorf("failed to get refresh token: %w", err)
		}
	}

	if len(userID) == 0 {
		return nil, model.NewCustomError(fiber.StatusUnprocessableEntity, model.ErrInvalidUserID)
	}

	newAccessToken, err := c.JWTToken.Create(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create access token: %w", err)
	}

	newRefreshToken := c.RefreshToken.Create()
	newRefreshKey := fmt.Sprintf("%s:%s", auth.PrefixRefreshKey, newRefreshToken)

	err = c.RedisClient.SetEx(ctx, newRefreshKey, userID, auth.RefreshTTL).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	_ = c.RedisClient.Del(ctx, oldRefreshKey)

	return &model.RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}
