package db

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type DebtDBEntry struct {
	gorm.Model
	OwedByUserId uint            `gorm:"not null;index:idx_from_to,unique"`
	OwedToUserId uint            `gorm:"not null;index:idx_from_to,unique"`
	Amount       decimal.Decimal `gorm:"type:numeric(19,4)"`
	Currency     string
}

func (DebtDBEntry) TableName() string { return "debts" }
