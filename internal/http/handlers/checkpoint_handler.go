package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/asdf57/bsw/internal/currency"
	"github.com/asdf57/bsw/internal/debts"
	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// ExportCheckpoint godoc
// @Summary Export a portable system checkpoint
// @Tags admin
// @Produce json
// @Success 200 {object} api.Checkpoint
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/export [get]
func (h *Handlers) ExportCheckpoint(c *gin.Context) {
	checkpoint, err := h.buildCheckpoint()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=bsw-checkpoint-%s.json", time.Now().UTC().Format("20060102-150405")))
	c.JSON(http.StatusOK, checkpoint)
}

// ImportCheckpoint godoc
// @Summary Import a portable system checkpoint and replace current app state
// @Tags admin
// @Accept json
// @Produce json
// @Param checkpoint body api.Checkpoint true "Checkpoint payload"
// @Success 200 {object} api.CheckpointImportResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/import [post]
func (h *Handlers) ImportCheckpoint(c *gin.Context) {
	var checkpoint apimodels.Checkpoint
	if err := c.ShouldBindJSON(&checkpoint); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if checkpoint.Version == 0 {
		checkpoint.Version = 1
	}
	if checkpoint.Version != 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported checkpoint version"})
		return
	}

	summary, err := h.importCheckpoint(checkpoint)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, summary)
}

func (h *Handlers) buildCheckpoint() (apimodels.Checkpoint, error) {
	var users []dbmodels.UserDBEntry
	if err := h.Db.DB.Order("id").Find(&users).Error; err != nil {
		return apimodels.Checkpoint{}, fmt.Errorf("export users: %w", err)
	}
	var tags []dbmodels.TagDBEntry
	if err := h.Db.DB.Order("id").Find(&tags).Error; err != nil {
		return apimodels.Checkpoint{}, fmt.Errorf("export tags: %w", err)
	}
	var payments []dbmodels.PaymentDBEntry
	if err := preloadPayment(h.Db.DB).Order("payments.id").Find(&payments).Error; err != nil {
		return apimodels.Checkpoint{}, fmt.Errorf("export payments: %w", err)
	}
	var debtRows []dbmodels.DebtDBEntry
	if err := h.Db.DB.Order("id").Find(&debtRows).Error; err != nil {
		return apimodels.Checkpoint{}, fmt.Errorf("export debts: %w", err)
	}
	var settlements []dbmodels.SettlementDBEntry
	if err := h.Db.DB.Preload("OwedBy").Preload("OwedTo").Order("id").Find(&settlements).Error; err != nil {
		return apimodels.Checkpoint{}, fmt.Errorf("export settlements: %w", err)
	}
	var exchangeRates []dbmodels.ExchangeDBEntry
	if err := h.Db.DB.Order("date, from_currency, to_currency").Find(&exchangeRates).Error; err != nil {
		return apimodels.Checkpoint{}, fmt.Errorf("export exchange rates: %w", err)
	}

	usersByID := make(map[uint]dbmodels.UserDBEntry, len(users))
	for _, user := range users {
		usersByID[user.ID] = user
	}

	checkpoint := apimodels.Checkpoint{
		Version:       1,
		ExportedAt:    time.Now().UTC(),
		Users:         make([]apimodels.CheckpointUser, 0, len(users)),
		Tags:          make([]apimodels.CheckpointTag, 0, len(tags)),
		Payments:      make([]apimodels.CheckpointPayment, 0, len(payments)),
		Debts:         make([]apimodels.CheckpointDebt, 0, len(debtRows)),
		Settlements:   make([]apimodels.CheckpointSettlement, 0, len(settlements)),
		ExchangeRates: make([]apimodels.Exchange, 0, len(exchangeRates)),
	}

	for _, user := range users {
		checkpoint.Users = append(checkpoint.Users, apimodels.CheckpointUser{ID: user.ID, Name: user.Name, DiscordHandle: user.DiscordHandle})
	}
	for _, tag := range tags {
		checkpoint.Tags = append(checkpoint.Tags, apimodels.CheckpointTag{ID: tag.ID, Name: tag.Name})
	}
	for _, payment := range payments {
		debtors := make([]string, 0, len(payment.Debtors))
		for _, debtor := range payment.Debtors {
			debtors = append(debtors, debtor.Name)
		}
		sort.Strings(debtors)

		paymentTags := make([]string, 0, len(payment.Tags))
		for _, tag := range payment.Tags {
			paymentTags = append(paymentTags, tag.Name)
		}
		sort.Strings(paymentTags)

		checkpoint.Payments = append(checkpoint.Payments, apimodels.CheckpointPayment{
			ID:          payment.ID,
			Amount:      payment.Amount,
			Description: payment.Description,
			Date:        payment.Date,
			Payer:       payment.Payer.Name,
			Currency:    payment.Currency,
			Debtors:     debtors,
			Tags:        paymentTags,
		})
	}
	for _, debt := range debtRows {
		owedByID := debts.GetOwedByUser(debt)
		owedToID := debts.GetOwedToUser(debt)
		checkpoint.Debts = append(checkpoint.Debts, apimodels.CheckpointDebt{
			ID:        debt.ID,
			OwedBy:    usersByID[owedByID].Name,
			OwedTo:    usersByID[owedToID].Name,
			Amount:    debt.NetAmount.Abs(),
			Currency:  debt.Currency,
			CreatedAt: debt.CreatedAt,
			UpdatedAt: debt.UpdatedAt,
		})
	}
	for _, settlement := range settlements {
		checkpoint.Settlements = append(checkpoint.Settlements, apimodels.CheckpointSettlement{
			ID:         settlement.ID,
			OwedBy:     settlement.OwedBy.Name,
			OwedTo:     settlement.OwedTo.Name,
			Amount:     settlement.Amount,
			Currency:   settlement.Currency,
			Date:       settlement.Date,
			ReversedAt: settlement.ReversedAt,
		})
	}
	for _, rate := range exchangeRates {
		checkpoint.ExchangeRates = append(checkpoint.ExchangeRates, apimodels.Exchange{
			FromCurrency: rate.FromCurrency,
			ToCurrency:   rate.ToCurrency,
			Date:         rate.Date,
			Rate:         rate.Rate,
		})
	}

	return checkpoint, nil
}

func (h *Handlers) importCheckpoint(checkpoint apimodels.Checkpoint) (apimodels.CheckpointImportResponse, error) {
	var summary apimodels.CheckpointImportResponse

	err := h.Db.DB.Transaction(func(tx *gorm.DB) error {
		if err := clearCheckpointTables(tx); err != nil {
			return err
		}

		usersByName, err := createCheckpointUsers(tx, checkpoint.Users)
		if err != nil {
			return err
		}
		tagsByName, err := createCheckpointTags(tx, checkpoint.Tags, checkpoint.Payments)
		if err != nil {
			return err
		}
		if err := createCheckpointPayments(tx, checkpoint.Payments, usersByName, tagsByName, checkpoint.Debts == nil); err != nil {
			return err
		}
		if err := createCheckpointSettlements(tx, checkpoint.Settlements, usersByName); err != nil {
			return err
		}
		if checkpoint.Debts == nil {
			if err := applyCheckpointSettlementsToDebts(tx, checkpoint.Settlements, usersByName); err != nil {
				return err
			}
		} else if err := createCheckpointDebts(tx, checkpoint.Debts, usersByName); err != nil {
			return err
		}
		if err := createCheckpointExchangeRates(tx, checkpoint.ExchangeRates); err != nil {
			return err
		}
		if err := resetCheckpointSequences(tx); err != nil {
			return err
		}

		summary = apimodels.CheckpointImportResponse{
			Users:         len(checkpoint.Users),
			Tags:          len(tagsByName),
			Payments:      len(checkpoint.Payments),
			Debts:         countCheckpointDebts(checkpoint),
			Settlements:   len(checkpoint.Settlements),
			ExchangeRates: len(checkpoint.ExchangeRates),
		}
		return nil
	})
	if err != nil {
		return apimodels.CheckpointImportResponse{}, err
	}

	return summary, nil
}

func clearCheckpointTables(tx *gorm.DB) error {
	for _, table := range []string{"debtors", "payment_tags"} {
		if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
			return fmt.Errorf("clear %s: %w", table, err)
		}
	}

	models := []any{
		&dbmodels.SettlementDBEntry{},
		&dbmodels.DebtDBEntry{},
		&dbmodels.PaymentDBEntry{},
		&dbmodels.TagDBEntry{},
		&dbmodels.UserDBEntry{},
		&dbmodels.ExchangeDBEntry{},
	}
	for _, model := range models {
		if err := tx.Unscoped().Where("1 = 1").Delete(model).Error; err != nil {
			return fmt.Errorf("clear table: %w", err)
		}
	}

	return nil
}

func createCheckpointUsers(tx *gorm.DB, users []apimodels.CheckpointUser) (map[string]dbmodels.UserDBEntry, error) {
	usersByName := make(map[string]dbmodels.UserDBEntry, len(users))
	for _, user := range users {
		name := strings.TrimSpace(user.Name)
		if name == "" {
			return nil, fmt.Errorf("user name is required")
		}
		if _, exists := usersByName[name]; exists {
			return nil, fmt.Errorf("duplicate user %q", name)
		}

		record := dbmodels.UserDBEntry{Name: name, DiscordHandle: strings.TrimSpace(user.DiscordHandle)}
		record.ID = user.ID
		if err := tx.Create(&record).Error; err != nil {
			return nil, fmt.Errorf("create user %q: %w", name, err)
		}
		usersByName[name] = record
	}
	return usersByName, nil
}

func createCheckpointTags(tx *gorm.DB, tags []apimodels.CheckpointTag, payments []apimodels.CheckpointPayment) (map[string]dbmodels.TagDBEntry, error) {
	required := map[string]uint{"general": 0}
	for _, tag := range tags {
		name := normalizeCheckpointTag(tag.Name)
		if name == "" {
			return nil, fmt.Errorf("tag name is required")
		}
		required[name] = tag.ID
	}
	for _, payment := range payments {
		paymentTags := NormalizeTags(payment.Tags)
		if len(paymentTags) == 0 {
			paymentTags = []string{"general"}
		}
		for _, tag := range paymentTags {
			if _, ok := required[tag]; !ok {
				required[tag] = 0
			}
		}
	}

	names := make([]string, 0, len(required))
	for name := range required {
		names = append(names, name)
	}
	sort.Strings(names)

	tagsByName := make(map[string]dbmodels.TagDBEntry, len(names))
	for _, name := range names {
		record := dbmodels.TagDBEntry{Name: name}
		record.ID = required[name]
		if err := tx.Create(&record).Error; err != nil {
			return nil, fmt.Errorf("create tag %q: %w", name, err)
		}
		tagsByName[name] = record
	}
	return tagsByName, nil
}

func createCheckpointPayments(tx *gorm.DB, payments []apimodels.CheckpointPayment, usersByName map[string]dbmodels.UserDBEntry, tagsByName map[string]dbmodels.TagDBEntry, rebuildDebts bool) error {
	for _, payment := range payments {
		payer, ok := usersByName[strings.TrimSpace(payment.Payer)]
		if !ok {
			return fmt.Errorf("payment payer %q does not exist", payment.Payer)
		}
		debtorsForPayment, err := checkpointUsersFromNames(usersByName, payment.Debtors)
		if err != nil {
			return err
		}
		for _, debtor := range debtorsForPayment {
			if debtor.ID == payer.ID {
				return fmt.Errorf("user %s cannot be indebted to themselves", debtor.Name)
			}
		}

		paymentTags := NormalizeTags(payment.Tags)
		if len(paymentTags) == 0 {
			paymentTags = []string{"general"}
		}
		tagsForPayment := make([]dbmodels.TagDBEntry, 0, len(paymentTags))
		for _, tagName := range paymentTags {
			tag, ok := tagsByName[tagName]
			if !ok {
				return fmt.Errorf("payment tag %q does not exist", tagName)
			}
			tagsForPayment = append(tagsForPayment, tag)
		}

		paymentDate := payment.Date
		if paymentDate.IsZero() {
			paymentDate = time.Now().UTC()
		}
		record := dbmodels.PaymentDBEntry{
			Amount:      payment.Amount,
			Description: payment.Description,
			Date:        paymentDate,
			PayerID:     payer.ID,
			Currency:    currency.NormalizeCurrencyCode(payment.Currency),
			Debtors:     debtorsForPayment,
			Tags:        tagsForPayment,
		}
		record.ID = payment.ID
		if record.Amount.IsNegative() {
			return fmt.Errorf("payment amount cannot be negative")
		}
		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("create payment %q: %w", payment.Description, err)
		}
		if rebuildDebts {
			if err := applyPaymentDebts(tx, record); err != nil {
				return err
			}
		}
	}
	return nil
}

func createCheckpointSettlements(tx *gorm.DB, settlements []apimodels.CheckpointSettlement, usersByName map[string]dbmodels.UserDBEntry) error {
	for _, settlement := range settlements {
		owedBy, ok := usersByName[strings.TrimSpace(settlement.OwedBy)]
		if !ok {
			return fmt.Errorf("settlement owedBy user %q does not exist", settlement.OwedBy)
		}
		owedTo, ok := usersByName[strings.TrimSpace(settlement.OwedTo)]
		if !ok {
			return fmt.Errorf("settlement owedTo user %q does not exist", settlement.OwedTo)
		}
		if owedBy.ID == owedTo.ID {
			return fmt.Errorf("cannot create settlement to self")
		}
		settlementDate := settlement.Date
		if settlementDate.IsZero() {
			settlementDate = time.Now().UTC()
		}
		record := dbmodels.SettlementDBEntry{
			OwedByID:   owedBy.ID,
			OwedToID:   owedTo.ID,
			Amount:     settlement.Amount,
			Currency:   currency.NormalizeCurrencyCode(settlement.Currency),
			Date:       settlementDate,
			ReversedAt: settlement.ReversedAt,
		}
		record.ID = settlement.ID
		if record.Amount.IsNegative() {
			return fmt.Errorf("settlement amount cannot be negative")
		}
		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("create settlement: %w", err)
		}
	}
	return nil
}

func createCheckpointDebts(tx *gorm.DB, debtRows []apimodels.CheckpointDebt, usersByName map[string]dbmodels.UserDBEntry) error {
	for _, debt := range debtRows {
		owedBy, ok := usersByName[strings.TrimSpace(debt.OwedBy)]
		if !ok {
			return fmt.Errorf("debt owedBy user %q does not exist", debt.OwedBy)
		}
		owedTo, ok := usersByName[strings.TrimSpace(debt.OwedTo)]
		if !ok {
			return fmt.Errorf("debt owedTo user %q does not exist", debt.OwedTo)
		}
		if debt.Amount.IsNegative() {
			return fmt.Errorf("debt amount cannot be negative")
		}
		if err := debts.ApplyNetDebt(tx, owedBy.ID, owedTo.ID, debt.Amount, currency.NormalizeCurrencyCode(debt.Currency)); err != nil {
			return err
		}
		if debt.ID != 0 || !debt.CreatedAt.IsZero() || !debt.UpdatedAt.IsZero() {
			if err := updateImportedDebtMetadata(tx, debt, owedBy.ID, owedTo.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyCheckpointSettlementsToDebts(tx *gorm.DB, settlements []apimodels.CheckpointSettlement, usersByName map[string]dbmodels.UserDBEntry) error {
	for _, settlement := range settlements {
		if settlement.ReversedAt != nil {
			continue
		}
		owedBy := usersByName[strings.TrimSpace(settlement.OwedBy)]
		owedTo := usersByName[strings.TrimSpace(settlement.OwedTo)]
		if err := debts.ApplyNetDebt(tx, owedTo.ID, owedBy.ID, settlement.Amount, currency.NormalizeCurrencyCode(settlement.Currency)); err != nil {
			return err
		}
	}
	return nil
}

func createCheckpointExchangeRates(tx *gorm.DB, exchangeRates []apimodels.Exchange) error {
	for _, rate := range exchangeRates {
		record := dbmodels.ExchangeDBEntry{
			FromCurrency: currency.NormalizeCurrencyCode(rate.FromCurrency),
			ToCurrency:   currency.NormalizeCurrencyCode(rate.ToCurrency),
			Date:         rate.Date,
			Rate:         rate.Rate,
		}
		if record.Date.IsZero() {
			return fmt.Errorf("exchange rate date is required")
		}
		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("create exchange rate: %w", err)
		}
	}
	return nil
}

func updateImportedDebtMetadata(tx *gorm.DB, debt apimodels.CheckpointDebt, owedByID, owedToID uint) error {
	lowID, highID, signedAmount, err := checkpointSignedDebt(owedByID, owedToID, debt.Amount)
	if err != nil {
		return err
	}

	updates := map[string]any{}
	if debt.ID != 0 {
		updates["id"] = debt.ID
	}
	if !debt.CreatedAt.IsZero() {
		updates["created_at"] = debt.CreatedAt
	}
	if !debt.UpdatedAt.IsZero() {
		updates["updated_at"] = debt.UpdatedAt
	}
	if len(updates) == 0 {
		return nil
	}

	result := tx.Model(&dbmodels.DebtDBEntry{}).
		Where("user_low_id = ? AND user_high_id = ? AND currency = ? AND net_amount = ?", lowID, highID, currency.NormalizeCurrencyCode(debt.Currency), signedAmount).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected != 1 {
		return errors.New("failed to update imported debt metadata")
	}
	return nil
}

func checkpointUsersFromNames(usersByName map[string]dbmodels.UserDBEntry, names []string) ([]dbmodels.UserDBEntry, error) {
	users := make([]dbmodels.UserDBEntry, 0, len(names))
	seen := make(map[string]struct{}, len(names))
	for _, rawName := range names {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		user, ok := usersByName[name]
		if !ok {
			return nil, fmt.Errorf("user %q does not exist", name)
		}
		users = append(users, user)
		seen[name] = struct{}{}
	}
	return users, nil
}

func normalizeCheckpointTag(tag string) string {
	return strings.ToLower(strings.TrimSpace(tag))
}

func checkpointSignedDebt(owedBy, owedTo uint, amount decimal.Decimal) (uint, uint, decimal.Decimal, error) {
	if owedBy == owedTo {
		return 0, 0, decimal.Zero, fmt.Errorf("cannot create debt to self")
	}
	if owedBy < owedTo {
		return owedBy, owedTo, amount, nil
	}
	return owedTo, owedBy, amount.Neg(), nil
}

func countCheckpointDebts(checkpoint apimodels.Checkpoint) int {
	if checkpoint.Debts != nil {
		return len(checkpoint.Debts)
	}
	count := 0
	for _, payment := range checkpoint.Payments {
		count += len(payment.Debtors)
	}
	return count
}

func resetCheckpointSequences(tx *gorm.DB) error {
	for _, table := range []string{"users", "tags", "payments", "debts", "settlements", "exchange_rates"} {
		if err := tx.Exec(fmt.Sprintf(
			"SELECT setval(pg_get_serial_sequence('%s', 'id'), COALESCE((SELECT MAX(id) FROM %s), 1), (SELECT COUNT(*) FROM %s) > 0)",
			table,
			table,
			table,
		)).Error; err != nil {
			return fmt.Errorf("reset %s sequence: %w", table, err)
		}
	}
	return nil
}
