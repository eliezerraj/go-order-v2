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
	OrderAdd(ctx context.Context, order entity.Order) (*entity.Order, error)
	OrderGet(ctx context.Context, order entity.Order) (*entity.Order, error)
	//CartItemAdd(ctx context.Context, cartItem entity.CartItem) (*entity.CartItem, error)
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

func (p *OrderRepository) OrderAdd(ctx context.Context, order entity.Order) (*entity.Order, error) {
	tracer := otel.Tracer("order.repository")
	ctx, span := tracer.Start(ctx, "OrderRepository.OrderAdd")
	defer span.End()
	
	logger.Info(ctx, "order repository OrderAdd called")

	var err error

	defer func() {
		if err != nil {
			span.RecordError(err) 
			span.SetStatus(codes.Error, err.Error())
			logger.Error(ctx, "order repository OrderAdd failed", zap.Error(err))
		}
	}()

	connectorWriter := p.dbConnector.Writer()

	query := `INSERT INTO public.order (transaction_id,
										order_date,
										fk_cart_id,
										user_id,
										status,
										currency,
										amount,
										address,
										created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`

	rows := connectorWriter.QueryRow(ctx, query, order.Transaction, order.Date, order.CartItem.ID, order.User, order.Status, order.Currency, order.Amount, order.CartItem.Price, order.CreatedAt)

	var id int
	if err := rows.Scan(&id); err != nil {
		return nil, err
	}

	order.ID = id
	return &order, nil
}

func (p *OrderRepository) OrderGet(ctx context.Context, order entity.Order) (*entity.Order, error) {
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

	logger.Info(ctx, "order repository OrderGet called")

	var err error

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
					 o.user_id,
					 o.fk_cart_id,
					 o.created_at,
					 o.updated_at
				from public.order o
				where o.id = $1`

	rows, err := connectorReader.Query(ctx, query, order.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(&order.ID, &order.OrderNumber, &order.Transaction, &order.Date, &order.Status, &order.Currency, &order.Amount, &order.User, &order.CartItem.ID, &order.CreatedAt, &order.UpdatedAt)
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
