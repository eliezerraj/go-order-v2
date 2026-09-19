package application

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/eliezerraj/go-core/v3/httpclient"
	"github.com/eliezerraj/go-core/v3/database/connector"
	"github.com/eliezerraj/go-core/v3/auth"

	"github.com/go-order-v2/application/infrastructure/module"
	"github.com/go-order-v2/application/config"
	"github.com/go-order-v2/application/controller"
	"github.com/go-order-v2/application/domain/usecase"
	"github.com/go-order-v2/application/infrastructure/repository"
)

type Application struct {
	OrderController *controller.OrderController
	CheckoutController *controller.CheckoutController
	DataProviderController *controller.DataProviderController
}

type UseCase struct {
	OrderUsecase usecase.IOrderUseCase
	CheckoutUsecase usecase.ICheckoutUseCase
	DataProviderUsecase usecase.IDataProviderUseCase
}

type Repository struct {
	OrderRepository repository.IOrderRepository
	CheckoutRepository repository.ICheckoutRepository
	DataProviderRepository repository.IDataProvider
}

func NewApplication(cfg *config.Config) (*Application, error) {
	logger.InfoOutCtx("initializing application SUCCESSFULLY")

	// Initialize database connector Reader.
	readerConfig := connector.ConnectorConfig{
		DSN:              "postgres://" + cfg.Database.Username + ":" + cfg.Database.Password + "@" + cfg.Database.Host + ":" + cfg.Database.Port + "/" + cfg.Database.Name,
		MaxConnIdleTime:  cfg.Database.ConnIdleTime * time.Minute,
		MaxConnLifeTime:  cfg.Database.ConnLifetime * time.Minute,
		MaxConns:         cfg.Database.MaxConnections,
		MinConns:         cfg.Database.MinConnections,
		DBConnTimeout:    cfg.Database.ConnTimeout,
		HealthCheckPeriod: cfg.Database.ConnIdleTime * time.Minute / 2,
	}

	// Initialize database connector Writer.
	writerConfig := connector.ConnectorConfig{
		DSN:              "postgres://" + cfg.Database.Username + ":" + cfg.Database.Password + "@" + cfg.Database.Host + ":" + cfg.Database.Port + "/" + cfg.Database.Name,
		MaxConnIdleTime:  cfg.Database.ConnIdleTime * time.Minute,
		MaxConnLifeTime:  cfg.Database.ConnLifetime * time.Minute,
		MaxConns:         cfg.Database.MaxConnections,
		MinConns:         cfg.Database.MinConnections,
		DBConnTimeout:    cfg.Database.ConnTimeout,
		HealthCheckPeriod: cfg.Database.ConnIdleTime * time.Minute / 2,
	}

	logger.InfoOutCtx("readerConfig initialized SUCCESSFULLY", zap.Any("readerConfig", readerConfig), zap.Any("writerConfig", writerConfig))

	// Initialize database connector
	dbConnector, err := connector.NewDatabaseConnector(cfg.App.Name, readerConfig, writerConfig)
	if err != nil {
		logger.FatalOutCtx("failed to initialize database connector")
		return nil, err
	}
	
	logger.InfoOutCtx("dbConnector initialized SUCCESSFULLY", zap.Any("dbConnector", dbConnector))
	
	pgConnection := &connector.PgConnection{}
	_, err = pgConnection.NewPool(context.Background(), readerConfig)
	if err != nil {
		logger.FatalOutCtx("failed to create database pool")
		return nil, err
	}
	err = pgConnection.Ping(context.Background())
	if err != nil {
		logger.FatalOutCtx("failed to ping pg connection")
		return nil, err
	}

	// Repository initialization
	orderRepository := repository.NewOrderRepository(dbConnector)
	checkoutRepository := repository.NewCheckoutRepository(dbConnector)
	dataProviderRepository := repository.NewDataProviderRepository(dbConnector)
	// Create the forwards modules.
	httpConfig := &httpclient.HttpConfig{
		Timeout:             cfg.HTTP.Timeout * time.Second,
		KeepAlive:           cfg.HTTP.KeepAlive * time.Second,
		IdleConnTimeout:     cfg.HTTP.IdleConnTimeout * time.Second,
		MaxIdleConns:        cfg.HTTP.MaxIdleConns,
		MaxIdleConnsPerHost: cfg.HTTP.MaxIdleConnsPerHost,
		MaxConnsPerHost:     cfg.HTTP.MaxConnsPerHost,
		ServiceName:         cfg.App.Name,
	}

	// Create HTTP clients for the modules.
	invHttpClient := httpclient.NewHttpClient(httpConfig)
	payHttpClient := httpclient.NewHttpClient(httpConfig)

	// Create the authentication module.
	authClientServiceOption := func(a *auth.AuthClientService) {
		a.AuthURL = cfg.Authentication.AuthURL
		a.RefreshURL = cfg.Authentication.RefreshURL
		a.ClientID = cfg.Authentication.ClientID
		a.ClientSecret = cfg.Authentication.ClientSecret
		a.DryRun = cfg.Authentication.DryRun
		a.RefreshInterval = cfg.Authentication.RefreshInterval
	}
	
	authClientService := auth.NewAuthClientService(authClientServiceOption)
	authClientService.StartAuthenticate(context.Background())

	// Create the inventory module.
	inventoryModule := module.NewInventoryModule(cfg, invHttpClient, authClientService)
	paymentModule := module.NewPaymentModule(cfg, payHttpClient, authClientService)

	// UseCase initialization
	orderUsecase := usecase.NewOrderUseCase(orderRepository, inventoryModule, paymentModule)
	checkoutUsecase := usecase.NewCheckoutUseCase(orderRepository, checkoutRepository, paymentModule, inventoryModule)
	
	// UseCase for DataProvider initialization
	dataProviderUsecase := usecase.NewDataProviderUseCase(dataProviderRepository, inventoryModule)

	// Controller initialization
	orderController := controller.NewOrderController(orderUsecase)
	checkoutController := controller.NewCheckoutController(checkoutUsecase)
	dataProviderController := controller.NewDataProviderController(dataProviderUsecase)

	return &Application{
		OrderController: orderController,
		CheckoutController: checkoutController,
		DataProviderController: dataProviderController,
	}, nil
}
