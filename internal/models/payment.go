package models

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type PaymentDBEntry struct {
	gorm.Model
	Amount      decimal.Decimal `gorm:"type:numeric(19,4)"`
	Description string
	Date        time.Time
	PayerID     uint

	ExchangeID uint
	Exchange   ExchangeDBEntry

	Debtors []UserDBEntry `gorm:"many2many:debtors;constraint:OnDelete:CASCADE;"`
}

func (PaymentDBEntry) TableName() string { return "payments" }

type Payment struct {
	Amount           float64
	Payer            string
	Description      string
	Date             time.Time
	FromExchangeRate string
	ToExchangeRate   string
	Debtors          []string
	DebtMode         string
}
