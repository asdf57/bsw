package api

import (
	"time"

	"github.com/shopspring/decimal"
)

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

type SettleDebtsRequest struct {
	OwedBy string           `json:"owedBy" binding:"required"`
	OwedTo string           `json:"owedTo"`
	Amount *decimal.Decimal `json:"amount,omitempty"`
}

type SettleDebtsResponse struct {
	OwedBy       string               `json:"owedBy"`
	OwedTo       string               `json:"owedTo,omitempty"`
	SettledCount int64                `json:"settledCount"`
	Settlements  []SettlementResponse `json:"settlements"`
}

type SettlementResponse struct {
	ID         uint            `json:"id"`
	OwedByUser UserSummary     `json:"owedByUser"`
	OwedToUser UserSummary     `json:"owedToUser"`
	Amount     decimal.Decimal `json:"amount"`
	Currency   string          `json:"currency"`
	Date       time.Time       `json:"date"`
	ReversedAt *time.Time      `json:"reversedAt,omitempty"`
}
