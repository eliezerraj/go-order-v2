package event_kafka

import (
	"context"
	"go.uber.org/zap"

	"github.com/go-order-v2/application/config"
	"github.com/eliezerraj/go-core/v3/logger"

	gocore_kafka "github.com/eliezerraj/go-core/v3/event/kafka"
	"github.com/eliezerraj/go-core/v3/event/kafka/consumer"
)


func Run(ctx context.Context, kafkaConsumer config.KafkaConsumer) {
	logger.InfoOutCtx("starting event Kafka worker process SUCCESSFULLY")
	
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

	messageProcessor := NewMessageProcessor(consumerWorker)
	messageProcessor.Start(ctx)

}