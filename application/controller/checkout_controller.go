package controller

import (
	"context"
	"errors"
	"go.uber.org/zap"

	"go.opentelemetry.io/otel"

	"github.com/eliezerraj/go-core/v3/logger"
	
	"github.com/go-order-v2/application/domain/usecase"
	"github.com/go-order-v2/application/domain/external"
	"github.com/go-order-v2/application/domain/entity"
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
	tracer := otel.Tracer("checkout.controller")
	ctx, span := tracer.Start(ctx, "CheckoutController.CheckoutAdd")
	defer span.End()

	logger.Info(ctx, "checkout controller CheckoutAdd called")

	var payment entity.Payment
	if req.CreditCardRequest != nil {
		payment.CreditCard = &entity.CreditCard{
			Pan:      req.CreditCardRequest.Pan,
			Holder:   req.CreditCardRequest.Holder,
			Password: req.CreditCardRequest.Password,
			CVV:      req.CreditCardRequest.CVV,
		}
	} else {
		logger.Error(ctx, "CreditCardRequest is nil in CheckoutRequest")
		return nil, errors.New("CreditCard is not provided informed")
	}

	// Create the order entity
	checkout := entity.Checkout{
		Order: entity.Order{
			OrderNumber: req.OrderNumber,
		},
		Payment: payment,
	}

	// Call the use case to add the order
	res, err := c.checkoutUseCase.CheckoutAdd(ctx, checkout)
	if err != nil {
		logger.Error(ctx, "checkout controller CheckoutAdd failed", zap.Error(err))
		return nil, err
	}

	return res, nil
}

func (c *CheckoutController) CheckoutGet(ctx context.Context, req external.CheckoutRequest) (*entity.Checkout, error) {
	tracer := otel.Tracer("checkout.controller")
	ctx, span := tracer.Start(ctx, "CheckoutController.CheckoutGet")
	defer span.End()

	logger.Info(ctx, "checkout controller CheckoutGet called", zap.String("order_number", req.OrderNumber))

	checkout := entity.Checkout{
		Order: entity.Order{
			OrderNumber: req.OrderNumber,
		},
	}

	res, err := c.checkoutUseCase.CheckoutGet(ctx, checkout)
	if err != nil {
		logger.Error(ctx, "checkout controller CheckoutGet failed", zap.Error(err))
		return nil, err
	}

	return res, nil
}