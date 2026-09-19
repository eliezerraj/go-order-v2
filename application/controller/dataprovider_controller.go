package controller

import (
	"context"
	"go.uber.org/zap"

	"github.com/go-order-v2/application/domain/entity"
	"github.com/go-order-v2/application/domain/usecase"
	"github.com/go-order-v2/application/domain/external"
	"github.com/go-order-v2/application/tracing"

	"github.com/eliezerraj/go-core/v3/logger"

	"go.opentelemetry.io/otel/trace"
)

type DataProviderController struct {
	dataProviderUsecase usecase.IDataProviderUseCase
}

func NewDataProviderController(dataProviderUsecase usecase.IDataProviderUseCase) *DataProviderController {
	logger.InfoOutCtx("initializing data provider controller SUCCESSFULLY")

	return &DataProviderController{
		dataProviderUsecase: dataProviderUsecase,
	}
}

func (c *DataProviderController) TimeSeriesOrderItemsGet(ctx context.Context, req external.TimeSeriesOrderItemRequest) (*entity.TimeSeriesOrderItem, error) {
	logger.Info(ctx, "executing TimeSeriesOrderItemsGet controller:", zap.Any("req", req))
	
	// Tracing context for the TimeSeriesOrderItemsGet operation
	ctx, span := tracing.CustomStartSpanCtx(ctx, "dataProviderController.TimeSeriesOrderItemsGet", trace.SpanKindInternal)
	defer span.End()
	
	product := entity.Product{
		Sku: req.ProductRequest.Sku,
	}

	limit := req.Limit
	offset := req.Offset

	res, err := c.dataProviderUsecase.TimeSeriesOrderItemsGet(ctx, product, limit, offset)
	if err != nil {
		logger.Error(ctx, "error executing TimeSeriesOrderItemsGet controller", zap.Error(err))
		return nil, err
	}

	return res, err
}