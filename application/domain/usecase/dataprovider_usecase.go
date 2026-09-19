package usecase

import (
	"context"
	
	"go.uber.org/zap"

	"github.com/go-order-v2/application/domain/entity"
	"github.com/go-order-v2/application/infrastructure/repository"
	"github.com/go-order-v2/application/infrastructure/module"
	"github.com/go-order-v2/application/tracing"

	"github.com/eliezerraj/go-core/v3/logger"

	"go.opentelemetry.io/otel/trace"
)

type DataProviderUseCase struct {
	dataproviderRepository repository.IDataProvider
	inventoryModule module.InventoryModule
}

type IDataProviderUseCase interface {
	TimeSeriesOrderItemsGet(ctx context.Context, product entity.Product, limit int, offset int) (*entity.TimeSeriesOrderItem, error)
}

func NewDataProviderUseCase(dataproviderRepository repository.IDataProvider, inventoryModule module.InventoryModule) IDataProviderUseCase {
	logger.InfoOutCtx("initializing TimeSeriesOrderItemsGet usecase SUCCESSFULLY")
	
	return &DataProviderUseCase{
		dataproviderRepository: dataproviderRepository,
		inventoryModule:        inventoryModule,
	}
}

func (u *DataProviderUseCase) TimeSeriesOrderItemsGet(ctx context.Context, product entity.Product, limit int, offset int) (*entity.TimeSeriesOrderItem, error) {
	logger.Info(ctx, "executing TimeSeriesOrderItemsGet usecase:", zap.Any("product", product), zap.Int("limit", limit), zap.Int("offset", offset))

	// Tracing.
	ctx, span := tracing.CustomStartSpanCtx(ctx, "dataproviderUsecase.TimeSeriesOrderItemsGet", trace.SpanKindInternal)
	defer span.End()

	res_inventory, err := u.inventoryModule.GetInventory(ctx, product)
	if err != nil {
		logger.Error(ctx, "dataprovider usecase TimeSeriesOrderItemsGet failed to check inventory", zap.Error(err))
		return nil, err
	}

	// Set the product with the inventory information (get the product ID)
	product = *res_inventory

	// Get the order from the repository
	res_timeseries, err := u.dataproviderRepository.TimeSeriesOrderItemsGet(ctx, product, limit, offset)
	if err != nil {
		logger.Error(ctx, "dataprovider usecase TimeSeriesOrderItemsGet failed", zap.Error(err))
		return nil, err
	}

	return res_timeseries, nil
}