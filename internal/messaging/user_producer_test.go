package messaging_test

import (
	"errors"
	"go-api-example/internal/messaging"
	"go-api-example/internal/mocks"
	"go-api-example/internal/model"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type UserProducerSuite struct {
	suite.Suite
	logger   *zap.Logger
	kafka    *mocks.KafkaProducer
	producer messaging.Producer[*model.UserEvent]
	topic    string
}

func (s *UserProducerSuite) SetupTest() {
	s.logger, _ = zap.NewDevelopment()
	s.kafka = mocks.NewKafkaProducer(s.T())
	s.topic = "user-registered"
	s.producer = messaging.NewUserProducer(s.logger, s.kafka, s.topic)
}

func (s *UserProducerSuite) TearDownTest() {
	s.kafka = mocks.NewKafkaProducer(s.T())
}

func (s *UserProducerSuite) TestUserProducer_GetTopic() {
	t := s.producer.GetTopic()

	s.Equal("user-registered", *t)
}

func (s *UserProducerSuite) TestUserProducer_Send() {
	tests := []struct {
		name       string
		mockFunc   func(k *mocks.KafkaProducer)
		param      *model.UserEvent
		wantErrMsg string
	}{
		{
			name: "error on produce",
			mockFunc: func(k *mocks.KafkaProducer) {
				k.On("Produce", mock.Anything, mock.Anything).
					Return(errors.New("something error"))
			},
			param: &model.UserEvent{
				ID:        1,
				Username:  "johndoe",
				CreatedAt: time.Now().Format(time.RFC3339),
				UpdatedAt: time.Now().Format(time.RFC3339),
			},
			wantErrMsg: "failed to produce message for user-registered: something error",
		},
		{
			name: "success",
			mockFunc: func(k *mocks.KafkaProducer) {
				k.On("Produce", mock.Anything, mock.Anything).Return(nil)
			},
			param: &model.UserEvent{
				ID:        1,
				Username:  "johndoe",
				CreatedAt: time.Now().Format(time.RFC3339),
				UpdatedAt: time.Now().Format(time.RFC3339),
			},
			wantErrMsg: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.kafka = mocks.NewKafkaProducer(s.T())
			s.producer = messaging.NewUserProducer(s.logger, s.kafka, s.topic)
			tt.mockFunc(s.kafka)

			err := s.producer.Send(tt.param)

			if tt.wantErrMsg == "" {
				s.Nil(err)
			} else {
				s.Equal(tt.wantErrMsg, err.Error())
			}
		})
	}
}

func TestUserProducerSuite(t *testing.T) {
	suite.Run(t, new(UserProducerSuite))
}
