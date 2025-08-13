package main

import (
	"context"
	"fmt"
	"go-api-example/internal/config"
	"go-api-example/internal/db"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger, err := config.NewLogger()
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	env, err := config.NewEnv()
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to initialize env: %v", err))
	}

	database, err := config.NewDatabase(env)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to initialize database: %v", err))
	}

	producer, err := config.NewKafkaProducer(env, logger)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to initialize producer: %v", err))
	}

	tx := db.NewTransactioner(database)
	validate := config.NewValidator()
	app := config.NewFiber(env, logger)

	config.NewApi(&config.ApiConfig{
		DB:       database,
		TX:       tx,
		App:      app,
		Log:      logger,
		Validate: validate,
		Config:   env,
		Producer: producer,
	})

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		logger.Info("shutting down server...")

		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		err := app.ShutdownWithContext(ctx)
		if err != nil {
			logger.Fatal(fmt.Sprintf("failed to shutdown server: %v", err))
		}
	}()

	logger.Info(fmt.Sprintf("starting server at port %d", env.AppPort))

	err = app.Listen(fmt.Sprintf(":%d", env.AppPort))
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to starting server: %v", err))
	}

	logger.Info("server exited properly")
}
