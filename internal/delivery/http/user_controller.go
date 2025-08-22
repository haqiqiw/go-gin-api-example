package http

import (
	"go-api-example/internal/delivery/http/middleware"
	"go-api-example/internal/model"
	"go-api-example/internal/usecase"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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

func (c *UserController) Register(ctx *gin.Context) {
	request := new(model.CreateUserRequest)
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

	res, err := c.UserUsecase.Create(ctx.Request.Context(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to register user", err)
		ctx.Error(err)
		return
	}

	ctx.JSON(
		http.StatusCreated,
		model.NewSuccessResponse(res, http.StatusCreated),
	)
}

func (c *UserController) Search(ctx *gin.Context) {
	var id *uint64
	var username *string

	idQuery, err := strconv.ParseUint(ctx.Query("id"), 10, 64)
	if err != nil {
		idQuery = 0
	}
	if idQuery > 0 {
		id = &idQuery
	}

	usernameQuery := ctx.Query("username")
	if usernameQuery != "" {
		username = &usernameQuery
	}

	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	if err != nil || limit < 0 {
		offset = 0
	}

	request := &model.SearchUserRequest{
		ID:       id,
		Username: username,
		Limit:    limit,
		Offset:   offset,
	}
	res, total, err := c.UserUsecase.List(ctx.Request.Context(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get users", err)
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

func (c *UserController) Me(ctx *gin.Context) {
	claims, err := middleware.GetJWTClaims(ctx)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get jwt claims", err)
		ctx.Error(model.ErrUnauthorized)
		return
	}

	userID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to convert id", err)
		ctx.Error(model.ErrBadRequest)
		return
	}

	res, err := c.UserUsecase.FindByID(ctx.Request.Context(), &model.GetUserRequest{
		ID: userID,
	})
	if err != nil {
		LogWarn(ctx, c.Log, "failed to get user", err)
		ctx.Error(err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		model.NewSuccessResponse(res, http.StatusOK),
	)
}

func (c *UserController) Update(ctx *gin.Context) {
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

	request := new(model.UpdateUserRequest)
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

	request.ID = userID
	err = c.UserUsecase.UpdateByID(ctx.Request.Context(), request)
	if err != nil {
		LogWarn(ctx, c.Log, "failed to update user", err)
		ctx.Error(err)
		return
	}

	ctx.JSON(
		http.StatusOK,
		model.NewSuccessMessageResponse("User updated", http.StatusOK),
	)
}
