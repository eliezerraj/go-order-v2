package entity

import (
	"time"
)

type Order struct {
	ID				int			`json:"id,omitempty"`
	OrderNumber		string		`json:"order_number,omitempty"`
	Transaction 	string		`json:"transaction_id,omitempty"`
	Date			time.Time 	`json:"order_date,omitempty"`
	Status			string 		`json:"status,omitempty"`
	Currency		string 		`json:"currency,omitempty"`
	Amount			float64 	`json:"amount,omitempty"`	
	CustomerID		string		`json:"customer_id,omitempty"`
	CreatedAt		time.Time 	`json:"created_at,omitempty"`
	UpdatedAt		*time.Time 	`json:"updated_at,omitempty"`
	OrderItem		*[]OrderItem	`json:"order_item,omitempty"`
	Payment			*[]Payment	`json:"payment,omitempty"`
	StepProcess		*[]StepProcess `json:"step_process,omitempty"`	
}

type OrderItem struct {
	ID				int			`json:"id,omitempty"`
	FkOrderID		int			`json:"fk_order_id,omitempty"`
	Product 		Product		`json:"product"`
	Status			string 		`json:"status,omitempty"`
	Quantity		int			`json:"quantity,omitempty"`
	Discount		float64		`json:"discount,omitempty"`
	Currency		string 		`json:"currency,omitempty"`	
	Amount			float64		`json:"amount,omitempty"`
	CreatedAt		time.Time 	`json:"created_at,omitempty"`
	UpdatedAt		*time.Time 	`json:"updated_at,omitempty"`	
}

type StepProcess struct {
	Name		string  	`json:"step_process,omitempty"`
	ProcessedAt	time.Time 	`json:"processed_at,omitempty"`
}

type Product struct {
	ID			int			`json:"id,omitempty"`
	Sku			string		`json:"sku,omitempty"`
	Type		string 		`json:"type,omitempty"`
	Name		string 		`json:"name,omitempty"`
	Status		string 		`json:"status,omitempty"`
	Price		*Price		`json:"price,omitempty"`
	Inventory	*Inventory	`json:"inventory,omitempty"`	
}

type Price struct {
	ID				int			`json:"id,omitempty"`
	ProductId 		int		 	`json:"product_id,omitempty"`
	Currency		string		`json:"currency,omitempty"`
	Amount			float64		`json:"amount,omitempty"`
	StartedAt		*time.Time 	`json:"started_at,omitempty"`
	EndedAt			*time.Time 	`json:"ended_at,omitempty"`
}

type Inventory struct {
	ID				int			`json:"id,omitempty"`
	Available		int			`json:"available,omitempty"`
	Pending			int			`json:"pending,omitempty"`
	Sold			int			`json:"sold,omitempty"`	
}

type Checkout struct {
	Order		Order		`json:"order,omitempty"`
	Payment		Payment		`json:"payment,omitempty"`
}

type Payment struct {
	ID			int			`json:"id,omitempty"`
	PaymentNumber string	`json:"payment_number,omitempty"`
	TransactionID string	`json:"transaction_id,omitempty"`
	Type		string 		`json:"type,omitempty"`
	Status		string 		`json:"status,omitempty"`
	PaymentDate	time.Time 	`json:"payment_date,omitempty"`
	Currency	string 		`json:"currency,omitempty"`
	Amount		float64 	`json:"amount,omitempty"`
	CreditCard	*CreditCard	`json:"credit_card,omitempty"`
	CreatedAt	time.Time 	`json:"created_at,omitempty"`		
}

type CreditCard struct {
	Pan				string	`json:"pan,omitempty"`
	Holder			string	`json:"holder,omitempty"`
	ExpirationDate	string	`json:"expiration_date,omitempty"`
	Password		string	`json:"password,omitempty"`
	CVV				string	`json:"cvv,omitempty"`
}