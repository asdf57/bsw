package api

import (
	"time"

	"github.com/shopspring/decimal"
)

type Payment struct {
	Amount           float64   `json:"amount"`
	Payer            string    `json:"payer"`
	Description      string    `json:"description"`
	Date             time.Time `json:"date"`
	FromExchangeRate string    `json:"fromExchangeRate"`
	ToExchangeRate   string    `json:"toExchangeRate"`
	Debtors          []string  `json:"debtors"`
	DebtMode         string    `json:"debtMode"`
}

type PaymentResponse struct {
	ID          uint            `json:"id"`
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description"`
	Date        time.Time       `json:"date"`
	Payer       UserSummary     `json:"payer"`
	Debtors     []UserSummary   `json:"debtors"`
	Exchange    ExchangeSummary `json:"exchange"`
}
