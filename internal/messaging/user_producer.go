package messaging

import (
	"go-api-example/internal/model"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

type UserProducer struct {
	Producer[*model.UserEvent]
}

func NewUserProducer(logger *zap.Logger, kProducer *kafka.Producer, topic string) *UserProducer {
	return &UserProducer{
		Producer: &producer[*model.UserEvent]{
			Producer: kProducer,
			Topic:    topic,
			Log:      logger,
		},
	}
}
