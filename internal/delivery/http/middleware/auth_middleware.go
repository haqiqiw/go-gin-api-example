package middleware

import (
	"errors"
	"fmt"
	"go-api-example/internal/auth"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewAuthMiddleware(logger *zap.Logger, redisClient *redis.Client, jwtToken auth.JWTToken) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authHeader := ctx.Get("Authorization")
		if authHeader == "" && !strings.HasPrefix(authHeader, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "missing or invalid auth header")
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtToken.Parse(token)
		if err != nil {
			logger.Warn(err.Error(),
				zap.Any("request_id", GetRequestID(ctx)),
				zap.Any("path", ctx.Path()),
				zap.Any("method", ctx.Method()),
			)
			return fiber.NewError(fiber.StatusUnauthorized, "invalid auth token")
		}

		revokeKey := fmt.Sprintf("%s:%s", auth.PrefixRevokeKey, claims.ID)
		exists, err := redisClient.Exists(ctx.UserContext(), revokeKey).Result()
		if err != nil {
			logger.Warn(err.Error(),
				zap.Any("request_id", GetRequestID(ctx)),
				zap.Any("path", ctx.Path()),
				zap.Any("method", ctx.Method()),
			)
			return fiber.NewError(fiber.StatusUnauthorized, "invalid auth token")
		}

		if exists == 1 {
			return fiber.NewError(fiber.StatusUnauthorized, "token revoked")
		}

		ctx.Locals("claims", claims)

		return ctx.Next()
	}
}

func GetJWTClaims(ctx *fiber.Ctx) (*auth.JWTClaims, error) {
	claims, ok := ctx.Locals("claims").(*auth.JWTClaims)
	if !ok {
		return nil, errors.New("invalid jwt claims")
	}

	return claims, nil
}
