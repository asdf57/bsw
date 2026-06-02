package debts

import (
	"fmt"

	dbmodels "github.com/asdf57/bsw/internal/models/db"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func canonicalPair(user1, user2 uint) (low, high uint) {
	if user1 < user2 {
		return user1, user2
	}

	return user2, user1
}

func GetOwedByUser(debt dbmodels.DebtDBEntry) uint {
	if debt.NetAmount.IsNegative() {
		return debt.UserHighId
	}

	return debt.UserLowId
}

func GetOwedToUser(debt dbmodels.DebtDBEntry) uint {
	if debt.NetAmount.IsNegative() {
		return debt.UserLowId
	}

	return debt.UserHighId
}

func UserOwesDebt(debt dbmodels.DebtDBEntry, userID uint) bool {
	return GetOwedByUser(debt) == userID
}

// Returns the canonical pair and delta amount
func signedDelta(owedBy, owedTo uint, amount decimal.Decimal) (low, high uint, delta decimal.Decimal, err error) {
	if owedBy == owedTo {
		return 0, 0, decimal.Zero, fmt.Errorf("cannot create debt to self")
	}

	lowUserId, highUserId := canonicalPair(owedBy, owedTo)

	if owedBy == lowUserId && owedTo == highUserId {
		// low owes high => return +amount
		return lowUserId, highUserId, amount, nil
	}

	// high owes low => return -amount
	return lowUserId, highUserId, amount.Neg(), nil
}

// Adhering to our axioms, apply the net debt entry to the DB
func ApplyNetDebt(tx *gorm.DB, owedBy, owedTo uint, amount decimal.Decimal, currency string) error {
	// Obtain our signedDelta so we know WHAT to add
	lowUserId, highUserId, delta, err := signedDelta(owedBy, owedTo, amount)
	if err != nil {
		return err
	}

	// Now we know who owes who (direction) and our canonical pair (so we can index in the debts table)

	key := dbmodels.DebtDBEntry{
		UserLowId:  lowUserId,
		UserHighId: highUserId,
		Currency:   currency,
	}

	insert := dbmodels.DebtDBEntry{
		UserLowId:  lowUserId,
		UserHighId: highUserId,
		NetAmount:  delta,
		Currency:   currency,
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_low_id"},
			{Name: "user_high_id"},
			{Name: "currency"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"net_amount": gorm.Expr("debts.net_amount + EXCLUDED.net_amount"),
		}),
	}).Create(&insert).Error; err != nil {
		return fmt.Errorf("upsert net debt: %w", err)
	}

	var debt dbmodels.DebtDBEntry
	if err := tx.Where(&key).First(&debt).Error; err != nil {
		return fmt.Errorf("fetch net debt: %w", err)
	}

	if debt.NetAmount.IsZero() {
		if err := tx.Unscoped().Delete(&debt).Error; err != nil {
			return fmt.Errorf("delete zero net debt: %w", err)
		}
	}

	return nil
}

func DebtsToSettle(tx *gorm.DB, owedBy uint, owedTo *uint) ([]dbmodels.DebtDBEntry, error) {
	if owedTo != nil && owedBy == *owedTo {
		return nil, fmt.Errorf("cannot settle debt to self")
	}

	query := tx.Model(&dbmodels.DebtDBEntry{}).Where(
		"(user_low_id = ? AND net_amount > 0) OR (user_high_id = ? AND net_amount < 0)",
		owedBy,
		owedBy,
	)

	if owedTo != nil {
		lowUserID, highUserID := canonicalPair(owedBy, *owedTo)
		query = query.Where("user_low_id = ? AND user_high_id = ?", lowUserID, highUserID)
	}

	var debts []dbmodels.DebtDBEntry
	if err := query.Find(&debts).Error; err != nil {
		return nil, fmt.Errorf("query debts to settle: %w", err)
	}

	return debts, nil
}

func SettleDebts(tx *gorm.DB, owedBy uint, owedTo *uint) (int64, error) {
	debts, err := DebtsToSettle(tx, owedBy, owedTo)
	if err != nil {
		return 0, err
	}
	if len(debts) == 0 {
		return 0, nil
	}

	ids := make([]uint, 0, len(debts))
	for _, debt := range debts {
		ids = append(ids, debt.ID)
	}

	result := tx.Unscoped().Delete(&dbmodels.DebtDBEntry{}, ids)
	if result.Error != nil {
		return 0, fmt.Errorf("settle debts: %w", result.Error)
	}

	return result.RowsAffected, nil
}

func SettleDebtAmount(tx *gorm.DB, debt dbmodels.DebtDBEntry, amount decimal.Decimal) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("settlement amount must be greater than zero")
	}
	total := debt.NetAmount.Abs()
	if amount.GreaterThan(total) {
		return fmt.Errorf("settlement amount cannot exceed %s", total.StringFixed(2))
	}

	owedBy := GetOwedByUser(debt)
	owedTo := GetOwedToUser(debt)
	if amount.Equal(total) {
		return tx.Unscoped().Delete(&debt).Error
	}

	return ApplyNetDebt(tx, owedTo, owedBy, amount, debt.Currency)
}
