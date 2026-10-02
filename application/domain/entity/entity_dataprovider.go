package entity

import (
	"time"
)

type TimeSeriesOrderItem struct {
	Product  Product 	`json:"product,omitempty"`
	TimeSeriesData 	[]*TimeSeriesData	`json:"time_series_data,omitempty"`
}

type TimeSeriesData struct {
	DatePart   time.Time 	`json:"date,omitempty"`
	HourPart   int 			`json:"hour,omitempty"`
	Count      int 			`json:"count,omitempty"`
	SumQuantity int 		`json:"sum_quantity,omitempty"`
	SumAmount   float64 	`json:"sum_amount,omitempty"`
	SumDiscount float64 	`json:"sum_discount,omitempty"`
}