package http

import (
	"fmt"
	"go-api-example/internal/delivery/http/middleware"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func LogWarn(ctx *fiber.Ctx, logger *zap.Logger, msg string, err error) {
	reqID := middleware.GetRequestID(ctx)
	path := ctx.Path()
	method := ctx.Method()

	logger.Warn(fmt.Sprintf("%s: %+v", msg, err),
		zap.Any("request_id", reqID),
		zap.Any("path", path),
		zap.Any("method", method),
		zap.Error(err),
	)
}
