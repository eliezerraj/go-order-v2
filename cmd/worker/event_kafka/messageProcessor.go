package event_kafka

import (
	"context"
	"os"
	"syscall"
	"os/signal"
	"fmt"
	"encoding/json"

	"go.uber.org/zap"

	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/eliezerraj/go-core/v3/event/kafka/consumer"
	
	"github.com/go-order-v2/application/domain/entity"
    "github.com/go-order-v2/application/shared/otelkafka"
	"github.com/go-order-v2/application/tracing"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"

	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Message struct {
	Header 	*map[string]string
	Payload string
	Raw     *kafka.Message
}

type PaymentEventHandler interface {
    ProcessPaymentMessage(ctx context.Context, payment entity.Payment) error
}

type MessageProcessor struct {
	consumerWorker *consumer.ConsumerWorker
	handler        PaymentEventHandler
}

func NewMessageProcessor(consumerWorker *consumer.ConsumerWorker, handler PaymentEventHandler) *MessageProcessor {
	logger.InfoOutCtx("starting NewMessageProcessor SUCCESSFULLY")
	return &MessageProcessor{
		consumerWorker: consumerWorker,
		handler: handler,
	}
}

func (mp *MessageProcessor) Start(ctx context.Context) error {
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
				return nil
			default:
				ev := mp.consumerWorker.Consumer.Poll(5000) // Poll for Kafka events with a 5-second timeout
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
					mp.handleMessage(ctx, e)
			}
		}
	}
}

func (mp *MessageProcessor) handleMessage(ctx context.Context, e *kafka.Message) {
    msgCtx, span := extractTraceContext(ctx, e)
    defer span.End()

    fmt.Println("++++++++++++ > KAFKA message received < ++++++++++++++")
    headers := extractHeaders(e.Headers)
    fmt.Println("+++++ > KAFKA headers:", headers)

    msg := Message{
        Header:  &headers,
        Payload: string(e.Value),
    }

    var event entity.Event
    if err := json.Unmarshal(e.Value, &event); err != nil {
        logger.Error(msgCtx, "failed to unmarshal kafka event", zap.Error(err))
        return
    }

    dataBytes, err := json.Marshal(event.Data)
    if err != nil {
        logger.Error(msgCtx, "failed to marshal kafka event data", zap.Error(err))
        return
    }

    var payment entity.Payment
    if err := json.Unmarshal(dataBytes, &payment); err != nil {
        logger.Error(msgCtx, "failed to unmarshal kafka payment", zap.Error(err))
        return
    }

    fmt.Println("+++++ > KAFKA Payload:", msg.Payload)
    fmt.Println("++++++++++++ > Finished processing KAFKA message < ++++++++++++++")

    if err := mp.handler.ProcessPaymentMessage(msgCtx, payment); err != nil {
        logger.Error(msgCtx, "failed to process payment message", zap.Error(err))
        return
    }
    
    mp.consumerWorker.Consumer.CommitMessage(e)
}

func extractHeaders(headers []kafka.Header) map[string]string {
	headerMap := make(map[string]string)
	for _, h := range headers {
		headerMap[string(h.Key)] = string(h.Value)
	}
	return headerMap
}

func extractTraceContext(ctx context.Context, msg *kafka.Message) (context.Context, trace.Span) {
    // 1. Wrap kafka.Headers with your carrier
    carrier := otelkafka.KafkaHeaderCarrier(msg.Headers)

    // 2. Extract parent trace context from incoming message headers
    parentCtx := otel.GetTextMapPropagator().Extract(ctx, &carrier)

    // 3. Extract x-request-id if present and attach to context
    requestID := carrier.Get("x-request-id")
    if requestID != "" {
        parentCtx = context.WithValue(parentCtx, tracing.RequestIDHeaderName, requestID)
    }

    // 4. Start a CONSUMER span as child of the incoming trace
    topic := ""
    if msg.TopicPartition.Topic != nil {
        topic = *msg.TopicPartition.Topic
    }

    tracer := otel.Tracer("order-worker-consumer")
    msgCtx, span := tracer.Start(
        parentCtx,
        topic+" receive",
        trace.WithSpanKind(trace.SpanKindConsumer),
        trace.WithAttributes(
            semconv.MessagingSystemKafka,
            semconv.MessagingDestinationName(topic),
            semconv.MessagingKafkaMessageKey(string(msg.Key)),
        ),
    )

	// Verify span and attributes
    if roSpan, ok := span.(sdktrace.ReadOnlySpan); ok {
        for _, attr := range roSpan.Attributes() {
            fmt.Printf("DEBUG SEMCONV -> %s: %v\n", attr.Key, attr.Value.AsInterface())
        }
    }

    return msgCtx, span
}