package route

import (
	"go-api-example/internal/delivery/http"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
)

type RouteConfig struct {
	App            *fiber.App
	AuthMiddlware  fiber.Handler
	AuthController *http.AuthController
	UserController *http.UserController
	TodoController *http.TodoController
}

func (c *RouteConfig) Setup() {
	c.App.Use(requestid.New())
	c.App.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))
	c.App.Use(logger.New())

	c.App.Use(healthcheck.New())
	c.App.Get("/healthz", func(c *fiber.Ctx) error {
		c.Context().SetStatusCode(fiber.StatusOK)
		return c.SendString("OK")
	})
	c.App.Get("/metrics", monitor.New())

	c.SetupPublicRoute()
	c.SetupAuthRoute()
}

func (c *RouteConfig) SetupPublicRoute() {
	c.App.Get("/", func(c *fiber.Ctx) error {
		c.Context().SetStatusCode(fiber.StatusOK)
		return c.SendString("OK")
	})

	c.App.Post("/api/login", c.AuthController.Login)
	c.App.Post("/api/refresh-token", c.AuthController.RefreshToken)

	c.App.Post("/api/users", c.UserController.Register)
}

func (c *RouteConfig) SetupAuthRoute() {
	c.App.Post("/api/logout", c.AuthMiddlware, c.AuthController.Logout)

	c.App.Get("/api/users", c.AuthMiddlware, c.UserController.Search)
	c.App.Get("/api/users/me", c.AuthMiddlware, c.UserController.Me)
	c.App.Patch("/api/users/me", c.AuthMiddlware, c.UserController.Update)

	c.App.Post("/api/todos", c.AuthMiddlware, c.TodoController.Create)
	c.App.Get("/api/todos", c.AuthMiddlware, c.TodoController.Search)
	c.App.Get("/api/todos/:id", c.AuthMiddlware, c.TodoController.Get)
	c.App.Patch("/api/todos/:id", c.AuthMiddlware, c.TodoController.Update)
}
