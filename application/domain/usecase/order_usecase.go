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
	OrderStatusPending = "order:pending"
	// OrderStatusCompleted represents the completed status of an order.
	OrderStatusCompleted = "completed"
)

type OrderUsecase struct {
	orderRepository repository.IOrderRepository
	inventoryModule module.InventoryModule
}

type IOrderUseCase interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	OrderAdd(ctx context.Context, order entity.Order) (*entity.Order, error)
	OrderGet(ctx context.Context, order entity.Order) (*entity.Order, error)
}

func NewOrderUseCase(orderRepository repository.IOrderRepository, inventoryModule module.InventoryModule) IOrderUseCase {
	logger.InfoOutCtx("initializing order usecase SUCCESSFULLY")

	return &OrderUsecase{
		orderRepository: orderRepository,
		inventoryModule: inventoryModule,
	}
}

// BeginTx starts a new database transaction with the specified options.
func (o *OrderUsecase) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	logger.Info(ctx, "order usecase BeginTx called")

	tx, err := o.orderRepository.BeginTx(ctx, opts)
	if err != nil {
		logger.Error(ctx, "order usecase BeginTx failed", zap.Error(err))
		return nil, err
	}

	return tx, nil
}

// AddOrder adds a new order to the repository.
func (o *OrderUsecase) OrderAdd(ctx context.Context, order entity.Order) (res_order *entity.Order, err error) {
	logger.Info(ctx, "order usecase OrderAdd called")

	// Tracing
	tracer := otel.Tracer("order.repository")
	ctx, span := tracer.Start(ctx, "OrderUsecase.OrderAdd")
	defer span.End()

	// Start a new transaction
	tx, err := o.orderRepository.BeginTx(ctx, pgx.TxOptions{ IsoLevel: pgx.ReadCommitted, AccessMode: pgx.ReadWrite })
	if err != nil {
		logger.Error(ctx, "order usecase OrderAdd failed to begin transaction", zap.Error(err))
		return nil, err
	}

	// Ensure that the transaction is either committed or rolled back
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && rollbackErr != pgx.ErrTxClosed {
				logger.Error(ctx, "order usecase OrderAdd failed to rollback transaction", zap.Error(rollbackErr))
			}
		} else {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				logger.Error(ctx, "order usecase OrderAdd failed to commit transaction", zap.Error(commitErr))
				err = commitErr
			}
		}
	}()

	//-----------------------------------------------------------
	// Inventory SECTION - Get Product Price
	//-----------------------------------------------------------
	// Initialize order amount
	var orderAmount float64 = 0.0	
	
	// Create a list to hold order items
	listOrderItem := []entity.OrderItem{}

	forderItem := order.OrderItem; for _, item := range *forderItem {
		// Call REST API to get product details from the inventory module
		res_inventory, err := o.inventoryModule.GetInventory(ctx, item.Product)
		if err != nil {
			logger.Error(ctx, "order usecase OrderAdd failed to check inventory", zap.Error(err))
			return nil, err
		}

		orderItem := entity.OrderItem{
			Product:   *res_inventory,
			Currency:  res_inventory.Price.Currency,
			Amount:    res_inventory.Price.Amount * float64(item.Quantity),
			Quantity:  item.Quantity,
			Discount:  item.Discount,
		}

		// Calculate the order amount based on product price, quantity, and discount
		orderAmount += res_inventory.Price.Amount * float64(item.Quantity) - (item.Discount * -1)
		listOrderItem = append(listOrderItem, orderItem)
	}

	//-----------------------------------------------------------
	// Ordem SECTION - Create Order Business logic: Set default values for order
	//-----------------------------------------------------------
	createAt := time.Now().UTC()
	order.CreatedAt = createAt 
	if order.Date == (time.Time{}) {
		order.Date = createAt
	}
	order.Status = OrderStatusPending
	order.Transaction = "txn_" + order.OrderNumber
	order.Currency = listOrderItem[0].Currency
	order.Amount = orderAmount

	// Add the order to the repository
	res_order, err = o.orderRepository.OrderAdd(ctx, tx, order)
	if err != nil {
		logger.Error(ctx, "order usecase OrderAdd failed", zap.Error(err))
		return nil, err
	}

	//-----------------------------------------------------------
	// Ordem Item SECTION - Create Order Items
	//-----------------------------------------------------------
	for _, item := range listOrderItem {
		// Set the foreign key for the order item to the newly created order ID
		item.FkOrderID = res_order.ID
		item.Status = OrderStatusPending
		item.CreatedAt = createAt

		// Add the order item to the repository
		res_order_item, err := o.orderRepository.OrderItemAdd(ctx, tx, item)
		if err != nil {
			logger.Error(ctx, "order usecase OrderAdd failed to add order item", zap.Error(err))
			return nil, err
		}
		item.ID = res_order_item.ID
	}

	//-----------------------------------------------------------
	// Ordem SECTION - Set the list of order items in the order response
	//-----------------------------------------------------------
	res_order.OrderItem = &listOrderItem

	logger.Info(ctx, "order usecase OrderAdd completed SUCCESSFULLY")

	return res_order, nil
}

// GetOrder retrieves an order from the repository based on the provided order details.
func (o *OrderUsecase) OrderGet(ctx context.Context, order entity.Order) (*entity.Order, error) {
	logger.Info(ctx, "order usecase OrderGet called")

	// Tracing
	tracer := otel.Tracer("order.repository")
	ctx, span := tracer.Start(ctx, "OrderUsecase.OrderGet")
	defer span.End()

	// Get the order from the repository
	res_order, err := o.orderRepository.OrderGet(ctx, order)
	if err != nil {
		logger.Error(ctx, "order usecase OrderGet failed", zap.Error(err))
		return nil, err
	}

	// Get order items for the order
	res_order_itens, err := o.orderRepository.OrderItensGet(ctx, *res_order)
	if err != nil {
		logger.Error(ctx, "order usecase OrderGet failed to get order items", zap.Error(err))
		return nil, err
	}

	// Get Product details for each order item from the inventory module
	for _, item := range *res_order_itens {
		res_product, err := o.inventoryModule.GetInventory(ctx, item.Product)
		if err != nil {
			logger.Error(ctx, "order usecase OrderGet failed to check inventory", zap.Error(err))
			return nil, err
		}
		item.Product = *res_product
	}

	res_order.OrderItem = res_order_itens

	return res_order, nil
}
