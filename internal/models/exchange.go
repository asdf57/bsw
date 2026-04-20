package models

import (
	"time"

	"gorm.io/gorm"
)

type ExchangeDBEntry struct {
	gorm.Model
	FromCurrency string    `gorm:"uniqueIndex:idx_exchange_lookup"`
	ToCurrency   string    `gorm:"uniqueIndex:idx_exchange_lookup"`
	Date         time.Time `gorm:"uniqueIndex:idx_exchange_lookup;type:date"`
	Rate         float64
}

type Exchange struct {
	FromCurrency string
	ToCurrency   string
	Date         time.Time
	Rate         float64
}

func (ExchangeDBEntry) TableName() string { return "exchange_rates" }
