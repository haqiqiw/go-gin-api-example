package middleware

import (
	"errors"
	"fmt"
	"go-api-example/internal/auth"
	"go-api-example/internal/model"
	"go-api-example/internal/storage"
	"strings"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewAuthMiddleware(logger *zap.Logger, redisClient storage.RedisClient, jwtToken auth.JWTToken) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			logger.Warn("missing or invalid auth header",
				zap.Any("request_id", requestid.Get(ctx)),
				zap.Any("path", ctx.Request.RequestURI),
				zap.Any("method", ctx.Request.Method),
			)
			ctx.Error(model.ErrMissingOrInvalidAuthHeader)
			ctx.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := jwtToken.Parse(token)
		if err != nil {
			logger.Warn(err.Error(),
				zap.Any("request_id", requestid.Get(ctx)),
				zap.Any("path", ctx.Request.RequestURI),
				zap.Any("method", ctx.Request.Method),
			)
			ctx.Error(model.ErrInvalidAuthToken)
			ctx.Abort()
			return
		}

		revokeKey := fmt.Sprintf("%s:%s", auth.PrefixRevokeKey, claims.ID)
		exists, err := redisClient.Exists(ctx.Request.Context(), revokeKey).Result()
		if err != nil {
			logger.Warn(err.Error(),
				zap.Any("request_id", requestid.Get(ctx)),
				zap.Any("path", ctx.Request.RequestURI),
				zap.Any("method", ctx.Request.Method),
			)
			ctx.Error(model.ErrInvalidAuthToken)
			ctx.Abort()
			return
		}

		if exists == 1 {
			ctx.Error(model.ErrTokenRevoked)
			ctx.Abort()
			return
		}

		ctx.Set("claims", claims)
		ctx.Next()
	}
}

func GetJWTClaims(c *gin.Context) (*auth.JWTClaims, error) {
	claims, exist := c.Get("claims")
	if !exist {
		return nil, errors.New("invalid jwt claims")
	}

	jwtClaims, ok := claims.(*auth.JWTClaims)
	if !ok {
		return nil, errors.New("invalid jwt claims type")
	}

	return jwtClaims, nil
}
