package controller

import (
	"context"

	"go.opentelemetry.io/otel"

	"github.com/eliezerraj/go-core/v3/logger"
	"go.uber.org/zap"
	"github.com/go-order-v2/application/domain/usecase"
	"github.com/go-order-v2/application/domain/external"
	"github.com/go-order-v2/application/domain/entity"
)

type OrderController struct {
	orderUseCase usecase.IOrderUseCase
}

func NewOrderController(orderUseCase usecase.IOrderUseCase) *OrderController {
	logger.InfoOutCtx("initializing order controller SUCCESSFULLY")

	return &OrderController{
		orderUseCase: orderUseCase,
	}
}

func (o *OrderController) OrderAdd(ctx context.Context, req external.OrderRequest) (*entity.Order, error) {
	tracer := otel.Tracer("order.controller")
	ctx, span := tracer.Start(ctx, "OrderController.OrderAdd")
	defer span.End()

	logger.Info(ctx, "order controller OrderAdd called")

	order := entity.Order{
		OrderNumber: req.OrderNumber,
		Date:        req.Date,
		Status:      req.Status,
		Currency:    req.Currency,
		Amount:      req.Amount,
		User:        req.User,
	}

	if req.CartItem != nil {
		cartItem := entity.CartItem{
			Product: entity.Product{
				Sku: req.CartItem.Product.Sku,
			},
			Quantity: req.CartItem.Quantity,
			Discount: req.CartItem.Discount,
			Currency: req.CartItem.Currency,
			Price:    req.CartItem.Price,
		}
		order.CartItem = &cartItem
	}

	res, err := o.orderUseCase.OrderAdd(ctx, order)
	if err != nil {
		logger.Error(ctx, "order controller OrderAdd failed", zap.Error(err))
		return nil, err
	}

	return res, nil
}

func (o *OrderController) OrderGet(ctx context.Context, req external.OrderRequest) (*entity.Order, error) {
	tracer := otel.Tracer("order.controller")
	ctx, span := tracer.Start(ctx, "OrderController.OrderGet")
	defer span.End()

	logger.Info(ctx, "order controller OrderGet called", zap.String("order_number", req.OrderNumber))

	order := entity.Order{
		OrderNumber: req.OrderNumber,
	}

	res, err := o.orderUseCase.OrderGet(ctx, order)
	if err != nil {
		logger.Error(ctx, "order controller OrderGet failed", zap.Error(err))
		return nil, err
	}

	return res, nil
}
