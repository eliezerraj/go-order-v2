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

type Payment struct {
	ID			int			`json:"id,omitempty"`
	Transaction string		`json:"transaction_id,omitempty"`
	Type		string 		`json:"type,omitempty"`
	Status		string 		`json:"status,omitempty"`
	Currency	string 		`json:"currency,omitempty"`
	Amount		float64 	`json:"amount,omitempty"`
	CreatedAt	time.Time 	`json:"created_at,omitempty"`
	UpdatedAt	*time.Time 	`json:"updated_at,omitempty"`
	StepProcess	*[]StepProcess `json:"step_process,omitempty"`			
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
}

type Price struct {
	ID				int			`json:"id,omitempty"`
	ProductId 		int		 	`json:"product_id,omitempty"`
	Currency		string		`json:"currency,omitempty"`
	Amount			float64		`json:"amount,omitempty"`
	StartedAt		*time.Time 	`json:"started_at,omitempty"`
	EndedAt			*time.Time 	`json:"ended_at,omitempty"`
}