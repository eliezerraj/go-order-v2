package event_kafka

import (
	"context"
	"go.uber.org/zap"

	"github.com/go-order-v2/application/config"
	"github.com/go-order-v2/application/infrastructure/application"
	
	"github.com/eliezerraj/go-core/v3/logger"

	gocore_kafka "github.com/eliezerraj/go-core/v3/event/kafka"
	"github.com/eliezerraj/go-core/v3/event/kafka/consumer"
)

func Run(ctx context.Context, kafkaConsumer config.KafkaConsumer, application *application.Application) {
	logger.InfoOutCtx("starting event Kafka worker process SUCCESSFULLY")
	
	// Configure the Kafka dialer with the provided consumer settings
	dialerConfig := gocore_kafka.DialerConfig{
		Username:   kafkaConsumer.Username,
		Password:   kafkaConsumer.Password,
		Protocol:   kafkaConsumer.Protocol,
		Mechanisms: kafkaConsumer.Mechanism,
		Brokers:    kafkaConsumer.BrokerList,
	}

	kafkaDialer := gocore_kafka.NewKafkaDialer(dialerConfig)
	consumerConfig := kafkaDialer.ConsumerConfig(kafkaConsumer.GroupID, kafkaConsumer.Name)

	// Create a new consumer worker
	consumerWorker, err := consumer.NewConsumerWorker(consumerConfig)
	if err != nil {
		logger.ErrorOutCtx("failed to create consumer worker: %v", zap.Error(err))
		return
	}
	// Subscribe to the specified topics
	topics := []string{kafkaConsumer.Topic}

	// Subscribe the consumer worker to the topics
	err = consumerWorker.SubscribeTopics(topics)
	if err != nil {
		logger.ErrorOutCtx("failed to subscribe to topics: %v", zap.Error(err))
		return
	}

	paymentEvent := NewPaymentEvent(application)

	messageProcessor := NewMessageProcessor(consumerWorker, paymentEvent)
	err = messageProcessor.Start(ctx)
	if err != nil {
		logger.ErrorOutCtx("failed to start message processor: %v", zap.Error(err))
		return
	}
}