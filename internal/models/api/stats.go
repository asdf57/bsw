package api

import "github.com/shopspring/decimal"

type UserStatsResponse struct {
	User             UserSummary     `json:"user"`
	Currency         string          `json:"currency"`
	TotalSpent       decimal.Decimal `json:"totalSpent"`
	SpentOut         decimal.Decimal `json:"spentOut"`
	PaymentsIncluded int             `json:"paymentsIncluded"`
}
