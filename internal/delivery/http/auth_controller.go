package http

import (
	"go-api-example/internal/delivery/http/middleware"
	"go-api-example/internal/model"
	"go-api-example/internal/usecase"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
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

func (c *AuthController) Login(ctx *fiber.Ctx) error {
	request := new(model.LoginRequest)

	err := ctx.BodyParser(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to parse request body", err)
		return fiber.ErrBadRequest
	}

	err = c.Validate.Struct(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to validate request body", err)
		return fiber.ErrBadRequest
	}

	res, err := c.AuthUsecase.Login(ctx.UserContext(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to login", err)
		return err
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(model.NewSuccessResponse(res, fiber.StatusOK))
}

func (c *AuthController) Logout(ctx *fiber.Ctx) error {
	claims, err := middleware.GetJWTClaims(ctx)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get jwt claims", err)
		return fiber.ErrUnauthorized
	}

	request := new(model.LogoutRequest)

	err = ctx.BodyParser(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to parse request body", err)
		return fiber.ErrBadRequest
	}

	err = c.Validate.Struct(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to validate request body", err)
		return fiber.ErrBadRequest
	}

	request.Claims = claims
	err = c.AuthUsecase.Logout(ctx.UserContext(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to logout", err)
		return err
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(model.NewSuccessMessageResponse("Logged out", fiber.StatusOK))
}

func (c *AuthController) RefreshToken(ctx *fiber.Ctx) error {
	request := new(model.RefreshRequest)

	err := ctx.BodyParser(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to parse request body", err)
		return fiber.ErrBadRequest
	}

	err = c.Validate.Struct(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to validate request body", err)
		return fiber.ErrBadRequest
	}

	res, err := c.AuthUsecase.Refresh(ctx.UserContext(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to refresh token", err)
		return err
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(model.NewSuccessResponse(res, fiber.StatusOK))
}
