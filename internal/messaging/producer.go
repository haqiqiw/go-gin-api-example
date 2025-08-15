package messaging

import (
	"encoding/json"
	"fmt"
	"go-api-example/internal/model"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

//go:generate mockery --name=KafkaProducer --structname KafkaProducer --outpkg=mocks --output=./../mocks
type KafkaProducer interface {
	Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
}

type Producer[T model.Event] interface {
	GetTopic() *string
	Send(event T) error
}

type producer[T model.Event] struct {
	Producer KafkaProducer
	Topic    string
	Log      *zap.Logger
}

func (p *producer[T]) GetTopic() *string {
	return &p.Topic
}

func (p *producer[T]) Send(event T) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event for %s: %w", *p.GetTopic(), err)
	}

	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     p.GetTopic(),
			Partition: kafka.PartitionAny,
		},
		Value: data,
		Key:   []byte(event.GetID()),
	}

	err = p.Producer.Produce(msg, nil)
	if err != nil {
		return fmt.Errorf("failed to produce message for %s: %w", *p.GetTopic(), err)
	}

	return nil
}
