package controller

import (
	"context"
	"go.uber.org/zap"

	"github.com/eliezerraj/go-core/v3/logger"
	
	"github.com/go-order-v2/application/domain/usecase"
	"github.com/go-order-v2/application/domain/external"
	"github.com/go-order-v2/application/domain/entity"
	"github.com/go-order-v2/application/tracing"

	"go.opentelemetry.io/otel/trace"
)

type CheckoutController struct {
	checkoutUseCase usecase.ICheckoutUseCase
}

func NewCheckoutController(checkoutUseCase usecase.ICheckoutUseCase) *CheckoutController {
	logger.InfoOutCtx("initializing checkout controller SUCCESSFULLY")

	return &CheckoutController{
		checkoutUseCase: checkoutUseCase,
	}
}

func (c *CheckoutController) CheckoutAdd(ctx context.Context, req external.CheckoutRequest) (*entity.Checkout, error) {
	logger.Info(ctx, "checkout controller CheckoutAdd called")

	// Trace
	ctx, span := tracing.CustomStartSpanCtx(ctx, "checkoutController.CheckoutAdd", trace.SpanKindInternal)
	defer span.End()

	payment := entity.PaymentCheckout{
		PaymentNumber: req.Payment.PaymentNumber,
		TransactionID: req.Payment.TransactionID,
		Type:          req.Payment.Type,     
	}

	var amount float64
	for _, item := range req.Payment.PaymentDetail {
		payment.CreditCard = &entity.CreditCard{
			Pan:      item.CreditCard.Pan,
			Holder:   item.CreditCard.Holder,
			Password: item.CreditCard.Password,
			CVV:      item.CreditCard.CVV,
		}
		amount += item.Amount
	}

	payment.Currency = req.Payment.PaymentDetail[0].Currency
	payment.Amount = amount

	// Create the order entity
	checkout := entity.Checkout{
		Order: entity.Order{
			ID: req.Order.ID,
			OrderNumber: req.Order.OrderNumber,
		},
		Payment: payment,
	}

	logger.Debug(ctx,"===>checkout controller CheckoutAdd request",	zap.Any("checkout", checkout))

	// Call the use case to add the order
	res, err := c.checkoutUseCase.CheckoutAdd(ctx, checkout)
	if err != nil {
		logger.Error(ctx, "checkout controller CheckoutAdd failed", zap.Error(err))
		return nil, err
	}

	return res, nil
}

func (c *CheckoutController) CheckoutGet(ctx context.Context, req external.CheckoutRequest) (*entity.Checkout, error) {
	logger.Info(ctx, "checkout controller CheckoutGet called", zap.Any("order", req.Order))

	ctx, span := tracing.CustomStartSpanCtx(ctx, "checkoutController.CheckoutGet", trace.SpanKindInternal)
	defer span.End()

	checkout := entity.Checkout{
		Order: entity.Order{
			OrderNumber: req.Order.OrderNumber,
		},
	}

	res, err := c.checkoutUseCase.CheckoutGet(ctx, checkout)
	if err != nil {
		logger.Error(ctx, "checkout controller CheckoutGet failed", zap.Error(err))
		return nil, err
	}

	return res, nil
}

func (c *CheckoutController) CheckoutPut(ctx context.Context, req external.CheckoutRequest) (*entity.Checkout, error) {
	logger.Info(ctx, "checkout controller CheckoutPut called", zap.Any("order", req.Order))

	ctx, span := tracing.CustomStartSpanCtx(ctx, "checkoutController.CheckoutPut", trace.SpanKindInternal)
	defer span.End()

	payment := entity.PaymentCheckout{
		PaymentNumber: req.Payment.PaymentNumber,
		TransactionID: req.Payment.TransactionID,
		Type:          req.Payment.Type,     
	}

	var amount float64
	for _, item := range req.Payment.PaymentDetail {
		payment.CreditCard = &entity.CreditCard{
			Pan:      item.CreditCard.Pan,
			Holder:   item.CreditCard.Holder,
			Password: item.CreditCard.Password,
			CVV:      item.CreditCard.CVV,
		}
		amount += item.Amount
	}

	payment.Currency = req.Payment.PaymentDetail[0].Currency
	payment.Amount = amount
	statusOrder := req.Payment.PaymentDetail[0].Status // update status from event-driven payment detail

	// Create the order entity
	checkout := entity.Checkout{
		Order: entity.Order{
			ID: req.Order.ID,
			OrderNumber: req.Order.OrderNumber,
			Status: statusOrder,
		},
		Payment: payment,
	}

	res, err := c.checkoutUseCase.CheckoutPut(ctx, checkout)
	if err != nil {
		logger.Error(ctx, "checkout controller CheckoutPut failed", zap.Error(err))
		return nil, err
	}

	return res, nil
}