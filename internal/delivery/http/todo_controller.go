package http

import (
	"go-api-example/internal/delivery/http/middleware"
	"go-api-example/internal/entity"
	"go-api-example/internal/model"
	"go-api-example/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

func (c *TodoController) Create(ctx *gin.Context) {
	claims, err := middleware.GetJWTClaims(ctx)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get jwt claims", err)
		ctx.Error(model.ErrUnauthorized)
		return
	}

	userID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert user id", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	request := new(model.CreateTodoRequest)
	err = ctx.ShouldBindJSON(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to parse request body", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	request.UserID = userID
	err = c.Validate.Struct(request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to validate request body", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	res, err := c.TodoUsecase.Create(ctx.Request.Context(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to create todo", err)
		ctx.Error(err)
		return
	}

	ctx.JSON(
		http.StatusCreated,
		model.NewSuccessResponse(res, http.StatusCreated),
	)
}

func (c *TodoController) Search(ctx *gin.Context) {
	claims, err := middleware.GetJWTClaims(ctx)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get jwt claims", err)
		ctx.Error(model.ErrUnauthorized)
		return
	}

	userID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert user id", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	var status *entity.TodoStatus

	statusQuery := ctx.Query("status")
	if statusQuery != "" {
		ts, err := entity.ParseTodoStatus(statusQuery)
		if err != nil {
			LogWarn(ctx, c.Log, "failed to convert todo status", err)
			ctx.Error(model.ErrBadRequest)
			return
		}
		status = &ts
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	if err != nil || limit < 0 {
		offset = 0
	}

	request := &model.SearchTodoRequest{
		UserID: userID,
		Status: status,
		Limit:  limit,
		Offset: offset,
	}
	res, total, err := c.TodoUsecase.List(ctx.Request.Context(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get todos", err)
		ctx.Error(err)
		return
	}

	meta := model.MetaWithPage{
		Limit:      limit,
		Offset:     offset,
		Total:      total,
		HTTPStatus: http.StatusOK,
	}
	ctx.JSON(
		http.StatusOK,
		model.NewSuccessListResponse(res, meta),
	)
}

func (c *TodoController) Get(ctx *gin.Context) {
	claims, err := middleware.GetJWTClaims(ctx)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get jwt claims", err)
		ctx.Error(model.ErrUnauthorized)
		return
	}

	userID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert user id", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert id", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	res, err := c.TodoUsecase.FindByID(ctx.Request.Context(), &model.GetTodoRequest{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get todo", err)
		ctx.Error(err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		model.NewSuccessResponse(res, http.StatusOK),
	)
}

func (c *TodoController) Update(ctx *gin.Context) {
	claims, err := middleware.GetJWTClaims(ctx)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get jwt claims", err)
		ctx.Error(model.ErrUnauthorized)
		return
	}

	userID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert user id", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert id", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	request := new(model.UpdateTodoRequest)
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

	request.ID = id
	request.UserID = userID
	request.IntStatus, err = entity.ParseTodoStatus(request.Status)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert todo status", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	err = c.TodoUsecase.UpdateByID(ctx.Request.Context(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to update todo", err)
		ctx.Error(err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		model.NewSuccessMessageResponse("Todo updated", http.StatusOK),
	)
}
