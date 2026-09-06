package event_kafka

import (
	"context"
	"go.uber.org/zap"

	"github.com/eliezerraj/go-core/v3/logger"

	"github.com/go-order-v2/application/domain/external"
	"github.com/go-order-v2/application/domain/entity"
	"github.com/go-order-v2/application/infrastructure/application"

	"go.opentelemetry.io/otel/trace"
	"github.com/go-order-v2/application/tracing"
)

type PaymentEvent struct {
	application *application.Application
}

func NewPaymentEvent(application *application.Application) *PaymentEvent {
	logger.InfoOutCtx("starting event Kafka worker process SUCCESSFULLY")
	return &PaymentEvent{
		application: application,
	}
}

func (p *PaymentEvent) ProcessPaymentMessage(ctx context.Context, payment entity.Payment) error{
	logger.Info(ctx, "Processing payment message", zap.Any("payment", payment))
	
	// Tracing and metrics
	ctx, span := tracing.CustomStartSpanCtx(ctx, "eventKafka.ProcessPaymentMessage", trace.SpanKindInternal)
	defer span.End()

	listPaymentDetailRequest := []*external.PaymentDetailRequest{}
	for _, detail := range payment.PaymentDetail {
		listPaymentDetailRequest = append(listPaymentDetailRequest, &external.PaymentDetailRequest{
			ID:     detail.ID,
			//DetailDate: detail.DetailDate,
			Status: detail.Status,
			Currency: detail.Currency,
			Amount: detail.Amount,
			CreditCard: &external.CreditCardRequest{
				Pan:     detail.CreditCard.Pan,
				Holder:  detail.CreditCard.Holder,
				CVV:     detail.CreditCard.CVV,
				Password: detail.CreditCard.Password,
			},
		})
	}

	paymentReq := external.PaymentRequest{
		ID: payment.ID,
		PaymentNumber: payment.PaymentNumber,
		PaymentDetail: listPaymentDetailRequest,
	}

	orderReq := external.OrderRequest{
		ID: payment.Order.ID,
		OrderNumber: payment.Order.OrderNumber,
		Status: listPaymentDetailRequest[0].Status,
	}

	CheckoutReq := external.CheckoutRequest{
		Order:   &orderReq,
		Payment: &paymentReq,
	}
	res, err :=p.application.CheckoutController.CheckoutPut(ctx, CheckoutReq)

	if err != nil {
		logger.Error(ctx, "Failed to process payment message", zap.Error(err))
		return err
	}

	logger.Info(ctx, " ******* Finished processing payment message ******", zap.Any("response", res))

	return nil
}