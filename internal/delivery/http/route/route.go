package route

import (
	internalHttp "go-api-example/internal/delivery/http"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RouteConfig struct {
	App            *gin.Engine
	AuthMiddlware  gin.HandlerFunc
	AuthController *internalHttp.AuthController
	UserController *internalHttp.UserController
	TodoController *internalHttp.TodoController
}

func (c *RouteConfig) Setup() {
	c.App.GET("/healthz", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "OK")
	})

	c.SetupPublicRoute()
	c.SetupAuthRoute()
}

func (c *RouteConfig) SetupPublicRoute() {
	c.App.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "OK")
	})

	c.App.POST("/api/login", c.AuthController.Login)
	c.App.POST("/api/refresh-token", c.AuthController.RefreshToken)

	c.App.POST("/api/users", c.UserController.Register)
}

func (c *RouteConfig) SetupAuthRoute() {
	c.App.POST("/api/logout", c.AuthMiddlware, c.AuthController.Logout)

	c.App.GET("/api/users", c.AuthMiddlware, c.UserController.Search)
	c.App.GET("/api/users/me", c.AuthMiddlware, c.UserController.Me)
	c.App.PATCH("/api/users/me", c.AuthMiddlware, c.UserController.Update)

	c.App.POST("/api/todos", c.AuthMiddlware, c.TodoController.Create)
	c.App.GET("/api/todos", c.AuthMiddlware, c.TodoController.Search)
	c.App.GET("/api/todos/:id", c.AuthMiddlware, c.TodoController.Get)
	c.App.PATCH("/api/todos/:id", c.AuthMiddlware, c.TodoController.Update)
}
