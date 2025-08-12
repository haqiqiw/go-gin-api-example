package http

import (
	"go-api-example/internal/delivery/http/middleware"
	"go-api-example/internal/model"
	"go-api-example/internal/usecase"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type UserController struct {
	Log         *zap.Logger
	Validate    *validator.Validate
	UserUsecase usecase.UserUsecase
}

func NewUserController(log *zap.Logger, validate *validator.Validate, userUsecase usecase.UserUsecase) *UserController {
	return &UserController{
		Log:         log,
		Validate:    validate,
		UserUsecase: userUsecase,
	}
}

func (c *UserController) Register(ctx *fiber.Ctx) error {
	request := new(model.CreateUserRequest)
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

	res, err := c.UserUsecase.Create(ctx.UserContext(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to register user", err)
		return err
	}

	return ctx.
		Status(fiber.StatusCreated).
		JSON(model.NewSuccessResponse(res, fiber.StatusCreated))
}

func (c *UserController) Search(ctx *fiber.Ctx) error {
	var id *uint64
	var username *string

	idQuery := uint64(ctx.QueryInt("id"))
	if idQuery > 0 {
		id = &idQuery
	}

	usernameQuery := ctx.Query("username")
	if usernameQuery != "" {
		username = &usernameQuery
	}

	limit := ctx.QueryInt("limit", 10)
	if limit <= 0 {
		limit = 10
	}
	offset := ctx.QueryInt("offset", 0)

	request := &model.SearchUserRequest{
		ID:       id,
		Username: username,
		Limit:    limit,
		Offset:   offset,
	}
	res, total, err := c.UserUsecase.List(ctx.UserContext(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get users", err)
		return err
	}

	meta := model.MetaWithPage{
		Limit:      limit,
		Offset:     offset,
		Total:      total,
		HTTPStatus: fiber.StatusOK,
	}
	return ctx.
		Status(fiber.StatusOK).
		JSON(model.NewSuccessListResponse(res, meta))
}

func (c *UserController) Me(ctx *fiber.Ctx) error {
	claims, err := middleware.GetJWTClaims(ctx)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get jwt claims", err)
		return fiber.ErrUnauthorized
	}

	id, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert id", err)
		return fiber.ErrBadRequest
	}

	res, err := c.UserUsecase.FindByID(ctx.UserContext(), &model.GetUserRequest{ID: id})
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get user", err)
		return err
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(model.NewSuccessResponse(res, fiber.StatusOK))
}

func (c *UserController) Update(ctx *fiber.Ctx) error {
	claims, err := middleware.GetJWTClaims(ctx)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get jwt claims", err)
		return fiber.ErrUnauthorized
	}

	id, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert id", err)
		return fiber.ErrBadRequest
	}

	request := new(model.UpdateUserRequest)
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

	request.ID = id
	err = c.UserUsecase.UpdateByID(ctx.UserContext(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to update user", err)
		return err
	}

	return ctx.
		Status(fiber.StatusCreated).
		JSON(model.NewSuccessMessageResponse("User updated", fiber.StatusCreated))
}
