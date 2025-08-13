package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"go-api-example/internal/model"
	"go-api-example/internal/usecase"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

type UserHandler struct {
	Log         *zap.Logger
	TodoUsecase usecase.TodoUsecase
}

func NewUserHandler(log *zap.Logger, TodoUsecase usecase.TodoUsecase) *UserHandler {
	return &UserHandler{
		Log:         log,
		TodoUsecase: TodoUsecase,
	}
}

func (c *UserHandler) Consume(ctx context.Context, message *kafka.Message) error {
	c.Log.Info(
		fmt.Sprintf("processing event for %s with key %s", message.TopicPartition.String(), string(message.Key)),
		zap.Any("event", string(message.Value)),
	)

	event := new(model.UserEvent)
	err := json.Unmarshal(message.Value, &event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal event for %s with key %s: %w", message.TopicPartition.String(), string(message.Key), err)
	}

	todoDescription := "Add your first real todo!"
	_, err = c.TodoUsecase.Create(ctx, &model.CreateTodoRequest{
		UserID:      event.ID,
		Title:       "Welcome to the Todo App",
		Description: &todoDescription,
	})
	if err != nil {
		return fmt.Errorf("failed to create todo: %w", err)
	}

	c.Log.Info(
		fmt.Sprintf("successfuly proceed event for %s with key %s", message.TopicPartition.String(), string(message.Key)),
		zap.Any("event", string(message.Value)),
	)

	return nil
}
