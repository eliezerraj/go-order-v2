package repository

import (
	"time"
	"context"
	"errors"
	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"

	"github.com/go-order-v2/application/domain/entity"

	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/eliezerraj/go-core/v3/database/connector"

	"go.opentelemetry.io/otel"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/attribute"
)

type OrderRepository struct {
	dbConnector connector.IDatabaseConnector
}

type IOrderRepository interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	OrderAdd(ctx context.Context, tx pgx.Tx, order entity.Order) (*entity.Order, error)
	OrderGet(ctx context.Context, order entity.Order) (*entity.Order, error)
	OrderPut(ctx context.Context, tx pgx.Tx, order entity.Order) (int64, error)
	OrderItemAdd(ctx context.Context, tx pgx.Tx, orderItem entity.OrderItem) (*entity.OrderItem, error)
	OrderItensGet(ctx context.Context, order entity.Order) (*[]entity.OrderItem, error)
}

func NewOrderRepository(dbConnector connector.IDatabaseConnector) IOrderRepository {
	logger.InfoOutCtx("initializing order repository SUCCESSFULLY")

	return &OrderRepository{
		dbConnector: dbConnector,
	}
}

func (p *OrderRepository) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	logger.InfoOutCtx("order repository BeginTx called")

	tx, err := p.dbConnector.Writer().BeginTx(ctx, opts)
	if err != nil {
		logger.ErrorOutCtx("order repository BeginTx failed", zap.Error(err))
		return nil, err
	}

	return tx, nil
}

func (p *OrderRepository) OrderGet(ctx context.Context, order entity.Order) (res_order *entity.Order, err error) {
	logger.Info(ctx, "order repository OrderGet called")

	// Tracing and metrics
	tracer := otel.Tracer("order.repository")
    ctx, span := tracer.Start(ctx, "OrderRepository.OrderGet")
    defer span.End()

    meter := otel.Meter("go-order-v2.repository")
    counter, _ := meter.Int64Counter("db_custom_order_get_requests_total")
    histogram, _ := meter.Float64Histogram("db_custom_order_get_duration_seconds")
    start := time.Now()

    counter.Add(ctx, 1, metric.WithAttributes(
        attribute.String("operation", "OrderGet"),
    ))

	// Defer function to handle error logging and metrics recording	
	defer func() {
		if err != nil {
			span.RecordError(err) 
			span.SetStatus(codes.Error, err.Error())
			logger.Error(ctx, "order repository OrderGet failed", zap.Error(err))
		}
        histogram.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(
            attribute.String("operation", "OrderGet"),
        ))
	}()

	// Get a reader connection from the database connector
	connectorReader := p.dbConnector.Reader()

	query := `select o.id,
					 o.order_number,
					 o.transaction_id,
					 o.order_date,
					 o.status,
					 o.currency,
					 o.amount,
					 o.customer_id,
					 o.created_at,
					 o.updated_at
				from public.order o
				where o.order_number = $1`

	rows, err := connectorReader.Query(ctx, query, order.OrderNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	order.OrderItem = &[]entity.OrderItem{}
	if rows.Next() {
		err = rows.Scan(&order.ID, &order.OrderNumber, &order.Transaction, &order.Date, &order.Status, &order.Currency, &order.Amount, &order.CustomerID, &order.CreatedAt, &order.UpdatedAt)
		if err != nil {
			return nil, err
		}
	} else {
		logger.Warn(ctx, "not found", zap.Int("order_id", order.ID))
		err = errors.New("order not found")
		return nil, err
	}

	return &order, nil
}

func (p *OrderRepository) OrderItensGet(ctx context.Context, order entity.Order) ( res_order_itens *[]entity.OrderItem, err error) {
	logger.Info(ctx, "order repository OrderItensGet called")

	// Tracing and metrics
	tracer := otel.Tracer("order.repository")
	ctx, span := tracer.Start(ctx, "OrderRepository.OrderItensGet")
	defer span.End()

	meter := otel.Meter("go-order-v2.repository")
    counter, _ := meter.Int64Counter("db_custom_order_itens_get_requests_total")
    histogram, _ := meter.Float64Histogram("db_custom_order_itens_get_duration_seconds")
    start := time.Now()

    counter.Add(ctx, 1, metric.WithAttributes(
        attribute.String("operation", "OrderItensGet"),
    ))

	// Defer function to handle error logging and metrics recording	
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			logger.Error(ctx, "order repository OrderItensGet failed", zap.Error(err))
		}
		histogram.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(
			attribute.String("operation", "OrderItensGet"),
		))
	}()

	// Get a reader connection from the database connector
	connectorReader := p.dbConnector.Reader()

	query := `select o.id,
					 o.fk_order_id,
					 o.fk_product_id,
					 o.status,
					 o.quantity,
					 o.discount,
					 o.currency,
					 o.amount,
					 o.created_at,
					 o.updated_at
				from public.order_item o
				where o.fk_order_id = $1`

	rows, err := connectorReader.Query(ctx, query, order.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderItems := []entity.OrderItem{}
	for rows.Next() {
		var item entity.OrderItem
		var itemProduct entity.Product
		err = rows.Scan(&item.ID, &item.FkOrderID, &itemProduct.ID, &item.Status, &item.Quantity, &item.Discount, &item.Currency, &item.Amount, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		item.Product = itemProduct
		orderItems = append(orderItems, item)
	}

	res_order_itens = &orderItems
	return res_order_itens, nil
}

func (p *OrderRepository) OrderAdd(ctx context.Context, tx pgx.Tx, order entity.Order) (res_order *entity.Order, err error) {
	logger.Info(ctx, "order repository OrderAdd called")

	// Tracing and metrics
	tracer := otel.Tracer("order.repository")
	ctx, span := tracer.Start(ctx, "OrderRepository.OrderAdd")
	defer span.End()
	
	meter := otel.Meter("go-order-v2.repository")
    counter, _ := meter.Int64Counter("db_custom_order_add_requests_total")
    histogram, _ := meter.Float64Histogram("db_custom_order_add_duration_seconds")
    start := time.Now()

    counter.Add(ctx, 1, metric.WithAttributes(
        attribute.String("operation", "OrderAdd"),
    ))

	// Defer function to handle error logging and metrics recording
	defer func() {
		if err != nil {
			span.RecordError(err) 
			span.SetStatus(codes.Error, err.Error())
			logger.Error(ctx, "order repository OrderAdd failed", zap.Error(err))
		}
		histogram.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(
			attribute.String("operation", "OrderAdd"),
		))
	}()

	// Insert the order into the database
	query := `INSERT INTO public.order (order_number,
										transaction_id,
										order_date,
										customer_id,
										status,
										currency,
										amount,
										created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	rows := tx.QueryRow(ctx, query, order.OrderNumber, order.Transaction, order.Date, order.CustomerID, order.Status, order.Currency, order.Amount, order.CreatedAt)

	var id int
	if err := rows.Scan(&id); err != nil {
		return nil, err
	}

	order.ID = id
	return &order, nil
}

func (p *OrderRepository) OrderItemAdd(ctx context.Context, tx pgx.Tx, orderItem entity.OrderItem) (res_order_item *entity.OrderItem, err error) {
	logger.Info(ctx, "order repository OrderItemAdd called", zap.Any("order_item", orderItem))

	// Tracing and metrics
	tracer := otel.Tracer("order.repository")
	ctx, span := tracer.Start(ctx, "OrderRepository.OrderItemAdd")
	defer span.End()

	meter := otel.Meter("go-order-v2.repository")
    counter, _ := meter.Int64Counter("db_custom_order_item_add_requests_total")
    histogram, _ := meter.Float64Histogram("db_custom_order_item_add_duration_seconds")
    start := time.Now()

    counter.Add(ctx, 1, metric.WithAttributes(
        attribute.String("operation", "OrderItemAdd"),
    ))

	// Defer function to handle error logging and metrics recording
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			logger.Error(ctx, "order repository OrderItemAdd failed", zap.Error(err))
		}
		histogram.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(
			attribute.String("operation", "OrderItemAdd"),
		))
	}()

	// Insert the order item into the database
	query := `INSERT INTO public.order_item (	fk_order_id,
												fk_product_id,
												status,
												quantity,
												discount,
												currency,
												amount,
												created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`

	rows := tx.QueryRow(ctx, query, orderItem.FkOrderID, orderItem.Product.ID, orderItem.Status, orderItem.Quantity, orderItem.Discount, orderItem.Currency, orderItem.Amount, orderItem.CreatedAt)

	var id int
	if err := rows.Scan(&id); err != nil {
		return nil, err
	}

	orderItem.ID = id
	return &orderItem, nil
}

func (p *OrderRepository) OrderPut(ctx context.Context, tx pgx.Tx, order entity.Order) (rowsAffected int64, err error) {
	logger.Info(ctx, "order repository OrderPut called")

	// Tracing and metrics
	tracer := otel.Tracer("order.repository")
	ctx, span := tracer.Start(ctx, "OrderRepository.OrderPut")
	defer span.End()

	meter := otel.Meter("go-order-v2.repository")
	counter, _ := meter.Int64Counter("db_custom_order_put_requests_total")
	histogram, _ := meter.Float64Histogram("db_custom_order_put_duration_seconds")
	start := time.Now()

	counter.Add(ctx, 1, metric.WithAttributes(
        attribute.String("operation", "OrderPut"),
    ))

	// Defer function to handle error logging and metrics recording
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			logger.Error(ctx, "order repository OrderPut failed", zap.Error(err))
		}
		histogram.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(
			attribute.String("operation", "OrderPut"),
		))
	}()

	query := `UPDATE public.order
				SET status = $1,
					updated_at = $2
				WHERE id = $3
				RETURNING id`

	row, err := tx.Exec(ctx, query, order.Status, order.UpdatedAt, order.ID)
	if err != nil {
		return 0, err
	}

	rowsAffected = row.RowsAffected()
	if rowsAffected == 0 {
		logger.Warn(ctx, "order repository OrderPut: no rows affected, order not found", zap.Int("order_id", order.ID))
		return 0, nil
	}

	return rowsAffected, nil
}