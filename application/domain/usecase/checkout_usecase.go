package usecase

import (
	"time"
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/eliezerraj/go-core/v3/logger"

	"github.com/go-order-v2/application/domain/entity"
	"github.com/go-order-v2/application/domain/external"
	"github.com/go-order-v2/application/infrastructure/repository"
	"github.com/go-order-v2/application/infrastructure/module"
	"github.com/go-order-v2/application/tracing"

	"github.com/jackc/pgx/v5"

	"go.opentelemetry.io/otel/trace"
)

const (
	CheckoutStatusPending = "checkout:pending"
)

type CheckoutUsecase struct {
	orderRepository    repository.IOrderRepository
	checkoutRepository repository.ICheckoutRepository
	paymentModule     module.PaymentModule
	inventoryModule   module.InventoryModule
}

type ICheckoutUseCase interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	CheckoutAdd(ctx context.Context, checkout entity.Checkout) (*entity.Checkout, error)
	CheckoutGet(ctx context.Context, checkout entity.Checkout) (*entity.Checkout, error)
	CheckoutPut(ctx context.Context, checkout entity.Checkout) (*entity.Checkout, error)
}

func NewCheckoutUseCase(orderRepository repository.IOrderRepository, checkoutRepository repository.ICheckoutRepository, paymentModule module.PaymentModule, inventoryModule module.InventoryModule) ICheckoutUseCase {
	logger.InfoOutCtx("initializing checkout usecase SUCCESSFULLY")

	return &CheckoutUsecase{
		orderRepository:    orderRepository,
		checkoutRepository: checkoutRepository,
		paymentModule:      paymentModule,
		inventoryModule:    inventoryModule,
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

	// Tracing and metrics
	ctx, span := tracing.CustomStartSpanCtx(ctx, "checkoutUsecase.CheckoutAdd", trace.SpanKindInternal)
	defer span.End()

	// Start a new transaction
	tx, err := c.checkoutRepository.BeginTx(ctx, pgx.TxOptions{ IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite })
	if err != nil {
		logger.Error(ctx, "checkout usecase CheckoutAdd failed to begin transaction", zap.Error(err))
		return nil, err
	}

	// Ensure that the transaction is either committed or rolled back
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
	// Order SECTION - Check if the order and order itens exists
	//-----------------------------------------------------------
	res_order, err := c.orderRepository.OrderGet(ctx, checkout.Order)
	if err != nil {
		logger.Error(ctx, "checkout usecase CheckoutAdd failed to get order", zap.Error(err))
		return nil, err
	}

	res_order_itens, err := c.orderRepository.OrderItensGet(ctx, *res_order)
	if err != nil {
		logger.Error(ctx, "checkout usecase CheckoutAdd failed to get order items", zap.Error(err))
		return nil, err
	}

	// Set the order items in the order response
	res_order.OrderItem = res_order_itens

	//-----------------------------------------------------------
	// Payment SECTION
	//-----------------------------------------------------------
	orderReq := external.OrderRequest{
		ID: res_order.ID,
		OrderNumber:  res_order.OrderNumber,
	}

	// Create a list of 1 payment
	listPaymentDetailsReq := make([]*external.PaymentDetailRequest, 1)

	for i := range listPaymentDetailsReq {
		paymentDetailReq := &external.PaymentDetailRequest{}
		creditCardRed := external.CreditCardRequest{
			Pan:            checkout.Payment.CreditCard.Pan,
			Holder:         checkout.Payment.CreditCard.Holder,
			Password:       checkout.Payment.CreditCard.Password,
			CVV:            checkout.Payment.CreditCard.CVV,
		}

		paymentDetailReq.Currency = checkout.Payment.Currency
		paymentDetailReq.Amount = checkout.Payment.Amount
		paymentDetailReq.DetailDate = time.Now().UTC()
		paymentDetailReq.CreditCard = &creditCardRed

		listPaymentDetailsReq[i] = paymentDetailReq
	}

	paymentRequest := external.PaymentRequest{
		TransactionID: res_order.Transaction,
		Type:	checkout.Payment.Type,
		Order: orderReq,
		PaymentDetail: listPaymentDetailsReq,
	}

	// Call the payment module to process the payment
	res_payment, err := c.paymentModule.PaymentAdd(ctx, paymentRequest)
	if err != nil {
		logger.Error(ctx, "checkout usecase CheckoutAdd failed in payment module", zap.Error(err))
		return nil, err
	}

	//-----------------------------------------------------------
	// Inventory SECTION - Update Product Stock
	//-----------------------------------------------------------
	for _, item := range *res_order_itens {
		// Call REST API to update product stock in the inventory module
		inventory := &entity.Inventory{
				Available: - item.Quantity,
				Sold: item.Quantity,
		}
		item.Product.Inventory = inventory

		err := c.inventoryModule.InventoryPatch(ctx, item.Product)
		if err != nil {
			logger.Error(ctx, "checkout usecase CheckoutAdd failed to update inventory", zap.Error(err))
			return nil, err
		}
	}

	//-----------------------------------------------------------
	// Order SECTION - Update Status to "checkout:completed"
	//-----------------------------------------------------------
	updatedAt := time.Now().UTC()
	res_order.UpdatedAt = &updatedAt
	res_order.Status = CheckoutStatusPending

	rowsAffected, err := c.orderRepository.OrderPut(ctx, tx, *res_order)
	if err != nil {
		logger.Error(ctx, "checkout usecase CheckoutAdd failed to update order status", zap.Error(err))
		return nil, err
	}
	if rowsAffected == 0 {
		logger.Warn(ctx, "checkout usecase CheckoutAdd: no rows affected, order not found", zap.Int("order_id", res_order.ID))
		return nil, nil
	}

	// -----------------------------------------------------------
	// OrderItem SECTION - Update Status to "checkout:completed" for each order item
	//-----------------------------------------------------------
	for i, item := range *res_order_itens {
		item.UpdatedAt = &updatedAt
		item.Status = CheckoutStatusPending

		rowsAffected, err := c.orderRepository.OrderItemPut(ctx, tx, item)
		if err != nil {
			logger.Error(ctx, "checkout usecase CheckoutAdd failed to update order item status", zap.Error(err))
			return nil, err
		}
		if rowsAffected == 0 {
			logger.Warn(ctx, "checkout usecase CheckoutAdd: no rows affected, order item not found", zap.Int("order_item_id", item.ID))
			return nil, nil
		}
		(*res_order_itens)[i] = item
	}

	// Set the payment details in the checkout response
	checkout.Payment = *res_payment
	checkout.Order = *res_order

	logger.Info(ctx, "checkout usecase CheckoutAdd completed SUCCESSFULLY")

	return &checkout, nil
}

// CheckoutGet retrieves a checkout from the repository based on the provided checkout details.
func (c *CheckoutUsecase) CheckoutGet(ctx context.Context, checkout entity.Checkout) (*entity.Checkout, error) {
	logger.Info(ctx, "checkout usecase CheckoutGet called")

	// Tracing.
	ctx, span := tracing.CustomStartSpanCtx(ctx, "checkoutUsecase.CheckoutGet", trace.SpanKindInternal)
	defer span.End()

	// Get the order from the repository
	res_checkout, err := c.checkoutRepository.CheckoutGet(ctx, checkout)
	if err != nil {
		logger.Error(ctx, "checkout usecase CheckoutGet failed", zap.Error(err))
		return nil, err
	}

	return res_checkout, nil
}

// CheckoutPut updates an existing checkout in the repository based on the provided checkout details.
func (c *CheckoutUsecase) CheckoutPut(ctx context.Context, checkout entity.Checkout) (*entity.Checkout, error) {
	logger.Info(ctx, "checkout usecase CheckoutPut called", zap.Any("checkout", checkout))

	// Tracing.
	ctx, span := tracing.CustomStartSpanCtx(ctx, "checkoutUsecase.CheckoutPut", trace.SpanKindInternal)
	defer span.End()

	// Start a new transaction
	tx, err := c.checkoutRepository.BeginTx(ctx, pgx.TxOptions{ IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite })
	if err != nil {
		logger.Error(ctx, "checkout usecase CheckoutPut failed to begin transaction", zap.Error(err))
		return nil, err
	}

	// Ensure that the transaction is either committed or rolled back
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && rollbackErr != pgx.ErrTxClosed {
				logger.Error(ctx, "checkout usecase CheckoutPut failed to rollback transaction", zap.Error(rollbackErr))
			}
		} else {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				logger.Error(ctx, "checkout usecase CheckoutPut failed to commit transaction", zap.Error(commitErr))
				err = commitErr
			}
		}
	}()

	// -------------------------------------------
	// Update Order in the repository
	// --------------------------------------------
	updatedAt := time.Now().UTC()
	checkout.Order.UpdatedAt = &updatedAt

	upd_checkout, err := c.checkoutRepository.CheckoutPut(ctx, tx, checkout)
	if err != nil {
		logger.Error(ctx, "checkout usecase CheckoutPut failed", zap.Error(err))
		return nil, err
	}

	if upd_checkout == 0{
		logger.Warn(ctx, "checkout usecase CheckoutPut: no rows affected, order not found", zap.String("order_number", checkout.Order.OrderNumber))
		return nil, errors.New("order not found")
	}

	// -------------------------------------------
	// Update Orderitem in the repository
	// --------------------------------------------
	res_order_itens, err := c.orderRepository.OrderItensGet(ctx, checkout.Order)
	if err != nil {
		logger.Error(ctx, "checkout usecase CheckoutAdd failed to get order items", zap.Error(err))
		return nil, err
	}

	for _, orderItem := range *res_order_itens {
		orderItem.UpdatedAt = &updatedAt
		orderItem.Status = checkout.Order.Status
		rowsAffected, err := c.orderRepository.OrderItemPut(ctx, tx, orderItem)
		if err != nil {
			logger.Error(ctx, "checkout usecase CheckoutPut failed to update order item", zap.Error(err))
			return nil, err
		}

		if rowsAffected == 0 {
			logger.Warn(ctx, "checkout usecase CheckoutPut: no rows affected, order item not found", zap.Int("order_item_id", orderItem.ID))
			return nil, errors.New("order item not found")
		}
	}

	return &checkout, nil
}