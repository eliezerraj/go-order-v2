package adapter

import (
	"context"
	"strconv"
	"errors"

	"go.uber.org/zap"

	"github.com/go-order-v2/application/domain/apperrors"
	"github.com/go-order-v2/application/config"
	"github.com/go-order-v2/application/infrastructure/application"
	"github.com/go-order-v2/application/domain/external"
	"github.com/go-order-v2/application/tracing"

	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/eliezerraj/go-core/v3/http/utils"

	"github.com/gofiber/fiber/v3"

	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/attribute"
)

type ApplicationAdapter struct {
	cfg *config.Config
	application *application.Application
}

func NewApplicationAdapter(cfg *config.Config, application *application.Application) *ApplicationAdapter {
	logger.Info(context.Background(), "initializing application adapter SUCCESSFULLY")

	return &ApplicationAdapter{
		cfg:         cfg,
		application: application,
	}
}

// Adapter methods for OrderController
func (a *ApplicationAdapter) OrderGet(ctxFiber fiber.Ctx) error {
	logger.Info(ctxFiber.Context(), "OrderGet called")

	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.Context(), a.cfg.HTTP.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxWithTimeout, "applicationAdapter.orderGet", trace.SpanKindInternal)
	defer span.End()

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
		var statusCode int
		var statusCodeMessage string

		if errors.Is(err, apperrors.ErrNotFound) {
			logger.Warn(ctx, "order not found", zap.Error(err))
			statusCode = fiber.StatusNotFound
			statusCodeMessage = fiber.ErrNotFound.Message
		} else {
			logger.Error(ctx, "failed to get order ", zap.Error(err))
			statusCode = fiber.StatusInternalServerError
			statusCodeMessage = fiber.ErrInternalServerError.Message
		}

		errorResponse := external.NewResponseError(ctx,
			statusCode,
			fiber.NewError(statusCode),
			statusCodeMessage,
			"failed to get order",
			err.Error(),
			external.BUSSINESS_ERROR)
		span.SetAttributes(
				attribute.Int("http.response.status_code", statusCode),
		)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	resp := external.OrderResponse{
		Response: "Order retrieved successfully",
		Order: res,
	}
	span.SetAttributes(
		attribute.Int("http.response.status_code", fiber.StatusOK),
	)
	return ctxFiber.Status(fiber.StatusOK).JSON(resp)
}

func (a *ApplicationAdapter) OrderAdd(ctxFiber fiber.Ctx) error {
	logger.Info(ctxFiber.Context(), "OrderAdd called")

	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.Context(), a.cfg.HTTP.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxWithTimeout, "applicationAdapter.orderAdd", trace.SpanKindInternal)
	defer span.End()

	logger.Debug(
		ctx,
		a.cfg.App.Name,
		zap.ByteString("headers", utils.FormatHeadersAsJSON(ctxFiber.GetReqHeaders())),
		zap.ByteString("query", ctxFiber.Request().URI().QueryString()),
		zap.ByteString("body", ctxFiber.Body()),
	)
	order := external.OrderRequest{}
	if err := ctxFiber.Bind().Body(&order); err != nil {
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
		
		var statusCode int
		var statusCodeMessage string

		if errors.Is(err, apperrors.ErrNotFound) {
			logger.Warn(ctx, "order not found", zap.Error(err))
			statusCode = fiber.StatusNotFound
			statusCodeMessage = fiber.ErrNotFound.Message
		} else {
			logger.Error(ctx, "failed to add order ", zap.Error(err))
			statusCode = fiber.StatusInternalServerError
			statusCodeMessage = fiber.ErrInternalServerError.Message
		}

		errorResponse := external.NewResponseError(ctx,
			statusCode,
			fiber.NewError(statusCode),
			statusCodeMessage,
			"failed to add order",
			err.Error(),
			external.BUSSINESS_ERROR)
		span.SetAttributes(
				attribute.Int("http.response.status_code", statusCode),
		)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	resp := external.OrderResponse{
		Response: "Order added successfully",
		Order: res,
	}

	span.SetAttributes(
		attribute.Int("http.response.status_code", fiber.StatusCreated),
	)
	return ctxFiber.Status(fiber.StatusCreated).JSON(resp)
}

// -----------------------------------
// Controller methods for CheckoutController
// -----------------------------------

func (a *ApplicationAdapter) CheckoutGet(ctxFiber fiber.Ctx) error {
	logger.Info(ctxFiber.Context(), "CheckoutGet called")

	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.Context(), a.cfg.HTTP.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxWithTimeout, "applicationAdapter.checkoutGet", trace.SpanKindInternal)
	defer span.End()

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

		var statusCode int
		var statusCodeMessage string

		if errors.Is(err, apperrors.ErrNotFound) {
			logger.Warn(ctx, "checkout not found", zap.Error(err))
			statusCode = fiber.StatusNotFound
			statusCodeMessage = fiber.ErrNotFound.Message
		} else {
			logger.Error(ctx, "failed to get checkout ", zap.Error(err))
			statusCode = fiber.StatusInternalServerError
			statusCodeMessage = fiber.ErrInternalServerError.Message
		}
		errorResponse := external.NewResponseError(ctx,
			statusCode,
			fiber.NewError(statusCode),
			statusCodeMessage,
			"failed to get checkout",
			err.Error(),
			external.BUSSINESS_ERROR)
		span.SetAttributes(
				attribute.Int("http.response.status_code", statusCode),
		)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	resp := external.CheckoutResponse{
		Response: "Checkout retrieved successfully",
		Checkout: res,
	}
	span.SetAttributes(
		attribute.Int("http.response.status_code", fiber.StatusOK),
	)
	return ctxFiber.Status(fiber.StatusOK).JSON(resp)
}

// -----------------------------------
// Controller methods for CheckoutController
// -----------------------------------
func (a *ApplicationAdapter) CheckoutAdd(ctxFiber fiber.Ctx) error {
	logger.Info(ctxFiber.Context(), "CheckoutAdd called")

	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.Context(), a.cfg.HTTP.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxWithTimeout, "applicationAdapter.checkoutAdd", trace.SpanKindInternal)
	defer span.End()

	logger.Debug(
		ctx,
		a.cfg.App.Name,
		zap.ByteString("headers", utils.FormatHeadersAsJSON(ctxFiber.GetReqHeaders())),
		zap.ByteString("query", ctxFiber.Request().URI().QueryString()),
		zap.ByteString("body", ctxFiber.Body()),
	)

	checkout := external.CheckoutRequest{}
	if err := ctxFiber.Bind().Body(&checkout); err != nil {
		logger.Error(ctx, "failed to parse request body", zap.Error(err))
		errorResponse := external.NewResponseError(ctx,
			fiber.StatusBadRequest,
			fiber.ErrBadRequest,
			fiber.ErrBadRequest.Message,
			"failed to parse request body",
			err.Error(),
			external.BUSSINESS_ERROR)
		span.SetAttributes(
				attribute.Int("http.response.status_code", fiber.StatusBadRequest),
		)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	res, err := a.application.CheckoutController.CheckoutAdd(ctx, checkout)
	if err != nil {

		var statusCode int
		var statusCodeMessage string

		if errors.Is(err, apperrors.ErrNotFound) {
			logger.Warn(ctx, "checkout not found", zap.Error(err))
			statusCode = fiber.StatusNotFound
			statusCodeMessage = fiber.ErrNotFound.Message
		} else {
			logger.Error(ctx, "failed to add checkout ", zap.Error(err))
			statusCode = fiber.StatusInternalServerError
			statusCodeMessage = fiber.ErrInternalServerError.Message
		}
		errorResponse := external.NewResponseError(ctx,
			statusCode,
			fiber.NewError(statusCode),
			statusCodeMessage,
			"failed to add checkout",
			err.Error(),
			external.BUSSINESS_ERROR)
		span.SetAttributes(
				attribute.Int("http.response.status_code", statusCode),
		)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	resp := external.CheckoutResponse{
		Response: "Checkout added successfully",
		Checkout: res,
	}
	span.SetAttributes(
		attribute.Int("http.response.status_code", fiber.StatusCreated),
	)
	return ctxFiber.Status(fiber.StatusCreated).JSON(resp)
}

// -----------------------------------
// Controller methods for DataProviderController
// -----------------------------------
func (a *ApplicationAdapter) TimeSeriesOrderItemsGet(ctxFiber fiber.Ctx) error {
	logger.Info(ctxFiber.Context(), "TimeSeriesOrderItemsGet called")

	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.Context(), a.cfg.HTTP.Timeout)
	defer cancel()

	ctx, span := tracing.CustomStartSpanCtx(ctxWithTimeout, "applicationAdapter.timeSeriesOrderItemsGet", trace.SpanKindInternal)
	defer span.End()

	logger.Debug(
		ctx,
		a.cfg.App.Name,
		zap.ByteString("headers", utils.FormatHeadersAsJSON(ctxFiber.GetReqHeaders())),
		zap.ByteString("query", ctxFiber.Request().URI().QueryString()),
		//zap.Any("params", ctxFiber.AllParams()),
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

		var statusCode int
		var statusCodeMessage string
		
		if errors.Is(err, apperrors.ErrNotFound) {
			statusCode = fiber.StatusNotFound
			statusCodeMessage = fiber.ErrNotFound.Message
			logger.Warn(ctx, "time series order items not found", zap.Error(err))
		} else {
			statusCode = fiber.StatusInternalServerError
			statusCodeMessage = fiber.ErrInternalServerError.Message
			logger.Error(ctx, "failed to get time series order items", zap.Error(err))
		}

		errorResponse := external.NewResponseError(ctx,
			statusCode,
			fiber.NewError(statusCode),
			statusCodeMessage,
			"failed to get time series order items",
			err.Error(),
			external.BUSSINESS_ERROR)
		span.SetAttributes(
				attribute.Int("http.response.status_code", statusCode),
		)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	resp := external.TimeSeriesOrderItemResponse{
		Response: "Time series order items retrieved successfully",
		TimeSeriesOrderItems: res,
	}
	span.SetAttributes(
		attribute.Int("http.response.status_code", fiber.StatusOK),
	)
	return ctxFiber.Status(fiber.StatusOK).JSON(resp)
}