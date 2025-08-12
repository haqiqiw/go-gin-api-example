package config

import (
	"errors"
	"go-api-example/internal/delivery/http/middleware"
	"go-api-example/internal/model"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func NewFiber(env *Env, logger *zap.Logger) *fiber.App {
	fCfg := fiber.Config{
		AppName:      env.AppName,
		ReadTimeout:  time.Duration(env.AppReadTimeout) * time.Second,
		WriteTimeout: time.Duration(env.AppWriteTimeout) * time.Second,
		ErrorHandler: NewErrorHandler(logger),
	}

	return fiber.New(fCfg)
}

func NewErrorHandler(logger *zap.Logger) fiber.ErrorHandler {
	return func(ctx *fiber.Ctx, err error) error {
		logger.Error(err.Error(),
			zap.Any("request_id", middleware.GetRequestID(ctx)),
			zap.Any("path", ctx.Path()),
			zap.Any("method", ctx.Method()),
			zap.Error(err),
		)

		resp := model.ErrorResponse{}

		var customErr *model.CustomError
		if errors.As(err, &customErr) {
			resp.Errors = customErr.Errors
			resp.Meta.HTTPStatus = customErr.HTTPStatus

			return ctx.Status(customErr.HTTPStatus).JSON(resp)
		}

		if e, ok := err.(*fiber.Error); ok {
			resp.Errors = []model.ErrorItem{
				{
					Code:    e.Code,
					Message: e.Message,
				},
			}
			resp.Meta.HTTPStatus = e.Code

			return ctx.Status(e.Code).JSON(resp)
		}

		resp.Errors = []model.ErrorItem{
			{
				Code:    model.ErrInternalServerError,
				Message: model.MessageFor(model.ErrInternalServerError),
			},
		}
		resp.Meta.HTTPStatus = fiber.StatusInternalServerError
		return ctx.Status(fiber.StatusInternalServerError).JSON(resp)
	}
}
