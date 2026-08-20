package external

import (
	"github.com/go-order-v2/application/domain/entity"
)

type OrderRequest struct {
	OrderNumber		string		`json:"order_number,omitempty"`
	Date			string	`json:"order_date,omitempty"`
	CustomerID		string		`json:"customer_id,omitempty"`
	OrderItem		[]*OrderItemRequest	`json:"order_item,omitempty"`
}

type OrderItemRequest struct {
	Product 		ProductRequest	`json:"product"`
	Quantity		int			`json:"quantity,omitempty"`
	Discount		float64		`json:"discount,omitempty"`
}

type ProductRequest struct {
	Sku			string		`json:"sku,omitempty"`
	Price		*PriceRequest	`json:"price,omitempty"`
}

type PriceRequest struct {
	Currency		string		`json:"currency,omitempty"`
	Amount			float64		`json:"amount,omitempty"`	
}

type OrderResponse struct {
	Response    string	`json:"response"`
	Order		any	`json:"order,omitempty"`
}

type InventoryResponse struct {
    Response string         `json:"response"`
    Product  entity.Product `json:"product"`
}