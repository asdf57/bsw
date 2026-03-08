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
	Owers       []UserDBEntry `gorm:"many2many:payment_owers;"`
}

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

type User struct {
	Name string
}
