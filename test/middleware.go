package test

import (
	"fmt"
	"go-api-example/internal/auth"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func NewAuthMiddleware(userID uint64) gin.HandlerFunc {
	now := time.Now()
	claims := &auth.JWTClaims{
		UserID: fmt.Sprint(userID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(1 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			ID:        uuid.NewString(),
		},
	}

	return func(ctx *gin.Context) {
		ctx.Set("claims", claims)
		ctx.Next()
	}
}
