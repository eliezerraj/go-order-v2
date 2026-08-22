package adapter

import (
	"context"

	"go.uber.org/zap"

	"github.com/go-order-v2/application/config"
	"github.com/go-order-v2/application/infrastructure/application"
	"github.com/go-order-v2/application/domain/external"

	"github.com/eliezerraj/go-core/v3/logger"
	"github.com/eliezerraj/go-core/v3/http/utils"

	"github.com/gofiber/fiber/v2"

	"go.opentelemetry.io/otel"
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

	tracer := otel.Tracer("order.adapter")
	ctx, span := tracer.Start(ctxWithTimeout, "ApplicationAdapter.OrderGet")
	defer span.End()

	logger.Info(ctx, "OrderGet called")

	logger.Debug(
		ctxWithTimeout,
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

	res, err := a.application.OrderController.OrderGet(ctxWithTimeout, order)
	if err != nil {
		logger.Error(ctxWithTimeout, "failed to get order ", zap.Error(err))
		errorResponse := external.NewResponseError(ctxWithTimeout,
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

	logger.Info(ctxWithTimeout, "OrderAdd called")

	logger.Debug(
		ctxWithTimeout,
		a.cfg.App.Name,
		zap.ByteString("headers", utils.FormatHeadersAsJSON(ctxFiber.GetReqHeaders())),
		zap.ByteString("query", ctxFiber.Request().URI().QueryString()),
		zap.ByteString("body", ctxFiber.Body()),
	)
	order := external.OrderRequest{}
	if err := ctxFiber.BodyParser(&order); err != nil {
		logger.Error(ctxWithTimeout, "failed to parse request body", zap.Error(err))
		errorResponse := external.NewResponseError(ctxWithTimeout,
			fiber.StatusBadRequest,
			fiber.ErrBadRequest,
			fiber.ErrBadRequest.Message,
			"failed to parse request body",
			err.Error(),
			external.BUSSINESS_ERROR)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	res, err := a.application.OrderController.OrderAdd(ctxWithTimeout, order)
	if err != nil {
		logger.Error(ctxWithTimeout, "failed to add order", zap.Error(err))
		errorResponse := external.NewResponseError(ctxWithTimeout,
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

	tracer := otel.Tracer("checkout.adapter")
	ctx, span := tracer.Start(ctxWithTimeout, "ApplicationAdapter.CheckoutGet")
	defer span.End()

	logger.Info(ctx, "CheckoutGet called")

	logger.Debug(
		ctxWithTimeout,
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
		OrderNumber: order_number,
	}

	res, err := a.application.CheckoutController.CheckoutGet(ctxWithTimeout, checkout)
	if err != nil {
		logger.Error(ctxWithTimeout, "failed to get checkout ", zap.Error(err))
		errorResponse := external.NewResponseError(ctxWithTimeout,
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

func (a *ApplicationAdapter) CheckoutAdd(ctxFiber *fiber.Ctx) error {
	ctxWithTimeout, cancel := context.WithTimeout(ctxFiber.UserContext(), a.cfg.HTTP.Timeout)
	defer cancel()

	logger.Info(ctxWithTimeout, "CheckoutAdd called")

	logger.Debug(
		ctxWithTimeout,
		a.cfg.App.Name,
		zap.ByteString("headers", utils.FormatHeadersAsJSON(ctxFiber.GetReqHeaders())),
		zap.ByteString("query", ctxFiber.Request().URI().QueryString()),
		zap.ByteString("body", ctxFiber.Body()),
	)

	checkout := external.CheckoutRequest{}
	if err := ctxFiber.BodyParser(&checkout); err != nil {
		logger.Error(ctxWithTimeout, "failed to parse request body", zap.Error(err))
		errorResponse := external.NewResponseError(ctxWithTimeout,
			fiber.StatusBadRequest,
			fiber.ErrBadRequest,
			fiber.ErrBadRequest.Message,
			"failed to parse request body",
			err.Error(),
			external.BUSSINESS_ERROR)
		return ctxFiber.Status(errorResponse.StatusCode).JSON(errorResponse)
	}

	res, err := a.application.CheckoutController.CheckoutAdd(ctxWithTimeout, checkout)
	if err != nil {
		logger.Error(ctxWithTimeout, "failed to add checkout", zap.Error(err))
		errorResponse := external.NewResponseError(ctxWithTimeout,
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