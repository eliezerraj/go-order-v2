package controller

import (
	"context"
	"time"
	"go.uber.org/zap"

	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/go-order-v2/application/controller/validator"
	"github.com/go-order-v2/application/domain/usecase"
	"github.com/go-order-v2/application/domain/external"
	"github.com/go-order-v2/application/domain/entity"
	"github.com/go-order-v2/application/shared/helpers"
	"github.com/go-order-v2/application/tracing"

	"go.opentelemetry.io/otel/trace"
)

type OrderController struct {
	schema validator.Schema
	orderUseCase usecase.IOrderUseCase
}

func NewOrderController(orderUseCase usecase.IOrderUseCase) *OrderController {
	logger.InfoOutCtx("initializing order controller SUCCESSFULLY")

	schema := validator.Schema{
		Validate: func(ctx context.Context, data any) error {
			return nil
		},
	}

	return &OrderController{
		schema: schema,
		orderUseCase: orderUseCase,
	}
}

func (o *OrderController) OrderAdd(ctx context.Context, req external.OrderRequest) (*entity.Order, error) {
	logger.Info(ctx, "order controller OrderAdd called")

	// Tracing context for the OrderAdd operation
	ctx, span := tracing.CustomStartSpanCtx(ctx, "orderController.OrderAdd", trace.SpanKindInternal)
	defer span.End()

	// Schema validation
	if err := o.schema.OrderAddSchema().Validate(ctx, req); err != nil {
		return nil, err
	}

	var orderDate time.Time
	if req.Date != "" {
		parsedDate, err := helpers.ParseDate(req.Date)
		if err != nil {
			logger.Error(ctx, "failed to parse order date", zap.Error(err), zap.String("date", req.Date))
			return nil, err
		}
		orderDate = *parsedDate
	}

	// Load order items from the request and create entity.OrderItem objects
	var orderItems []entity.OrderItem
	if req.OrderItem != nil {
		orderItems = make([]entity.OrderItem, 0, len(req.OrderItem))
		for _, item := range req.OrderItem {
			if item == nil {
				continue
			}
			orderItem := entity.OrderItem{
				Quantity: item.Quantity,
				Discount: item.Discount,
				Product: entity.Product{
					Sku: item.Product.Sku,
				},
			}
			orderItems = append(orderItems, orderItem)
		}
	}

	// Create the order entity
	order := entity.Order{
		OrderNumber: req.OrderNumber,
		Date:        orderDate,
		CustomerID:  req.CustomerID,
		OrderItem:   &orderItems,
	}

	// Call the use case to add the order
	res, err := o.orderUseCase.OrderAdd(ctx, order)
	if err != nil {
		logger.Error(ctx, "order controller OrderAdd failed", zap.Error(err))
		return nil, err
	}

	return res, nil
}

func (o *OrderController) OrderGet(ctx context.Context, req external.OrderRequest) (*entity.Order, error) {
	logger.Info(ctx, "order controller OrderGet called", zap.String("order_number", req.OrderNumber))
	
	// Tracing context for the OrderGet operation
	ctx, span := tracing.CustomStartSpanCtx(ctx, "orderController.OrderGet", trace.SpanKindInternal)
	defer span.End()

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
