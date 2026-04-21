package db

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
	Payer       UserDBEntry
	ExchangeID  uint
	Exchange    ExchangeDBEntry
	Debtors     []UserDBEntry `gorm:"many2many:debtors;constraint:OnDelete:CASCADE;"`
}

func (PaymentDBEntry) TableName() string { return "payments" }
