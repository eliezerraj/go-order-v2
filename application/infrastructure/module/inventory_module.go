package module

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	//"errors"
	"go.uber.org/zap"

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

func (im *InventoryModule) GetInventory(ctx context.Context, productID string) (*entity.Product, error) {
	logger.Info(ctx, "inventory module GetInventory called")

	method := "GET"
	endpoint := fmt.Sprintf("%s%s/%s", im.cfg.Inventory.Endpoint, im.cfg.Inventory.UrlPath, productID)

	logger.Info(ctx, "inventory module GetInventory request", zap.String("method", method), zap.String("endpoint", endpoint))
	
	ctxHttpTimeout, cancel := context.WithTimeout(ctx, im.cfg.Inventory.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctxHttpTimeout, method, endpoint, nil)
	if err != nil {
		logger.Error(ctx, "Failed to create request", zap.Error(err))
		return nil, err
	}
	resp, err := im.client.Do(req)
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
