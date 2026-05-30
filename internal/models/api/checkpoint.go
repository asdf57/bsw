package api

import (
	"time"

	"github.com/shopspring/decimal"
)

type Checkpoint struct {
	Version       int                    `json:"version"`
	ExportedAt    time.Time              `json:"exportedAt,omitempty"`
	Users         []CheckpointUser       `json:"users"`
	Tags          []CheckpointTag        `json:"tags,omitempty"`
	Payments      []CheckpointPayment    `json:"payments"`
	Debts         []CheckpointDebt       `json:"debts"`
	Settlements   []CheckpointSettlement `json:"settlements"`
	ExchangeRates []Exchange             `json:"exchangeRates"`
}

type CheckpointUser struct {
	ID            uint   `json:"id,omitempty"`
	Name          string `json:"name"`
	DiscordHandle string `json:"discordHandle,omitempty"`
}

type CheckpointTag struct {
	ID   uint   `json:"id,omitempty"`
	Name string `json:"name"`
}

type CheckpointPayment struct {
	ID          uint            `json:"id,omitempty"`
	Amount      decimal.Decimal `json:"amount"`
	Description string          `json:"description,omitempty"`
	Date        time.Time       `json:"date"`
	Payer       string          `json:"payer"`
	Currency    string          `json:"currency"`
	Debtors     []string        `json:"debtors,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
}

type CheckpointDebt struct {
	ID        uint            `json:"id,omitempty"`
	OwedBy    string          `json:"owedBy"`
	OwedTo    string          `json:"owedTo"`
	Amount    decimal.Decimal `json:"amount"`
	Currency  string          `json:"currency"`
	CreatedAt time.Time       `json:"createdAt,omitempty"`
	UpdatedAt time.Time       `json:"updatedAt,omitempty"`
}

type CheckpointSettlement struct {
	ID         uint            `json:"id,omitempty"`
	OwedBy     string          `json:"owedBy"`
	OwedTo     string          `json:"owedTo"`
	Amount     decimal.Decimal `json:"amount"`
	Currency   string          `json:"currency"`
	Date       time.Time       `json:"date"`
	ReversedAt *time.Time      `json:"reversedAt,omitempty"`
}

type CheckpointImportResponse struct {
	Users         int `json:"users"`
	Tags          int `json:"tags"`
	Payments      int `json:"payments"`
	Debts         int `json:"debts"`
	Settlements   int `json:"settlements"`
	ExchangeRates int `json:"exchangeRates"`
}
