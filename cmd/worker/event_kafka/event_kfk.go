package event_kafka

import (

	"github.com/go-order-v2/application/config"
	"github.com/eliezerraj/go-core/v3/logger"

	gocore_kafka "github.com/eliezerraj/go-core/v3/event/kafka"
)

func Run(kafkaConsumer config.KafkaConsumer) {
	logger.InfoOutCtx("starting event Kafka worker process")
	
	dialerConfig := gocore_kafka.DialerConfig{
		Username:   kafkaConsumer.Username,
		Password:   kafkaConsumer.Password,
		Protocol:   kafkaConsumer.Protocol,
		Mechanisms: kafkaConsumer.Mechanism,
		Brokers:    kafkaConsumer.BrokerList,
	}

	kafkaDialer := gocore_kafka.NewKafkaDialer(dialerConfig)
	consumerConfig := kafkaDialer.ConsumerConfig(kafkaConsumer.GroupID, kafkaConsumer.Name)
	_ = consumerConfig
}