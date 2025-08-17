package test

import (
	"fmt"
	"go-api-example/internal/auth"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func NewAuthMiddleware(userID uint64) fiber.Handler {
	now := time.Now()
	claims := &auth.JWTClaims{
		UserID: fmt.Sprint(userID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}

	return func(c *fiber.Ctx) error {
		c.Locals("claims", claims)
		return c.Next()
	}
}
