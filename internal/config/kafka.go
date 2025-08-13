package config

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

func NewKafkaProducer(env *Env, logger *zap.Logger) (*kafka.Producer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers": env.KafkaBrokerHost,
	}

	producer, err := kafka.NewProducer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	return producer, nil
}

func NewKafkaConsumer(env *Env, logger *zap.Logger) (*kafka.Consumer, error) {
	cfg := &kafka.ConfigMap{
		"bootstrap.servers": env.KafkaBrokerHost,
		"group.id":          env.KafkaConsumerGroup,
		"auto.offset.reset": env.KafkaAutoOffsetReset,
	}

	consumer, err := kafka.NewConsumer(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	return consumer, nil
}
