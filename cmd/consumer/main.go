package main

import (
	"context"
	"fmt"
	"go-api-example/internal/config"
	"go-api-example/internal/delivery/messaging"
	"go-api-example/internal/repository"
	"go-api-example/internal/usecase"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"
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

	todoRepository := repository.NewTodoRepository(database)
	todoUsecase := usecase.NewTodoUsecase(logger, todoRepository)

	kafkaConsumer, err := config.NewKafkaConsumer(env, logger)
	if err != nil {
		logger.Fatal(fmt.Sprintf("failed to initialize user consumer: %v", err))
	}

	userHandler := messaging.NewUserHandler(logger, todoUsecase)

	consumerCfg := &messaging.ConsumerConfig{
		Topic:              env.KafkaTopicUserRegistered,
		MaxRetries:         3,
		BackoffDuration:    1 * time.Second,
		MaxExecuteDuration: 10 * time.Second,
	}
	userConsumer := messaging.NewConsumer(logger, kafkaConsumer, consumerCfg, userHandler.Consume)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := userConsumer.Consume(ctx)
		if err != nil {
			errCh <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-quit:
		logger.Info("stop signal received, shutting down...", zap.String("signal", s.String()))
	case e := <-errCh:
		logger.Error("consumer error, shutting down...", zap.Error(e))
	}

	cancel()
	wg.Wait()

	logger.Info("consumer exited properly")
}
