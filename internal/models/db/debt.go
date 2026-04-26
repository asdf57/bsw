package db

import (
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

/*
* DebtDBEntry represents a canonical row per unordered pair + currency.
* By forming one row per unique pair of users and currency, we can gurantee
* that all debt calculations remain atomic.
*
* Debt entries have an invariant requiring that there may not be > 1 row per
* unique pair of users and currency. We gurantee that the SIMPLIFIED form
* of the debt relation is stored, meaning that one user will ALWAYS owe the
* other 0 of that currency.
*
* Convention:
* - UserLowID < UserHighID: Guranteed ordering of users in the debt row
* - NetAmount > 0: UserLow owes UserHigh
* - NetAmount < 0: UserHigh owes UserLow
* - NetAmount == 0: delete the row
 */
type DebtDBEntry struct {
	gorm.Model
	UserLowId  uint            `gorm:"not null;uniqueIndex:idx_pair_currency"`
	UserHighId uint            `gorm:"not null;uniqueIndex:idx_pair_currency"`
	NetAmount  decimal.Decimal `gorm:"type:numeric(19,4);not null"`
	Currency   string          `gorm:"not null;uniqueIndex:idx_pair_currency"`
}

func (DebtDBEntry) TableName() string { return "debts" }
