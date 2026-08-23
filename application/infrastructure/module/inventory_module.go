package module

import (
	"bytes"
	"context"
	"encoding/json"
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

const RequestIDHeaderName = "x-request-id"

type InventoryModule struct {
	cfg *config.Config
	client	httpclient.IHTTPClient
}

func NewInventoryModule(cfg *config.Config, client	httpclient.IHTTPClient) InventoryModule {
	logger.InfoOutCtx("NewInventoryModule called")

	return InventoryModule{
		cfg: cfg,
		client: client,
	}
}

func (im *InventoryModule) GetInventory(ctx context.Context, product entity.Product) (*entity.Product, error) {
	logger.Info(ctx, "inventory module GetInventory called")

	ctxHttpTimeout, cancel := context.WithTimeout(ctx, im.cfg.Inventory.Timeout)
	defer cancel()

	// Trace
	ctx, span := tracing.CustomStartSpanCtx(ctxHttpTimeout, "inventoryModule.GetInventory", trace.SpanKindInternal)
	defer span.End()

	var endpoint string
	method := "GET"
	
	if product.Sku != "" {
		endpoint = fmt.Sprintf("%s%s/%s", im.cfg.Inventory.Endpoint, im.cfg.Inventory.UrlPath, product.Sku)
	} else if product.ID != 0 {
		endpoint = fmt.Sprintf("%s%s/%d", im.cfg.Inventory.Endpoint, im.cfg.Inventory.UrlPath, product.ID)
	} else {
		err := fmt.Errorf("product must have either SKU or ID")
		logger.Error(ctx, "inventory module GetInventory failed", zap.Error(err))
		return nil, err
	}

	logger.Info(ctx, "inventory module GetInventory request", zap.String("method", method), zap.String("endpoint", endpoint))
	
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
	case http.StatusNotFound:
		logger.Error(ctx, "Inventory service returned 404 Not Found")
		return nil, fmt.Errorf("inventory service returned status: %d", resp.StatusCode)
	default:
		logger.Error(ctx, "Inventory service returned unexpected status", zap.Int("status", resp.StatusCode))
		return nil, fmt.Errorf("inventory service returned status: %d", resp.StatusCode)
	}

	// Decode the response body into a Product struct
    var res external.InventoryResponse
    if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
        logger.Error(ctx, "Failed to decode response", zap.Error(err))
        return nil, err
    }

    return &res.Product, nil
}

func (im *InventoryModule) InventoryPatch(ctx context.Context, product entity.Product) error {
	logger.Info(ctx, "inventory module InventoryPatch called")

	ctxHttpTimeout, cancel := context.WithTimeout(ctx, im.cfg.Inventory.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxHttpTimeout, "inventoryModule.InventoryPatch", trace.SpanKindInternal)
	defer span.End()

	endpoint := fmt.Sprintf("%s%s/inventory/%d", im.cfg.Inventory.Endpoint, im.cfg.Inventory.UrlPath, product.ID)
	method := "PATCH"

	logger.Info(ctx, "inventory module InventoryPatch request", zap.String("method", method), zap.String("endpoint", endpoint))

	inventory := &entity.Inventory{
		Available: product.Inventory.Available,
		Sold:      product.Inventory.Sold,
		Pending:   product.Inventory.Pending,
	}
	product.Inventory = inventory
	payload := product

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		logger.Error(ctx, "Failed to marshal payload", zap.Error(err))
		return err
	}

	logger.Debug(ctx, "inventory module InventoryPatch payload", zap.Any("payload", payload))

	req, err := http.NewRequestWithContext(ctx, method, endpoint, bytes.NewReader(payloadBytes))
	if err != nil {
		logger.Error(ctx, "Failed to create request", zap.Error(err))
		return err
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
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Error(ctx, "Inventory service returned unexpected status", zap.Int("status", resp.StatusCode))
		return fmt.Errorf("inventory service returned status: %d", resp.StatusCode)
	}

	return nil
}