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

func NewPaymentModule(cfg *config.Config, client httpclient.IHTTPClient) PaymentModule {
	logger.InfoOutCtx("NewPaymentModule called")

	return PaymentModule{
		cfg: cfg,
		client: client,
	}
}

func (im *PaymentModule) PaymentAdd(ctx context.Context, payment entity.Payment) (*entity.Payment, error) {
	ctxHttpTimeout, cancel := context.WithTimeout(ctx, im.cfg.Inventory.Timeout)
	defer cancel()

	tracer := otel.Tracer("payment.module")
	ctxSpan, span := tracer.Start(ctxHttpTimeout, "PaymentModule.PaymentAdd")
	defer span.End()

	logger.Info(ctx, "payment module PaymentAdd called")

	method := "POST"	
	endpoint := fmt.Sprintf("%s%s", im.cfg.Payment.Endpoint, im.cfg.Payment.UrlPath)

	logger.Info(ctx, "payment module PaymentAdd request", zap.String("method", method), zap.String("endpoint", endpoint))
	
	body, err := json.Marshal(payment)
	if err != nil {
		logger.Error(ctx, "Failed to marshal payment request", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctxSpan, method, endpoint, bytes.NewReader(body))
	if err != nil {
		logger.Error(ctx, "Failed to create request", zap.Error(err))
		return nil, err
	}
	resp, err := im.client.Do(req.WithContext(ctxSpan))
	if err != nil {
		logger.Error(ctx, "Failed to perform request", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	// Check for non-200 status codes
	if resp.StatusCode != http.StatusOK {
		logger.Error(ctx, "Inventory service returned non-OK status", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("inventory service returned status: %d", resp.StatusCode)
	}

	// Decode the response body into a Payment struct
    var res external.PaymentResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logger.Error(ctx, "Failed to decode response", zap.Error(err))
		return nil, err
	}

    return &res.Payment, nil
}
