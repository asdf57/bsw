package api

import "github.com/shopspring/decimal"

type DebtEntry struct {
	OwedByUser string  `json:"owedByUser"`
	OwedToUser string  `json:"owedToUser"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
}

type DebtResponse struct {
	ID         uint            `json:"id"`
	OwedByUser UserSummary     `json:"owedByUser"`
	OwedToUser UserSummary     `json:"owedToUser"`
	Amount     decimal.Decimal `json:"amount"`
	Currency   string          `json:"currency"`
}
