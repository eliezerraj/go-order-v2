package repository

import (
	"time"
    "context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/eliezerraj/go-core/v3/database/connector"

	"github.com/go-order-v2/application/tracing"
	"github.com/go-order-v2/application/domain/entity"
)

type DataProviderRepository struct {
    dbConnector connector.IDatabaseConnector
}

type IDataProvider interface {
    TimeSeriesOrderItemsGet(ctx context.Context, product entity.Product, limit int, offset int) (*entity.TimeSeriesOrderItem, error)
}

func NewDataProviderRepository(dbConnector connector.IDatabaseConnector) IDataProvider {
    logger.InfoOutCtx("initializing data provider repository SUCCESSFULLY")

    return &DataProviderRepository{
        dbConnector: dbConnector,
    }
}

func (r *DataProviderRepository) TimeSeriesOrderItemsGet(ctx context.Context, product entity.Product, limit int, offset int) (*entity.TimeSeriesOrderItem, error) {
    logger.Info(ctx, "data provider repository TimeSeriesOrderItemsGet called")

	// Tracing and metrics
	ctx, span := tracing.CustomStartSpanCtx(ctx, "dataProviderRepository.TimeSeriesOrderItemsGet", trace.SpanKindInternal)
	defer span.End()

    meter := otel.Meter("go-order-v2.repository")
    counter, _ := meter.Int64Counter("db_custom_time_series_order_items_get_requests_total")
    histogram, _ := meter.Float64Histogram("db_custom_time_series_order_items_get_duration_seconds")
    start := time.Now()

    counter.Add(ctx, 1, metric.WithAttributes(
        attribute.String("operation", "TimeSeriesOrderItemsGet"),
    ))
	defer func() {
		histogram.Record(ctx, time.Since(start).Seconds(), metric.WithAttributes(
			attribute.String("operation", "TimeSeriesOrderItemsGet"),
		))
	}()

    // Get a reader connection from the database connector
	connectorReader := r.dbConnector.Reader()
    
    query := `select oi.fk_product_id, 
                        created_at::date AS date_part,
                        EXTRACT(HOUR FROM created_at) AS hour_part,
                        count(oi.fk_product_id) as count,
                        sum (oi.quantity) as sum_quantity,
                        sum(oi.amount) as sum_amount,
                        sum(oi.discount) as sum_discount
                from order_item oi
                where oi.fk_product_id = $1
                group by oi.fk_product_id, date_part ,hour_part
                order by date_part ,hour_part asc
                limit $2 offset $3`

    rows, err := connectorReader.Query(ctx, query, product.ID, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var timeSeriesOrderItem *entity.TimeSeriesOrderItem
    for rows.Next() {
        var datePart time.Time
        var hourPart int
        var count int
        var sumQuantity int
        var sumAmount float64
        var sumDiscount float64

        if err := rows.Scan(&product.ID, &datePart, &hourPart, &count, &sumQuantity, &sumAmount, &sumDiscount); err != nil {
            return nil, err
        }

        timeSeriesData := &entity.TimeSeriesData{
            DatePart: datePart,
            HourPart: hourPart,
            Count:    count,
            SumQuantity: sumQuantity,
            SumAmount:   sumAmount,
            SumDiscount: sumDiscount,
        }

        if timeSeriesOrderItem == nil {
            timeSeriesOrderItem = &entity.TimeSeriesOrderItem{
                Product:        product,
                TimeSeriesData: []*entity.TimeSeriesData{timeSeriesData},
            }
        } else {
            timeSeriesOrderItem.TimeSeriesData = append(timeSeriesOrderItem.TimeSeriesData, timeSeriesData)
        }

        timeSeriesOrderItem = timeSeriesOrderItem
    }

    if timeSeriesOrderItem == nil {
        logger.Warn(ctx, "data provider repository TimeSeriesOrderItemsGet returned no results")
        timeSeriesOrderItem = &entity.TimeSeriesOrderItem{
            Product:        product,
            TimeSeriesData: []*entity.TimeSeriesData{},
        }
    }
    return timeSeriesOrderItem, nil
}