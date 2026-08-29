package event_kafka

import (
	"context"
	"os"
	"syscall"
	"os/signal"
	"fmt"

	"go.uber.org/zap"

	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/eliezerraj/go-core/v3/event/kafka/consumer"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type Message struct {
	Header 	*map[string]string
	Payload string
	Raw     *kafka.Message
}

type MessageProcessor struct {
	consumerWorker *consumer.ConsumerWorker
}

func NewMessageProcessor(consumerWorker *consumer.ConsumerWorker) *MessageProcessor {
	logger.InfoOutCtx("starting NewMessageProcessor SUCCESSFULLY")
	return &MessageProcessor{
		consumerWorker: consumerWorker,
	}
}

func (mp *MessageProcessor) Start(ctx context.Context) {
	logger.InfoOutCtx("starting MessageProcessor SUCCESSFULLY")
	
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	defer func() { 
	    if err := mp.consumerWorker.Close(); err != nil {
            logger.ErrorOutCtx("error closing consumer: %v", zap.Error(err))
        }
		logger.InfoOutCtx("stopping MessageProcessor SUCCESSFULLY")
	}()

	for {
		select {
			case <-ctx.Done():
				logger.InfoOutCtx("received signal to stop message processor")
				return
			default:
				ev := mp.consumerWorker.Consumer.Poll(100)
				if ev == nil {
					continue
				}
			switch e := ev.(type) {
				case kafka.AssignedPartitions:
					mp.consumerWorker.Consumer.Assign(e.Partitions)
				case kafka.RevokedPartitions:
					mp.consumerWorker.Consumer.Unassign()
				case kafka.PartitionEOF:
					logger.InfoOutCtx("+++++ > KAFKA reached end of partition", zap.Any("partition", e))
				case kafka.Error:
					logger.ErrorOutCtx("+++++ > KAFKA error occurred", zap.Any("error", e))
				case *kafka.Message:
					headers := extractHeaders(e.Headers)
					fmt.Println("++++++++++++ > KAFKA message received < ++++++++++++++")
					fmt.Println("+++++ > KAFKA message headers:", headers)

					msg := Message{
						Header: &headers,
						Payload: string(e.Value),
						Raw: e,
					}

					fmt.Println(zap.Any("", msg))
					fmt.Println("++++++++++++ > KAFKA message received < ++++++++++++++")
					logger.InfoOutCtx("+++++ > KAFKA message read successfully", zap.Any("message", msg))
			}
		}
	}
}

func extractHeaders(headers []kafka.Header) map[string]string {
	headerMap := make(map[string]string)
	for _, h := range headers {
		headerMap[string(h.Key)] = string(h.Value)
	}
	return headerMap
}