package messaging_test

import (
	"context"
	"encoding/json"
	"errors"
	"go-api-example/internal/delivery/messaging"
	"go-api-example/internal/mocks"
	"go-api-example/internal/model"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestUserHandler_Consume(t *testing.T) {
	ctx := context.Background()
	logger, _ := zap.NewDevelopment()
	topic := "user-registered"

	validMsg := func() *kafka.Message {
		event := &model.UserEvent{
			ID:        1,
			Username:  "johndoe",
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
		}
		data, _ := json.Marshal(event)
		msg := &kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Value: data,
			Key:   []byte(event.GetID()),
		}

		return msg
	}

	tests := []struct {
		name       string
		message    *kafka.Message
		mockFunc   func(t *mocks.TodoUsecase)
		wantErrMsg string
	}{
		{
			name: "error on unrmarshal",
			message: func() *kafka.Message {
				data, _ := json.Marshal("dummy")
				msg := &kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     &topic,
						Partition: kafka.PartitionAny,
					},
					Value: data,
					Key:   []byte("1"),
				}

				return msg
			}(),
			mockFunc:   func(t *mocks.TodoUsecase) {},
			wantErrMsg: "failed to unmarshal event for user-registered",
		},
		{
			name:    "error on create",
			message: validMsg(),
			mockFunc: func(t *mocks.TodoUsecase) {
				matcher := mock.MatchedBy(func(r *model.CreateTodoRequest) bool {
					return r.UserID == uint64(1) && r.Title == "Welcome to the Todo App" &&
						*r.Description == "Add your first real todo!"
				})
				t.On("Create", mock.Anything, matcher).Return(nil, errors.New("something error"))
			},
			wantErrMsg: "something error",
		},
		{
			name:    "success",
			message: validMsg(),
			mockFunc: func(t *mocks.TodoUsecase) {
				matcher := mock.MatchedBy(func(r *model.CreateTodoRequest) bool {
					return r.UserID == uint64(1) && r.Title == "Welcome to the Todo App" &&
						*r.Description == "Add your first real todo!"
				})
				t.On("Create", mock.Anything, matcher).Return(&model.TodoResponse{}, nil)
			},
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			todoUsecase := mocks.NewTodoUsecase(t)
			handler := messaging.NewUserHandler(logger, todoUsecase)
			tt.mockFunc(todoUsecase)

			err := handler.Consume(ctx, tt.message)

			if tt.wantErrMsg != "" {
				assert.Contains(t, err.Error(), tt.wantErrMsg)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
