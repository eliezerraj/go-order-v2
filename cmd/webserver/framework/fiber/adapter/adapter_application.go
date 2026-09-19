package adapter

import (
	"context"
	"strconv"
	"go.uber.org/zap"

	"github.com/go-order-v2/application/config"
	"github.com/go-order-v2/application/infrastructure/application"
	"github.com/go-order-v2/application/domain/external"
	"github.com/go-order-v2/application/tracing"

	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/eliezerraj/go-core/v3/http/utils"

	"github.com/gofiber/fiber/v2"

	"go.opentelemetry.io/otel/trace"
)

type ApplicationAdapter struct {
	cfg *config.Config
	application *application.Application
}

func NewApplicationAdapter(cfg *config.Config, application *application.Application) *ApplicationAdapter {
	logger.InfoOutCtx("initializing application adapter SUCCESSFULLY")

	return &ApplicationAdapter{
		cfg:         cfg,
		application: application,
	}
}

// Adapter methods for OrderController
func (a *ApplicationAdapter) OrderGet(ctxFiber *fiber.Ctx) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.UserContext(), a.cfg.HTTP.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxWithTimeout, "applicationAdapter.orderGet", trace.SpanKindInternal)
	defer span.End()

	logger.Info(ctx, "OrderGet called")

	logger.Debug(
		ctx,
		a.cfg.App.Name,
		zap.ByteString("headers", utils.FormatHeadersAsJSON(ctxFiber.GetReqHeaders())),
		zap.ByteString("query", ctxFiber.Request().URI().QueryString()),
		zap.ByteString("body", ctxFiber.Body()),
	)

	order_number := ctxFiber.Params("order_number")
	if order_number == "" {
		order_number = ctxFiber.Query("order_number")
	}

	order := external.OrderRequest{
		OrderNumber: order_number,
	}

	res, err := a.application.OrderController.OrderGet(ctx, order)
	if err != nil {
		logger.Error(ctx, "failed to get order ", zap.Error(err))
		errorResponse := external.NewResponseError(ctx,
			fiber.StatusNotFound,
			fiber.ErrNotFound,
			fiber.ErrNotFound.Message,
			"failed to get order",
			err.Error(),
			external.BUSSINESS_ERROR)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	resp := external.OrderResponse{
		Response: "Order retrieved successfully",
		Order: res,
	}

	return ctxFiber.Status(fiber.StatusOK).JSON(resp)
}

func (a *ApplicationAdapter) OrderAdd(ctxFiber *fiber.Ctx) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.UserContext(), a.cfg.HTTP.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxWithTimeout, "applicationAdapter.orderAdd", trace.SpanKindInternal)
	defer span.End()

	logger.Info(ctx, "OrderAdd called")

	logger.Debug(
		ctx,
		a.cfg.App.Name,
		zap.ByteString("headers", utils.FormatHeadersAsJSON(ctxFiber.GetReqHeaders())),
		zap.ByteString("query", ctxFiber.Request().URI().QueryString()),
		zap.ByteString("body", ctxFiber.Body()),
	)
	order := external.OrderRequest{}
	if err := ctxFiber.BodyParser(&order); err != nil {
		logger.Error(ctx, "failed to parse request body", zap.Error(err))
		errorResponse := external.NewResponseError(ctx,
			fiber.StatusBadRequest,
			fiber.ErrBadRequest,
			fiber.ErrBadRequest.Message,
			"failed to parse request body",
			err.Error(),
			external.BUSSINESS_ERROR)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	res, err := a.application.OrderController.OrderAdd(ctx, order)
	if err != nil {
		logger.Error(ctx, "failed to add order", zap.Error(err))
		errorResponse := external.NewResponseError(ctx,
			fiber.StatusInternalServerError,
			fiber.ErrInternalServerError,
			fiber.ErrInternalServerError.Message,
			"failed to add order",
			err.Error(),
			external.BUSSINESS_ERROR)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	resp := external.OrderResponse{
		Response: "Order added successfully",
		Order: res,
	}

	return ctxFiber.Status(fiber.StatusCreated).JSON(resp)
}

// -----------------------------------
// Controller methods for CheckoutController
// -----------------------------------

func (a *ApplicationAdapter) CheckoutGet(ctxFiber *fiber.Ctx) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.UserContext(), a.cfg.HTTP.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxWithTimeout, "applicationAdapter.checkoutGet", trace.SpanKindInternal)
	defer span.End()

	logger.Info(ctx, "CheckoutGet called")

	logger.Debug(
		ctx,
		a.cfg.App.Name,
		zap.ByteString("headers", utils.FormatHeadersAsJSON(ctxFiber.GetReqHeaders())),
		zap.ByteString("query", ctxFiber.Request().URI().QueryString()),
		zap.ByteString("body", ctxFiber.Body()),
	)

	order_number := ctxFiber.Params("order_number")
	if order_number == "" {
		order_number = ctxFiber.Query("order_number")
	}

	checkout := external.CheckoutRequest{
		Order: &external.OrderRequest{
			OrderNumber: order_number,
		},
	}

	res, err := a.application.CheckoutController.CheckoutGet(ctx, checkout)
	if err != nil {
		logger.Error(ctx, "failed to get checkout ", zap.Error(err))
		errorResponse := external.NewResponseError(ctx,
			fiber.StatusNotFound,
			fiber.ErrNotFound,
			fiber.ErrNotFound.Message,
			"failed to get checkout",
			err.Error(),
			external.BUSSINESS_ERROR)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	resp := external.CheckoutResponse{
		Response: "Checkout retrieved successfully",
		Checkout: res,
	}

	return ctxFiber.Status(fiber.StatusOK).JSON(resp)
}

// -----------------------------------
// Controller methods for CheckoutController
// -----------------------------------
func (a *ApplicationAdapter) CheckoutAdd(ctxFiber *fiber.Ctx) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.UserContext(), a.cfg.HTTP.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxWithTimeout, "applicationAdapter.checkoutAdd", trace.SpanKindInternal)
	defer span.End()

	logger.Info(ctx, "CheckoutAdd called")

	logger.Debug(
		ctx,
		a.cfg.App.Name,
		zap.ByteString("headers", utils.FormatHeadersAsJSON(ctxFiber.GetReqHeaders())),
		zap.ByteString("query", ctxFiber.Request().URI().QueryString()),
		zap.ByteString("body", ctxFiber.Body()),
	)

	checkout := external.CheckoutRequest{}
	if err := ctxFiber.BodyParser(&checkout); err != nil {
		logger.Error(ctx, "failed to parse request body", zap.Error(err))
		errorResponse := external.NewResponseError(ctx,
			fiber.StatusBadRequest,
			fiber.ErrBadRequest,
			fiber.ErrBadRequest.Message,
			"failed to parse request body",
			err.Error(),
			external.BUSSINESS_ERROR)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	res, err := a.application.CheckoutController.CheckoutAdd(ctx, checkout)
	if err != nil {
		logger.Error(ctx, "failed to add checkout", zap.Error(err))
		errorResponse := external.NewResponseError(ctx,
			fiber.StatusInternalServerError,
			fiber.ErrInternalServerError,
			fiber.ErrInternalServerError.Message,
			"failed to add checkout",
			err.Error(),
			external.BUSSINESS_ERROR)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	resp := external.CheckoutResponse{
		Response: "Checkout added successfully",
		Checkout: res,
	}

	return ctxFiber.Status(fiber.StatusCreated).JSON(resp)
}

// -----------------------------------
// Controller methods for DataProviderController
// -----------------------------------
func (a *ApplicationAdapter) TimeSeriesOrderItemsGet(ctxFiber *fiber.Ctx) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.UserContext(), a.cfg.HTTP.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxWithTimeout, "applicationAdapter.timeSeriesOrderItemsGet", trace.SpanKindInternal)
	defer span.End()

	logger.Info(ctx, "TimeSeriesOrderItemsGet called")

	logger.Debug(
		ctx,
		a.cfg.App.Name,
		zap.ByteString("headers", utils.FormatHeadersAsJSON(ctxFiber.GetReqHeaders())),
		zap.ByteString("query", ctxFiber.Request().URI().QueryString()),
		zap.Any("params", ctxFiber.AllParams()),
		zap.ByteString("body", ctxFiber.Body()),
	)

	product_sku := ctxFiber.Query("product")
	limit, err := strconv.Atoi(ctxFiber.Query("limit"))
	if err != nil {
		logger.Warn(ctx, "failed to parse limit, using default value 7", zap.Error(err))
		limit = 7
	}
	offset, err := strconv.Atoi(ctxFiber.Query("offset"))
	if err != nil {
		logger.Warn(ctx, "failed to parse offset, using default value 0", zap.Error(err))
		offset = 0
	}

	timeSeriesOrderItemRequest := external.TimeSeriesOrderItemRequest{
		ProductRequest: external.ProductRequest{
			Sku: product_sku,
		},
		Limit:  limit,
		Offset: offset,
	}

	res, err := a.application.DataProviderController.TimeSeriesOrderItemsGet(ctx, timeSeriesOrderItemRequest)
	if err != nil {
		logger.Error(ctx, "failed to get time series order items", zap.Error(err))
		errorResponse := external.NewResponseError(ctx,
			fiber.StatusInternalServerError,
			fiber.ErrInternalServerError,
			fiber.ErrInternalServerError.Message,
			"failed to get time series order items",
			err.Error(),
			external.BUSSINESS_ERROR)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	resp := external.TimeSeriesOrderItemResponse{
		Response: "Time series order items retrieved successfully",
		TimeSeriesOrderItems: res,
	}

	return ctxFiber.Status(fiber.StatusOK).JSON(resp)
}