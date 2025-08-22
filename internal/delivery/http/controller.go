package http

import (
	"fmt"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func LogWarn(ctx *gin.Context, logger *zap.Logger, msg string, err error) {
	logger.Warn(fmt.Sprintf("%s: %+v", msg, err),
		zap.Any("request_id", requestid.Get(ctx)),
		zap.Any("path", ctx.Request.RequestURI),
		zap.Any("method", ctx.Request.Method),
		zap.Error(err),
	)
}
