package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"go.uber.org/zap"
)

type Handler func(ctx context.Context, message *kafka.Message) error

type ConsumerConfig struct {
	Topic              string
	MaxRetries         int
	BackoffDuration    time.Duration
	MaxExecuteDuration time.Duration
}

type Consumer interface {
	Consume(tx context.Context) error
}

type consumer struct {
	Logger   *zap.Logger
	Consumer *kafka.Consumer
	Config   *ConsumerConfig
	Handler  Handler
}

func NewConsumer(logger *zap.Logger, kconsumer *kafka.Consumer, config *ConsumerConfig, handler Handler) Consumer {
	cfg := &ConsumerConfig{
		Topic:              "",
		MaxRetries:         5,
		BackoffDuration:    3 * time.Second,
		MaxExecuteDuration: 60 * time.Second,
	}

	if config != nil {
		if config.Topic != "" {
			cfg.Topic = config.Topic
		}
		if config.MaxRetries > 0 {
			cfg.MaxRetries = config.MaxRetries
		}
		if config.BackoffDuration > 0 {
			cfg.BackoffDuration = config.BackoffDuration
		}
		if config.MaxExecuteDuration > 0 {
			cfg.MaxExecuteDuration = config.MaxExecuteDuration
		}
	}

	return &consumer{
		Logger:   logger,
		Consumer: kconsumer,
		Config:   cfg,
		Handler:  handler,
	}
}

func (c *consumer) Consume(ctx context.Context) error {
	c.Logger.Info(
		"starting consumer",
		zap.String("topic", c.Config.Topic),
	)

	err := c.Consumer.Subscribe(c.Config.Topic, nil)
	if err != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", c.Config.Topic, err)
	}

	defer func() {
		if closeErr := c.Consumer.Close(); closeErr != nil {
			c.Logger.Error("failed to close consumer",
				zap.String("topic", c.Config.Topic),
				zap.Error(closeErr),
			)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			c.Logger.Error(
				"context cancelled, stopping consumer",
				zap.String("topic", c.Config.Topic),
			)
			return ctx.Err()
		default:
			// continue to read message
		}

		message, err := c.Consumer.ReadMessage(time.Second)
		if err != nil {
			if kafkaErr, ok := err.(kafka.Error); ok && kafkaErr.IsTimeout() {
				continue
			}

			c.Logger.Error("failed to read message",
				zap.String("topic", c.Config.Topic),
				zap.Error(err),
			)
			continue
		}

		err = c.executeWithRetry(ctx, message)
		if err != nil {
			c.Logger.Error("failed to execute message",
				zap.String("topic", c.Config.Topic),
				zap.String("key", string(message.Key)),
				zap.Error(err),
			)
		}

		_, err = c.Consumer.CommitMessage(message)
		if err != nil {
			c.Logger.Error("failed to commit message",
				zap.String("topic", c.Config.Topic),
				zap.String("key", string(message.Key)),
				zap.Error(err),
			)
		}
	}
}

func (c *consumer) executeWithRetry(ctx context.Context, message *kafka.Message) error {
	var lastErr error

	for attempt := 0; attempt < c.Config.MaxRetries; attempt++ {
		handlerCtx, cancel := context.WithTimeout(ctx, c.Config.MaxExecuteDuration)
		defer cancel()

		resultCh := make(chan error, 1)

		go func() {
			defer func() {
				if r := recover(); r != nil {
					resultCh <- fmt.Errorf("handler panic: %+v", r)
				}
			}()
			resultCh <- c.Handler(handlerCtx, message)
		}()

		select {
		case err := <-resultCh:
			cancel()
			if err == nil {
				return nil
			}
			lastErr = err
		case <-handlerCtx.Done():
			cancel()
			lastErr = fmt.Errorf("handler execution timeout")
		}

		if attempt < c.Config.MaxRetries {
			backoff := c.Config.BackoffDuration * time.Duration(1<<attempt+1)
			c.Logger.Error(
				fmt.Sprintf("handler error, retrying %d/%d after %+v", attempt+1, c.Config.MaxRetries, backoff),
				zap.String("topic", c.Config.Topic),
				zap.String("key", string(message.Key)),
				zap.Error(lastErr),
			)

			select {
			case <-time.After(backoff):
				// continue to next attemp
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("handler failed after %d attempts, last error: %w", c.Config.MaxRetries, lastErr)
}
