package usecase

import (
	"time"
	"context"

	"go.uber.org/zap"

	"github.com/eliezerraj/go-core/v3/logger"

	"github.com/go-order-v2/application/domain/entity"
	"github.com/go-order-v2/application/infrastructure/repository"

	"github.com/jackc/pgx/v5"

	"go.opentelemetry.io/otel"
)

const (
	// OrderStatusPending represents the pending status of an order.
	OrderStatusPending = "pending"
	// OrderStatusCompleted represents the completed status of an order.
	OrderStatusCompleted = "completed"
)

type OrderUsecase struct {
	orderRepository repository.IOrderRepository
}

type IOrderUseCase interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	OrderAdd(ctx context.Context, order entity.Order) (*entity.Order, error)
	OrderGet(ctx context.Context, order entity.Order) (*entity.Order, error)
}

func NewOrderUseCase(orderRepository repository.IOrderRepository) IOrderUseCase {
	logger.InfoOutCtx("initializing order usecase SUCCESSFULLY")

	return &OrderUsecase{
		orderRepository: orderRepository,
	}
}

func (o *OrderUsecase) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	logger.Info(ctx, "order usecase BeginTx called")

	tx, err := o.orderRepository.BeginTx(ctx, opts)
	if err != nil {
		logger.Error(ctx, "order usecase BeginTx failed", zap.Error(err))
		return nil, err
	}

	return tx, nil
}

func (o *OrderUsecase) OrderAdd(ctx context.Context, order entity.Order) (*entity.Order, error) {
	tracer := otel.Tracer("order.repository")
	ctx, span := tracer.Start(ctx, "OrderUsecase.OrderAdd")
	defer span.End()

	logger.Info(ctx, "order usecase OrderAdd called")

	tx, err := o.orderRepository.BeginTx(ctx, pgx.TxOptions{ IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite })
	if err != nil {
		logger.Error(ctx, "order usecase OrderAdd failed to begin transaction", zap.Error(err))
		return nil, err
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				logger.Error(ctx, "order usecase OrderAdd failed to rollback transaction", zap.Error(rollbackErr))
			}
		} else {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				logger.Error(ctx, "order usecase OrderAdd failed to commit transaction", zap.Error(commitErr))
				err = commitErr
			}
		}
	}()

	// Business logic: Set default values for order
	createAt := time.Now().UTC()
	order.CreatedAt = createAt 

	if order.Date == (time.Time{}) {
		order.Date = createAt
	}

	order.Status = OrderStatusPending
	order.Transaction = "txn_" + order.OrderNumber

	// Create the order item first if it exists
	if order.OrderItem != nil {
		orderItem := *order.OrderItem
		orderItem.Status = OrderStatusPending
		orderItem.CreatedAt = createAt

		resOrderItem, err := o.orderRepository.OrderItemAdd(ctx, orderItem)
		if err != nil {
			logger.Error(ctx, "order usecase OrderAdd failed to add order item", zap.Error(err))
			return nil, err
		}
		order.OrderItem.ID = resOrderItem.ID
	}

	// Add the order to the repository
	res, err := o.orderRepository.OrderAdd(ctx, order)
	if err != nil {
		logger.Error(ctx, "order usecase OrderAdd failed", zap.Error(err))
		return nil, err
	}

	if err != nil {
		logger.Error(ctx, "order usecase OrderAdd failed to commit transaction", zap.Error(err))
		return nil, err
	}

	logger.Info(ctx, "order usecase OrderAdd completed SUCCESSFULLY")
	return res, nil
}

func (o *OrderUsecase) OrderGet(ctx context.Context, order entity.Order) (*entity.Order, error) {
	tracer := otel.Tracer("order.repository")
	ctx, span := tracer.Start(ctx, "OrderUsecase.OrderGet")
	defer span.End()

	logger.Info(ctx, "order usecase OrderGet called")

	res, err := o.orderRepository.OrderGet(ctx, order)
	if err != nil {
		logger.Error(ctx, "order usecase OrderGet failed", zap.Error(err))
		return nil, err
	}

	return res, nil
}
