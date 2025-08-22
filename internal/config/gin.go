package config

import (
	"errors"
	"fmt"
	"go-api-example/internal/model"
	"net/http"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func NewGin(logger *zap.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(requestid.New())
	engine.Use(requestLoggerHandler(logger))
	engine.Use(recoverHandler(logger))
	engine.Use(errorHandler(logger))

	return engine
}

func recoverHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				var err error
				switch t := r.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = fmt.Errorf("%+v", t)
				}

				logger.Error(fmt.Sprintf("panic recovered: %+v", err),
					zap.Any("request_id", requestid.Get(ctx)),
					zap.Any("path", ctx.Request.RequestURI),
					zap.Any("method", ctx.Request.Method),
					zap.Error(err),
				)

				resp := model.ErrorResponse{}
				resp.Errors = model.ErrInternalServerError.Errors
				resp.Meta.HTTPStatus = model.ErrInternalServerError.HTTPStatus

				ctx.JSON(http.StatusInternalServerError, resp)
			}
		}()

		ctx.Next()
	}
}

func errorHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if len(ctx.Errors) > 0 {
			err := ctx.Errors.Last().Err

			logger.Error(err.Error(),
				zap.Any("request_id", requestid.Get(ctx)),
				zap.Any("path", ctx.Request.RequestURI),
				zap.Any("method", ctx.Request.Method),
				zap.Error(err),
			)

			resp := model.ErrorResponse{}

			var customErr *model.CustomError
			if errors.As(err, &customErr) {
				resp.Errors = customErr.Errors
				resp.Meta.HTTPStatus = customErr.HTTPStatus

				ctx.JSON(customErr.HTTPStatus, resp)
			} else {
				resp.Errors = model.ErrInternalServerError.Errors
				resp.Meta.HTTPStatus = model.ErrInternalServerError.HTTPStatus

				ctx.JSON(http.StatusInternalServerError, resp)
			}
		}
	}
}

func requestLoggerHandler(logger *zap.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()

		logger.Info("request finished",
			zap.Any("request_id", requestid.Get(ctx)),
			zap.Any("path", ctx.Request.RequestURI),
			zap.Any("method", ctx.Request.Method),
			zap.Any("status", ctx.Writer.Status()),
			zap.Duration("duration", time.Since(start)),
		)
	}
}
