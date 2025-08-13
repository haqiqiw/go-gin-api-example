package config

import (
	"database/sql"
	"fmt"
	"go-api-example/internal/auth"
	"go-api-example/internal/db"
	"go-api-example/internal/delivery/http"
	"go-api-example/internal/delivery/http/middleware"
	"go-api-example/internal/delivery/http/route"
	"go-api-example/internal/messaging"
	"go-api-example/internal/repository"
	"go-api-example/internal/usecase"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type ApiConfig struct {
	DB       *sql.DB
	TX       db.Transactioner
	App      *fiber.App
	Log      *zap.Logger
	Validate *validator.Validate
	Config   *Env
	Producer *kafka.Producer
}

func NewApi(cfg *ApiConfig) {
	redisClient := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", cfg.Config.RedisHost, cfg.Config.RedistPort),
		DB:   cfg.Config.RedistDB,
	})

	jwtToken := auth.NewJWTToken(cfg.Config.JWTSecretKey)
	refreshToken := auth.NewRefreshToken()

	authMiddleware := middleware.NewAuthMiddleware(cfg.Log, redisClient, jwtToken)

	userProducer := messaging.NewUserProducer(cfg.Log, cfg.Producer, cfg.Config.KafkaTopicUserRegistered)

	userRepository := repository.NewUserRepository(cfg.DB)
	todoRepository := repository.NewTodoRepository(cfg.DB)

	authUsecase := usecase.NewAuthUsecase(cfg.Log, redisClient, jwtToken, refreshToken, userRepository)
	userUsecase := usecase.NewUserUsecase(cfg.Log, cfg.TX, userProducer, userRepository)
	todoUsecase := usecase.NewTodoUsecase(cfg.Log, todoRepository)

	authController := http.NewAuthController(cfg.Log, cfg.Validate, authUsecase, userUsecase)
	userController := http.NewUserController(cfg.Log, cfg.Validate, userUsecase)
	todoController := http.NewTodoController(cfg.Log, cfg.Validate, todoUsecase)

	routeCfg := route.RouteConfig{
		App:            cfg.App,
		AuthMiddlware:  authMiddleware,
		AuthController: authController,
		UserController: userController,
		TodoController: todoController,
	}
	routeCfg.Setup()
}
