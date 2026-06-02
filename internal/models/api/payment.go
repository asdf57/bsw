package api

import (
	"time"

	"github.com/shopspring/decimal"
)

type Payment struct {
	Amount      float64   `json:"amount"`
	Payer       string    `json:"payer"`
	Description string    `json:"description"`
	Date        time.Time `json:"date"`
	Currency    string    `json:"currency"`
	Debtors     []string  `json:"debtors"`
	Tags        []string  `json:"tags"`
	DebtMode    string    `json:"debtMode"`
}

type PaymentCreateResponse struct {
	ID   uint   `json:"id"`
	Info string `json:"info"`
}

type PaymentResponse struct {
	ID          uint            `json:"id"`
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description"`
	Date        time.Time       `json:"date"`
	Payer       UserSummary     `json:"payer"`
	Debtors     []UserSummary   `json:"debtors"`
	Currency    string          `json:"currency"`
	Tags        []string        `json:"tags"`
}
