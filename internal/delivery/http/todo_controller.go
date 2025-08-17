package http

import (
	"go-api-example/internal/delivery/http/middleware"
	"go-api-example/internal/entity"
	"go-api-example/internal/model"
	"go-api-example/internal/usecase"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type TodoController struct {
	Log         *zap.Logger
	Validate    *validator.Validate
	TodoUsecase usecase.TodoUsecase
}

func NewTodoController(log *zap.Logger, validate *validator.Validate, todoUsecase usecase.TodoUsecase) *TodoController {
	return &TodoController{
		Log:         log,
		Validate:    validate,
		TodoUsecase: todoUsecase,
	}
}

func (c *TodoController) Create(ctx *fiber.Ctx) error {
	claims, err := middleware.GetJWTClaims(ctx)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get jwt claims", err)
		return fiber.ErrUnauthorized
	}

	userID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert user id", err)
		return fiber.ErrBadRequest
	}

	request := new(model.CreateTodoRequest)
	err = ctx.BodyParser(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to parse request body", err)
		return fiber.ErrBadRequest
	}

	request.UserID = userID
	err = c.Validate.Struct(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to validate request body", err)
		return fiber.ErrBadRequest
	}

	res, err := c.TodoUsecase.Create(ctx.UserContext(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to create todo", err)
		return err
	}

	return ctx.
		Status(fiber.StatusCreated).
		JSON(model.NewSuccessResponse(res, fiber.StatusCreated))
}

func (c *TodoController) Search(ctx *fiber.Ctx) error {
	claims, err := middleware.GetJWTClaims(ctx)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get jwt claims", err)
		return fiber.ErrUnauthorized
	}

	userID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert user id", err)
		return fiber.ErrBadRequest
	}

	var status *entity.TodoStatus

	statusQuery := ctx.Query("status")
	if statusQuery != "" {
		ts, err := entity.ParseTodoStatus(statusQuery)
		if err != nil {
			LogWarn(ctx, c.Log, "failed to convert todo status", err)
			return fiber.ErrBadRequest
		}
		status = &ts
	}

	limit := ctx.QueryInt("limit", 10)
	if limit <= 0 {
		limit = 10
	}
	offset := ctx.QueryInt("offset", 0)

	request := &model.SearchTodoRequest{
		UserID: userID,
		Status: status,
		Limit:  limit,
		Offset: offset,
	}
	res, total, err := c.TodoUsecase.List(ctx.UserContext(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get todos", err)
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

func (c *TodoController) Get(ctx *fiber.Ctx) error {
	claims, err := middleware.GetJWTClaims(ctx)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get jwt claims", err)
		return fiber.ErrUnauthorized
	}

	userID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert user id", err)
		return fiber.ErrBadRequest
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert id", err)
		return fiber.ErrBadRequest
	}

	res, err := c.TodoUsecase.FindByID(ctx.UserContext(), &model.GetTodoRequest{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get todo", err)
		return err
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(model.NewSuccessResponse(res, fiber.StatusOK))
}

func (c *TodoController) Update(ctx *fiber.Ctx) error {
	claims, err := middleware.GetJWTClaims(ctx)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get jwt claims", err)
		return fiber.ErrUnauthorized
	}

	userID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert user id", err)
		return fiber.ErrBadRequest
	}

	id, err := strconv.ParseUint(ctx.Params("id"), 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert id", err)
		return fiber.ErrBadRequest
	}

	request := new(model.UpdateTodoRequest)
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
	request.UserID = userID
	request.IntStatus, err = entity.ParseTodoStatus(request.Status)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert todo status", err)
		return fiber.ErrBadRequest
	}

	err = c.TodoUsecase.UpdateByID(ctx.UserContext(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to update todo", err)
		return err
	}

	return ctx.
		Status(fiber.StatusOK).
		JSON(model.NewSuccessMessageResponse("Todo updated", fiber.StatusOK))
}
