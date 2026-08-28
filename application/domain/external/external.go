package external

import ( 
	"time"
)

import (
	"github.com/go-order-v2/application/domain/entity"
)

type OrderRequest struct {
	ID				int		`json:"id,omitempty"`
	OrderNumber		string	`json:"order_number,omitempty"`
	Date			string	`json:"order_date,omitempty"`
	CustomerID		string	`json:"customer_id,omitempty"`
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

type CheckoutRequest struct {
	Order		OrderRequest		`json:"order,omitempty"`
	Payment		PaymentRequest		`json:"payment,omitempty"`
}

type PaymentRequest struct {
	PaymentNumber 	string	`json:"payment_number,omitempty"`
	TransactionID	string	`json:"transaction_id,omitempty"`
	Type			string	`json:"type,omitempty"`
	Order			OrderRequest	`json:"order,omitempty"`
	PaymentDetail	[]*PaymentDetailRequest	`json:"payment_detail,omitempty"`
}

type PaymentDetailRequest struct {
	DetailDate	time.Time 	`json:"payment_detail_date,omitempty"`
	Status		string 		`json:"status,omitempty"`
	Currency	string 		`json:"currency,omitempty"`
	Amount		float64 	`json:"amount,omitempty"`
	CreditCard	*CreditCardRequest	`json:"credit_card,omitempty"`
}

type CreditCardRequest struct {
	Pan		string	`json:"pan,omitempty"`
	Holder	string	`json:"holder,omitempty"`
	CVV		string	`json:"cvv,omitempty"`
	Password string	`json:"password,omitempty"`
}

type PaymentResponse struct {
	Response    string	`json:"response"`
	Payment		entity.PaymentCheckout	`json:"payment,omitempty"`
}

type CheckoutResponse struct {
    Response string         `json:"response"`
    Checkout  any `json:"checkout"`
}