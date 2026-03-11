package models

import (
	"time"

	"gorm.io/gorm"
)

// payments table -- id,  amount, description, date, payer_id, user ids of those who owe money
// Tracks who made a payment and who owes money to whom
type PaymentDBEntry struct {
	gorm.Model
	Amount      float64
	Description string
	Date        time.Time
	PayerID     uint
	Owers       []UserDBEntry `gorm:"many2many:payment_owers;constraint:OnDelete:CASCADE;"`
}

func (PaymentDBEntry) TableName() string { return "payments" }

type Payment struct {
	Amount      float64
	Payer       string
	Description string
	Owers       []string
}

// User table -- id, name
type UserDBEntry struct {
	gorm.Model
	Name string
}

func (UserDBEntry) TableName() string { return "users" }

type User struct {
	Name string
}

type BalanceDBEntry struct {
	gorm.Model
	FromUserID uint    `gorm:"not null;index:idx_from_to,unique"`
	ToUserID   uint    `gorm:"not null;index:idx_from_to,unique"`
	Amount     float64 `gorm:"not null"`
}

func (BalanceDBEntry) TableName() string { return "balances" }

type Balance struct {
	FromUser string
	ToUser   string
	Amount   float64
}
