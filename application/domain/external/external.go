package external

import (
	"time"
)

type OrderRequest struct {
	OrderNumber		string		`json:"order_number,omitempty"`
	Date			time.Time 	`json:"order_date,omitempty"`
	Status			string 		`json:"status,omitempty"`
	Currency		string 		`json:"currency,omitempty"`
	Amount			float64 	`json:"amount,omitempty"`	
	User			string		`json:"user_id,omitempty"`
	CartItem		*CartItemRequest	`json:"cart_item,omitempty"`
}

type CartItemRequest struct {
	Product 		ProductRequest	`json:"product"`
	Quantity		int			`json:"quantity,omitempty"`
	Discount		float64		`json:"discount,omitempty"`
	Currency		string 		`json:"currency,omitempty"`	
	Price			float64		`json:"price,omitempty"`
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