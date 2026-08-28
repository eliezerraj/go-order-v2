package worker

import (
	"os"
	"sync"

	"github.com/eliezerraj/go-core/v3/logger"

	"github.com/go-order-v2/cmd/worker/event_kafka"
	"github.com/go-order-v2/application/config"
)

var (
	startKafkaConsumer = os.Getenv("START_KAFKA_CONSUMER") == "true"
)

func Run(kafkaConsumer config.KafkaConsumer) {
	logger.InfoOutCtx("starting worker process SUCCESSFULLY")

	wg := sync.WaitGroup{}
	wg.Add(1)
	
	go func() {

		event_kafka.Run(kafkaConsumer)
		defer wg.Done()
	}()

	wg.Wait()
}