package event_kafka

import (
	"context"

	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/go-order-v2/application/domain/entity"
)

type PaymentEvent struct {
	Interval   int
	MaxRetries int
}

func NewPaymentEvent(interval int, maxRetries int) *PaymentEvent {
	logger.InfoOutCtx("starting event Kafka worker process SUCCESSFULLY")
	return &PaymentEvent{
		Interval:   interval,
		MaxRetries: maxRetries,
	}
}

func (p *PaymentEvent) ProcessPaymentMessage(ctx context.Context, event entity.Event){
	logger.Info(ctx, "Processing payment message")
	//retries := 0
	logger.Info(ctx, " ******* Finished processing payment message ******")
}