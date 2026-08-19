package controller

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/eliezerraj/go-core/v3/logger"
	"go.uber.org/zap"
	"github.com/go-order-v2/application/domain/usecase"
	"github.com/go-order-v2/application/domain/external"
	"github.com/go-order-v2/application/domain/entity"
	"github.com/go-order-v2/application/shared/helpers"
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

	var orderDate *time.Time
	if req.Date != "" {
		parsedDate, err := helpers.ParseDate(req.Date)
		if err != nil {
			logger.Error(ctx, "failed to parse order date", zap.Error(err), zap.String("date", req.Date))
			return nil, err
		}
		orderDate = parsedDate
	}

	order := entity.Order{
		OrderNumber: req.OrderNumber,
		Date:        *orderDate,
		CustomerID:  req.CustomerID,
	}

	if req.OrderItem != nil {
		orderItem := entity.OrderItem{
			Product: entity.Product{
				Sku: req.OrderItem[0].Product.Sku,
			},
			Quantity: req.OrderItem[0].Quantity,
			Discount: req.OrderItem[0].Discount,
			Currency: req.OrderItem[0].Product.Price.Currency,
			Price:    req.OrderItem[0].Product.Price.Amount,
		}
		order.OrderItem = &orderItem
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
