package module

import (
	"context"
	"encoding/json"
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
	ctxHttpTimeout, cancel := context.WithTimeout(ctx, im.cfg.Inventory.Timeout)
	defer cancel()

	tracer := otel.Tracer("inventory.module")
	ctxSpan, span := tracer.Start(ctxHttpTimeout, "InventoryModule.GetInventory")
	defer span.End()

	logger.Info(ctx, "inventory module GetInventory called")

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
	
	req, err := http.NewRequestWithContext(ctxSpan, method, endpoint, nil)
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

	// Decode the response body into a Product struct
    var res external.InventoryResponse
    if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
        logger.Error(ctx, "Failed to decode response", zap.Error(err))
        return nil, err
    }

    return &res.Product, nil
}
