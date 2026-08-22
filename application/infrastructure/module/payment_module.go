package module

import (
	"bytes"
	"encoding/json"
	"context"
	"fmt"
	"net/http"
	"go.uber.org/zap"

	"go.opentelemetry.io/otel"

	"github.com/eliezerraj/go-core/v3/httpclient"
	"github.com/eliezerraj/go-core/v3/logger"

	"github.com/go-order-v2/application/config"
	"github.com/go-order-v2/application/domain/entity"
	"github.com/go-order-v2/application/domain/external"
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
)

func NewPaymentModule(cfg *config.Config, client httpclient.IHTTPClient) PaymentModule {
	logger.InfoOutCtx("NewPaymentModule called")

	return PaymentModule{
		cfg: cfg,
		client: client,
	}
}

func (im *PaymentModule) PaymentAdd(ctx context.Context, paymentRequest external.PaymentRequest) (*entity.Payment, error) {
	ctxHttpTimeout, cancel := context.WithTimeout(ctx, im.cfg.Payment.Timeout)
	defer cancel()

	tracer := otel.Tracer("payment.module")
	ctxSpan, span := tracer.Start(ctxHttpTimeout, "PaymentModule.PaymentAdd")
	defer span.End()

	logger.Info(ctxSpan, "payment module PaymentAdd called")

	method := "POST"	
	endpoint := fmt.Sprintf("%s%s", im.cfg.Payment.Endpoint, im.cfg.Payment.UrlPath)

	logger.Debug(ctxSpan, "payment module PaymentAdd request", zap.String("method", method), zap.String("endpoint", endpoint))
	logger.Debug(ctxSpan, "payment module PaymentAdd request body", zap.Any("payment_request", paymentRequest))
	
	body, err := json.Marshal(paymentRequest)
	if err != nil {
		logger.Error(ctxSpan, "Failed to marshal payment request", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctxSpan, method, endpoint, bytes.NewReader(body))
	if err != nil {
		logger.Error(ctxSpan, "Failed to create request", zap.Error(err))
		return nil, err
	}

	headers := map[string]string{
		ConnectionHeader:  KeepAlive,
		AcceptHeader:      "application/json",
		ContentTypeHeader: "application/json",
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp, err := im.client.Do(req.WithContext(ctxSpan))
	if err != nil {
		logger.Error(ctxSpan, "Failed to perform request", zap.Error(err))
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
		logger.Error(ctx, "Inventory service returned 404 Not Found")
		return nil, fmt.Errorf("inventory service returned status: %d", resp.StatusCode)
	default:
		logger.Error(ctx, "Inventory service returned unexpected status", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("inventory service returned status: %d", resp.StatusCode)
	}

	// Decode the response body into a Payment struct
    var res external.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logger.Error(ctxSpan, "Failed to decode response", zap.Error(err))
		return nil, err
	}

    return &res.Payment, nil
}
