package worker

import (
	"context"
	"os"

	"github.com/eliezerraj/go-core/v3/logger"

	"github.com/go-order-v2/cmd/worker/event_kafka"
	"github.com/go-order-v2/application/config"
)

var (
	startKafkaConsumer = os.Getenv("START_KAFKA_CONSUMER") == "true"
)

func Run(ctx context.Context, kafkaConsumer config.KafkaConsumer) {
	logger.InfoOutCtx("KAFKA starting worker process SUCCESSFULLY")

	event_kafka.Run(ctx, kafkaConsumer)

	logger.InfoOutCtx("KAFKA worker process finished cleanly")

}