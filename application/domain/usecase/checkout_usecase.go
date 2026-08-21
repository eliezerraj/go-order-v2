package usecase

import (
	"time"
	"context"

	"go.uber.org/zap"

	"github.com/eliezerraj/go-core/v3/logger"

	"github.com/go-order-v2/application/domain/entity"
	"github.com/go-order-v2/application/infrastructure/repository"
	"github.com/go-order-v2/application/infrastructure/module"

	"github.com/jackc/pgx/v5"

	"go.opentelemetry.io/otel"
)

const (
	// OrderStatusPending represents the pending status of an order.
	CheckoutStatusPending = "pending"
	// CheckoutStatusCompleted represents the completed status of an order.
	CheckoutStatusCompleted = "completed"
)

type CheckoutUsecase struct {
	checkoutRepository repository.ICheckoutRepository
	paymentModule     module.PaymentModule
}

type ICheckoutUseCase interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	CheckoutAdd(ctx context.Context, checkout entity.Checkout) (*entity.Checkout, error)
	CheckoutGet(ctx context.Context, checkout entity.Checkout) (*entity.Checkout, error)
}

func NewCheckoutUseCase(checkoutRepository repository.ICheckoutRepository, paymentModule module.PaymentModule) ICheckoutUseCase {
	logger.InfoOutCtx("initializing checkout usecase SUCCESSFULLY")

	return &CheckoutUsecase{
		checkoutRepository: checkoutRepository,
		paymentModule:    paymentModule,
	}
}

// BeginTx starts a new database transaction with the specified options.
func (c *CheckoutUsecase) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	logger.Info(ctx, "checkout usecase BeginTx called")

	tx, err := c.checkoutRepository.BeginTx(ctx, opts)
	if err != nil {
		logger.Error(ctx, "checkout usecase BeginTx failed", zap.Error(err))
		return nil, err
	}

	return tx, nil
}

// AddOrder adds a new order to the repository.
func (c *CheckoutUsecase) CheckoutAdd(ctx context.Context, checkout entity.Checkout) (res_checkout *entity.Checkout, err error) {
	logger.Info(ctx, "checkout usecase CheckoutAdd called")

	tracer := otel.Tracer("order.repository")
	ctx, span := tracer.Start(ctx, "CheckoutUsecase.CheckoutAdd")
	defer span.End()

	tx, err := c.checkoutRepository.BeginTx(ctx, pgx.TxOptions{ IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite })
	if err != nil {
		logger.Error(ctx, "checkout usecase CheckoutAdd failed to begin transaction", zap.Error(err))
		return nil, err
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && rollbackErr != pgx.ErrTxClosed {
				logger.Error(ctx, "checkout usecase CheckoutAdd failed to rollback transaction", zap.Error(rollbackErr))
			}
		} else {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				logger.Error(ctx, "checkout usecase CheckoutAdd failed to commit transaction", zap.Error(commitErr))
				err = commitErr
			}
		}
	}()

	//-----------------------------------------------------------
	// Order SECTION
	//-----------------------------------------------------------

	// Business logic: Set default values for order
	createAt := time.Now().UTC()
	checkout.Order.CreatedAt = createAt 
	if checkout.Order.Date == (time.Time{}) {
		checkout.Order.Date = createAt
	}
	checkout.Order.Status = CheckoutStatusPending
	checkout.Order.Transaction = "txn_" + checkout.Order.OrderNumber

	// Add the order to the repository
	res_checkout, err = c.checkoutRepository.CheckoutAdd(ctx, tx, checkout)
	if err != nil {
		logger.Error(ctx, "checkout usecase CheckoutAdd failed", zap.Error(err))
		return nil, err
	}

	//-----------------------------------------------------------
	// Ordem Item SECTION
	//-----------------------------------------------------------

	return res_checkout, nil
}

// CheckoutGet retrieves a checkout from the repository based on the provided checkout details.
func (c *CheckoutUsecase) CheckoutGet(ctx context.Context, checkout entity.Checkout) (*entity.Checkout, error) {
	tracer := otel.Tracer("checkout.repository")
	ctx, span := tracer.Start(ctx, "CheckoutUsecase.CheckoutGet")
	defer span.End()

	logger.Info(ctx, "checkout usecase CheckoutGet called")

	// Get the order from the repository
	res_checkout, err := c.checkoutRepository.CheckoutGet(ctx, checkout)
	if err != nil {
		logger.Error(ctx, "checkout usecase CheckoutGet failed", zap.Error(err))
		return nil, err
	}

	return res_checkout, nil
}
