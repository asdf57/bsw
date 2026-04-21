package api

import "time"

type Exchange struct {
	FromCurrency string    `json:"fromCurrency"`
	ToCurrency   string    `json:"toCurrency"`
	Date         time.Time `json:"date"`
	Rate         float64   `json:"rate"`
}

type ExchangeSummary struct {
	ID           uint      `json:"id"`
	FromCurrency string    `json:"fromCurrency"`
	ToCurrency   string    `json:"toCurrency"`
	Date         time.Time `json:"date"`
	Rate         float64   `json:"rate"`
}
