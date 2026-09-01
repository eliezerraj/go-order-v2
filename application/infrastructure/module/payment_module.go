package module

import (
	"bytes"
	"encoding/json"
	"context"
	"fmt"
	"net/http"
	"go.uber.org/zap"

	"go.opentelemetry.io/otel/trace"

	"github.com/eliezerraj/go-core/v3/httpclient"
	"github.com/eliezerraj/go-core/v3/logger"

	"github.com/go-order-v2/application/config"
	"github.com/go-order-v2/application/domain/entity"
	"github.com/go-order-v2/application/domain/external"
	"github.com/go-order-v2/application/tracing"
)

type PaymentModule struct {
	cfg *config.Config
	client	httpclient.IHTTPClient
}

const (
	AcceptHeader      = "Accept"
	ContentTypeHeader = "Content-Type"
	ConnectionHeader  = "Connection"
	KeepAlive         = "keep-alive"
	XResquestID		 = "X-Request-ID"
)

func NewPaymentModule(cfg *config.Config, client httpclient.IHTTPClient) PaymentModule {
	logger.InfoOutCtx("NewPaymentModule called")

	return PaymentModule{
		cfg: cfg,
		client: client,
	}
}

func (im *PaymentModule) PaymentAdd(ctx context.Context, paymentRequest external.PaymentRequest) (*entity.PaymentCheckout, error) {
	logger.Info(ctx, "payment module PaymentAdd called")

	ctxHttpTimeout, cancel := context.WithTimeout(ctx, im.cfg.Payment.Timeout)
	defer cancel()

	// Trace
	ctx, span := tracing.CustomStartSpanCtx(ctxHttpTimeout, "paymentModule.PaymentAdd", trace.SpanKindInternal)
	defer span.End()

	method := "POST"	
	endpoint := fmt.Sprintf("%s%s", im.cfg.Payment.Endpoint, im.cfg.Payment.UrlPath)

	logger.Debug(ctx, "payment module PaymentAdd request", zap.String("method", method), zap.String("endpoint", endpoint))
	logger.Debug(ctx, "payment module PaymentAdd request body", zap.Any("payment_request", paymentRequest))
	
	body, err := json.Marshal(paymentRequest)
	if err != nil {
		logger.Error(ctx, "Failed to marshal payment request", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(body))
	if err != nil {
		logger.Error(ctx, "Failed to create request", zap.Error(err))
		return nil, err
	}

	// Set headers for the request. the const are in payment_module.go file
	xrequestid, ok := ctx.Value(RequestIDHeaderName).(string)
	if !ok {
		xrequestid = "not-informed"
	}

	headers := map[string]string{
		ConnectionHeader:  KeepAlive,
		AcceptHeader:      "application/json",
		ContentTypeHeader: "application/json",
		KeepAlive: "timeout=5, max=1000",
		XResquestID: xrequestid,
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := im.client.Do(req.WithContext(ctx))
	if err != nil {
		logger.Error(ctx, "Failed to perform request", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	switch resp.StatusCode {
	case http.StatusOK:
		// Continue processing
	case http.StatusCreated:
		// Continue processing
	case http.StatusNotFound:
		logger.Error(ctx, "Payment service returned 404 Not Found")
		return nil, fmt.Errorf("payment service returned status: %d", resp.StatusCode)
	default:
		logger.Error(ctx, "Payment service returned unexpected status", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("payment service returned status: %d", resp.StatusCode)
	}

	// Decode the response body into a Payment struct
    var res external.PaymentResponse[entity.PaymentCheckout]
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logger.Error(ctx, "Failed to decode response", zap.Error(err))
		return nil, err
	}

    return &res.Payment, nil
}

func (im *PaymentModule) PaymentGet(ctx context.Context, paymentRequest external.PaymentRequest) ([]*entity.Payment, error) {
	logger.Info(ctx, "payment module PaymentGet called")

	ctxHttpTimeout, cancel := context.WithTimeout(ctx, im.cfg.Payment.Timeout)
	defer cancel()

	// Trace
	ctx, span := tracing.CustomStartSpanCtx(ctxHttpTimeout, "paymentModule.PaymentGet", trace.SpanKindInternal)
	defer span.End()

	method := "GET"
	endpoint := fmt.Sprintf("%s%s/%v", im.cfg.Payment.Endpoint, im.cfg.Payment.UrlPath + "/order", paymentRequest.Order.ID)

	req, err := http.NewRequestWithContext(ctx, method, endpoint, nil)
	if err != nil {
		logger.Error(ctx, "Failed to create request", zap.Error(err))
		return nil, err
	}

	// Set headers for the request. the const are in payment_module.go file
	xrequestid, ok := ctx.Value(RequestIDHeaderName).(string)
	if !ok {
		xrequestid = "not-informed"
	}

	headers := map[string]string{
		ConnectionHeader:  KeepAlive,
		AcceptHeader:      "application/json",
		ContentTypeHeader: "application/json",
		KeepAlive: "timeout=5, max=1000",
		XResquestID: xrequestid,
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := im.client.Do(req.WithContext(ctx))
	if err != nil {
		logger.Error(ctx, "Failed to perform request", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	switch resp.StatusCode {
	case http.StatusOK:
		// Continue processing
	case http.StatusCreated:
		// Continue processing
	case http.StatusNotFound:
		logger.Error(ctx, "Payment service returned 404 Not Found")
		return nil, fmt.Errorf("payment service returned status: %d", resp.StatusCode)
	default:
		logger.Error(ctx, "Payment service returned unexpected status", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("payment service returned status: %d", resp.StatusCode)
	}

	var res external.PaymentResponse[[]*entity.Payment]
	// Decode the response body into a Payment struct
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logger.Error(ctx, "Failed to decode response", zap.Error(err))
		return nil, err
	}

    return res.Payment, nil
}