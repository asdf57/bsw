package api

import "github.com/shopspring/decimal"

type DebtEntry struct {
	OwedByUser string  `json:"owedByUser"`
	OwedToUser string  `json:"owedToUser"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
}

type DebtResponse struct {
	ID             uint            `json:"id"`
	OwedByUserID   uint            `json:"owedByUserId"`
	OwedToUserID   uint            `json:"owedToUserId"`
	OwedByUserName string          `json:"owedByUserName,omitempty"`
	OwedToUserName string          `json:"owedToUserName,omitempty"`
	Amount         decimal.Decimal `json:"amount"`
	Currency       string          `json:"currency"`
}
