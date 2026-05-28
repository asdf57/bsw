package db

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type SettlementDBEntry struct {
	gorm.Model
	OwedByID   uint `gorm:"not null"`
	OwedBy     UserDBEntry
	OwedToID   uint `gorm:"not null"`
	OwedTo     UserDBEntry
	Amount     decimal.Decimal `gorm:"type:numeric(19,4);not null"`
	Currency   string          `gorm:"not null"`
	Date       time.Time       `gorm:"not null"`
	ReversedAt *time.Time
}

func (SettlementDBEntry) TableName() string { return "settlements" }
