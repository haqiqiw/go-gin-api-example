package http

import (
	"go-api-example/internal/delivery/http/middleware"
	"go-api-example/internal/model"
	"go-api-example/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
)

type AuthController struct {
	Log         *zap.Logger
	Validate    *validator.Validate
	AuthUsecase usecase.AuthUsecase
}

func NewAuthController(log *zap.Logger, validate *validator.Validate,
	authUsecase usecase.AuthUsecase) *AuthController {
	return &AuthController{
		Log:         log,
		Validate:    validate,
		AuthUsecase: authUsecase,
	}
}

func (c *AuthController) Login(ctx *gin.Context) {
	request := new(model.LoginRequest)
	err := ctx.ShouldBindJSON(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to parse request body", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	err = c.Validate.Struct(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to validate request body", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	res, err := c.AuthUsecase.Login(ctx.Request.Context(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to login", err)
		ctx.Error(err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		model.NewSuccessResponse(res, http.StatusOK),
	)
}

func (c *AuthController) Logout(ctx *gin.Context) {
	claims, err := middleware.GetJWTClaims(ctx)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get jwt claims", err)
		ctx.Error(model.ErrUnauthorized)
		return
	}

	request := new(model.LogoutRequest)
	err = ctx.ShouldBindJSON(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to parse request body", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	err = c.Validate.Struct(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to validate request body", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	request.Claims = claims
	err = c.AuthUsecase.Logout(ctx.Request.Context(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to logout", err)
		ctx.Error(err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		model.NewSuccessMessageResponse("Logged out", http.StatusOK),
	)
}

func (c *AuthController) RefreshToken(ctx *gin.Context) {
	request := new(model.RefreshRequest)
	err := ctx.ShouldBindJSON(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to parse request body", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	err = c.Validate.Struct(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to validate request body", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	res, err := c.AuthUsecase.Refresh(ctx.Request.Context(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to refresh token", err)
		ctx.Error(err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		model.NewSuccessResponse(res, http.StatusOK),
	)
}
