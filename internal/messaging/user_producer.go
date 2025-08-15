package messaging

import (
	"go-api-example/internal/model"

	"go.uber.org/zap"
)

type UserProducer struct {
	Producer[*model.UserEvent]
}

func NewUserProducer(logger *zap.Logger, kProducer KafkaProducer, topic string) *UserProducer {
	return &UserProducer{
		Producer: &producer[*model.UserEvent]{
			Producer: kProducer,
			Topic:    topic,
			Log:      logger,
		},
	}
}
