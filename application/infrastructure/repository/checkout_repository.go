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

type CheckoutRepository struct {
	dbConnector connector.IDatabaseConnector
}

type ICheckoutRepository interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	CheckoutGet(ctx context.Context, checkout entity.Checkout) (*entity.Checkout, error)
}

func NewCheckoutRepository(dbConnector connector.IDatabaseConnector) ICheckoutRepository {
	logger.InfoOutCtx("initializing checkout repository SUCCESSFULLY")

	return &CheckoutRepository{
		dbConnector: dbConnector,
	}
}

func (p *CheckoutRepository) BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	logger.InfoOutCtx("checkout repository BeginTx called")

	tx, err := p.dbConnector.Writer().BeginTx(ctx, opts)
	if err != nil {
		logger.ErrorOutCtx("checkout repository BeginTx failed", zap.Error(err))
		return nil, err
	}

	return tx, nil
}

func (p *CheckoutRepository) CheckoutGet(ctx context.Context, checkout entity.Checkout) (res_checkout *entity.Checkout, err error) {
	logger.Info(ctx, "checkout repository CheckoutGet called")

	// Tracing and metrics
	tracer := otel.Tracer("order.repository")
    ctx, span := tracer.Start(ctx, "CheckoutRepository.CheckoutGet")
    defer span.End()

    meter := otel.Meter("go-order-v2.repository")
    counter, _ := meter.Int64Counter("db_custom_checkout_get_requests_total")
    histogram, _ := meter.Float64Histogram("db_custom_checkout_get_duration_seconds")
    start := time.Now()

    counter.Add(ctx, 1, metric.WithAttributes(
        attribute.String("operation", "CheckoutGet"),
    ))

	// Defer function to handle error logging and metrics recording	
	defer func() {
		if err != nil {
			span.RecordError(err) 
			span.SetStatus(codes.Error, err.Error())
			logger.Error(ctx, "checkout repository CheckoutGet failed", zap.Error(err))
		}
        histogram.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(
            attribute.String("operation", "CheckoutGet"),
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

	rows, err := connectorReader.Query(ctx, query, checkout.Order.OrderNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checkout.Order.OrderItem = &[]entity.OrderItem{}
	if rows.Next() {
		err = rows.Scan(&checkout.Order.ID, &checkout.Order.OrderNumber, &checkout.Order.Transaction, &checkout.Order.Date, &checkout.Order.Status, &checkout.Order.Currency, &checkout.Order.Amount, &checkout.Order.CustomerID, &checkout.Order.CreatedAt, &checkout.Order.UpdatedAt)
		if err != nil {
			return nil, err
		}
	} else {
		logger.Warn(ctx, "not found", zap.String("order_number", checkout.Order.OrderNumber))
		err = errors.New("order not found")
		return nil, err
	}

	return &checkout, nil
}
