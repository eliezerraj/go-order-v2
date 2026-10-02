package worker

import (
	"context"
	"os"

	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/go-order-v2/application/infrastructure/application"
	"github.com/go-order-v2/cmd/worker/event_kafka"
	"github.com/go-order-v2/application/config"
)

var (
	startKafkaConsumer = os.Getenv("START_KAFKA_CONSUMER") == "true"
)

func Run(ctx context.Context, kafkaConsumer config.KafkaConsumer, application *application.Application) {
	logger.Info(ctx, "KAFKA starting worker process SUCCESSFULLY")

	event_kafka.Run(ctx, kafkaConsumer, application)

	logger.Info(ctx, "KAFKA worker process finished cleanly SUCCESSFULLY")
}