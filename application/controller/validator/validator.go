package validator

import (
    "strconv"
	"context"
	"errors"

	"github.com/go-order-v2/application/domain/external"
    "github.com/go-order-v2/application/shared/helpers"
)

// Schema struct defines a validation schema for product requests.
type Schema struct {
    Validate func(context.Context, any) error
}

func (s *Schema) OrderAddSchema() Schema {
    return Schema{
        Validate: func(ctx context.Context, data any) error {
			
			req, ok := data.(external.OrderRequest)
			if !ok {
                return errors.New("schema validation failed ! Please check the request body and try again.")
            }

			if req.OrderItem == nil {
                 return errors.New("schema validation failed ! field order_item is mandatory")
            }

			if req.OrderNumber == "" {
                return errors.New("schema validation failed ! field order_number is mandatory")
            }

            for i, item := range req.OrderItem {
                if item.Product.Sku == "" {
                    return errors.New("schema validation failed ! field product.sku is mandatory for order item at index " + strconv.Itoa(i))
                }
                if item.Quantity <= 0 {
                    return errors.New("schema validation failed ! field quantity must be greater than 0 for order item at index " + strconv.Itoa(i))
                }            
            }

            if req.Date != "" {
                _, err := helpers.ParseDate(req.Date)
                if err != nil {
                    return errors.New("schema validation failed ! field date is not in a valid format")
                }
            }

            return nil
        },
    }
}

func (s *Schema) CheckoutAddSchema() Schema {
    return Schema{
        Validate: func(ctx context.Context, data any) error {
			
			req, ok := data.(external.CheckoutRequest)
			if !ok {
                return errors.New("schema validation failed ! Please check the request body and try again.")
            }

			if req.Order == nil {
                 return errors.New("schema validation failed ! field order is mandatory")
            }

			if req.Payment == nil {
                 return errors.New("schema validation failed ! field payment is mandatory")
            }

            for i, item := range req.Payment.PaymentDetail {
                if item.Amount <= 0 {
                    return errors.New("schema validation failed ! field amount must be greater than 0 for payment detail at index " + strconv.Itoa(i))
                }
                if item.Currency == "" {
                    return errors.New("schema validation failed ! field currency is mandatory for payment detail at index " + strconv.Itoa(i))
                }
                if item.CreditCard == nil {
                    return errors.New("schema validation failed ! field credit_card is mandatory for payment detail at index " + strconv.Itoa(i))
                }
            }

            return nil
        },
    }
}
