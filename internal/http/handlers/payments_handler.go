package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/asdf57/bsw/internal/currency"
	"github.com/asdf57/bsw/internal/debts"
	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
	"github.com/asdf57/bsw/internal/models/mappers"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// GetPayment godoc
// @Summary Get a payment by ID
// @Tags payment
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} api.PaymentResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/payment/{id} [get]
func (h *Handlers) GetPayment(c *gin.Context) {
	id := c.Param("id")

	var p dbmodels.PaymentDBEntry
	if err := preloadPayment(h.Db.DB).First(&p, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, mappers.PaymentResponseFromDB(p))
}

// GetPaymentsByTag godoc
// @Summary Get payments by tag
// @Tags payment
// @Produce json
// @Param tag path string true "Payment tag"
// @Success 200 {array} api.PaymentResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/payment/tag/{tag} [get]
func (h *Handlers) GetPaymentsByTag(c *gin.Context) {
	h.getPaymentsByTags(c, []string{c.Param("tag")}, "and")
}

// GetPaymentsByTags godoc
// @Summary Get payments by tags
// @Tags payment
// @Produce json
// @Param tags query []string true "Payment tags" collectionFormat(multi)
// @Param op query string false "Tag operation: and/or" Enums(and, or) default(and)
// @Success 200 {array} api.PaymentResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/payment/tags [get]
func (h *Handlers) GetPaymentsByTags(c *gin.Context) {
	tags := c.QueryArray("tags")
	if len(tags) == 0 {
		tags = sharedCSVQueryValues(c.Query("tags"))
	}
	op := c.DefaultQuery("op", "and")

	h.getPaymentsByTags(c, tags, op)
}

func (h *Handlers) getPaymentsByTags(c *gin.Context, rawTags []string, op string) {
	tags := NormalizeTags(rawTags)
	if len(tags) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provide at least one tag"})
		return
	}

	op = strings.ToLower(strings.TrimSpace(op))
	if op == "" {
		op = "and"
	}
	if op != "and" && op != "or" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "op must be either 'and' or 'or'"})
		return
	}

	subquery := h.Db.DB.Table("payment_tags").
		Select("payment_tags.payment_db_entry_id").
		Joins("JOIN tags ON tags.id = payment_tags.tag_db_entry_id").
		Where("tags.name IN ?", tags)
	if op == "and" {
		subquery = subquery.Group("payment_tags.payment_db_entry_id").
			Having("COUNT(DISTINCT tags.name) = ?", len(tags))
	} else {
		subquery = subquery.Group("payment_tags.payment_db_entry_id")
	}

	var payments []dbmodels.PaymentDBEntry
	if err := preloadPayment(h.Db.DB).Where("payments.id IN (?)", subquery).Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, mappers.PaymentResponsesFromDB(payments))
}

// GetPayments godoc
// @Summary Get all payments
// @Tags payment
// @Produce json
// @Success 200 {array} api.PaymentResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/payment/all [get]
func (h *Handlers) GetAllPayments(c *gin.Context) {
	var p []dbmodels.PaymentDBEntry

	if err := preloadPayment(h.Db.DB).Find(&p).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, mappers.PaymentResponsesFromDB(p))
}

// PostPayment godoc
// @Summary Create a new payment
// @Tags payment
// @Accept json
// @Produce json
// @Param payment body api.Payment true "Payment payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/payment [post]
func (h *Handlers) PostPayment(c *gin.Context) {
	var payment apimodels.Payment

	if err := c.ShouldBind(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payment.Tags = NormalizeTags(payment.Tags)
	if len(payment.Tags) == 0 {
		payment.Tags = []string{"general"}
	}

	var paymentId uint
	payment.Currency = currency.NormalizeCurrencyCode(payment.Currency)
	if strings.TrimSpace(payment.DebtMode) == "" {
		payment.DebtMode = "equal"
	}

	if payment.Amount < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment amount cannot be negative!"})
		return
	}
	if len(payment.Debtors) > 0 && payment.DebtMode != "equal" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "requested debt mode is either invalid or unsupported"})
		return
	}

	payerId, debtors, err := h.validatePaymentUsers(payment)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.Db.DB.Transaction(func(tx *gorm.DB) error {
		tags, err := getExistingTags(tx, payment.Tags)
		if err != nil {
			return err
		}

		parsedAmount := decimal.NewFromFloat(payment.Amount)

		// Store the original payment data in the DB (NOT in baseline currency)
		record := dbmodels.PaymentDBEntry{
			Amount:      parsedAmount,
			Description: payment.Description,
			Date:        payment.Date,
			PayerID:     payerId,
			Debtors:     debtors,
			Currency:    payment.Currency,
			Tags:        tags,
		}

		if err := tx.Create(&record).Error; err != nil {
			return fmt.Errorf("failed to create payment: %s", err.Error())
		}

		paymentId = record.ID

		if err := applyPaymentDebts(tx, record); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create payment: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"info": fmt.Sprintf("payment created with id %d", paymentId)})
}

// PutPayment godoc
// @Summary Update a payment by ID
// @Tags payment
// @Accept json
// @Produce json
// @Param id path int true "Payment ID"
// @Param payment body api.Payment true "Payment payload"
// @Success 200 {object} api.PaymentResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/payment/{id} [put]
func (h *Handlers) PutPayment(c *gin.Context) {
	id := c.Param("id")
	var payment apimodels.Payment
	if err := c.ShouldBind(&payment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payment.Tags = NormalizeTags(payment.Tags)
	if len(payment.Tags) == 0 {
		payment.Tags = []string{"general"}
	}
	if payment.Amount < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment amount cannot be negative!"})
		return
	}

	var updated dbmodels.PaymentDBEntry
	err := h.Db.DB.Transaction(func(tx *gorm.DB) error {
		var existing dbmodels.PaymentDBEntry
		if err := tx.Preload("Debtors").Preload("Tags").First(&existing, id).Error; err != nil {
			return err
		}

		if payment.Payer == "" {
			var payer dbmodels.UserDBEntry
			if err := tx.First(&payer, existing.PayerID).Error; err != nil {
				return err
			}
			payment.Payer = payer.Name
		}
		if payment.Date.IsZero() {
			payment.Date = existing.Date
		}
		payment.Currency = currency.NormalizeCurrencyCode(payment.Currency)
		if payment.Currency == "" {
			payment.Currency = existing.Currency
		}
		if strings.TrimSpace(payment.DebtMode) == "" {
			payment.DebtMode = "equal"
		}
		if len(payment.Debtors) > 0 && payment.DebtMode != "equal" {
			return fmt.Errorf("requested debt mode is either invalid or unsupported")
		}

		payerID, debtors, err := h.validatePaymentUsers(payment)
		if err != nil {
			return err
		}
		tags, err := getExistingTags(tx, payment.Tags)
		if err != nil {
			return err
		}

		if err := reversePaymentDebts(tx, existing); err != nil {
			return err
		}

		existing.Amount = decimal.NewFromFloat(payment.Amount)
		existing.Description = payment.Description
		existing.Date = payment.Date
		existing.PayerID = payerID
		existing.Currency = payment.Currency

		if err := tx.Save(&existing).Error; err != nil {
			return fmt.Errorf("update payment: %w", err)
		}
		if err := tx.Model(&existing).Association("Debtors").Replace(debtors); err != nil {
			return err
		}
		if err := tx.Model(&existing).Association("Tags").Replace(tags); err != nil {
			return err
		}

		existing.Debtors = debtors
		existing.Tags = tags
		if err := applyPaymentDebts(tx, existing); err != nil {
			return err
		}

		return preloadPayment(tx).First(&updated, existing.ID).Error
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mappers.PaymentResponseFromDB(updated))
}

// DeletePayment godoc
// @Summary Delete a payment by ID
// @Tags payment
// @Produce json
// @Param id path int true "Payment ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/payment/{id} [delete]
func (h *Handlers) DeletePayment(c *gin.Context) {
	id := c.Param("id")

	var payment dbmodels.PaymentDBEntry
	err := h.Db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("Debtors").First(&payment, id).Error; err != nil {
			return err
		}

		if err := reversePaymentDebts(tx, payment); err != nil {
			return err
		}

		if err := tx.Model(&payment).Association("Debtors").Clear(); err != nil {
			return err
		}

		if err := tx.Model(&payment).Association("Tags").Clear(); err != nil {
			return err
		}

		if err := tx.Unscoped().Delete(&payment).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to delete payment: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"info": fmt.Sprintf("deleted payment with id %s", id)})
}

func NormalizeTags(raw []string) []string {
	seen := make(map[string]struct{}, len(raw))
	tags := make([]string, 0, len(raw))
	for _, tag := range raw {
		for _, part := range strings.Split(tag, ",") {
			cleaned := strings.ToLower(strings.TrimSpace(part))
			if cleaned == "" {
				continue
			}
			if _, ok := seen[cleaned]; ok {
				continue
			}
			seen[cleaned] = struct{}{}
			tags = append(tags, cleaned)
		}
	}
	return tags
}

func getExistingTags(tx *gorm.DB, names []string) ([]dbmodels.TagDBEntry, error) {
	names = NormalizeTags(names)
	if len(names) == 0 {
		names = []string{"general"}
	}

	var tags []dbmodels.TagDBEntry
	if err := tx.Where("name IN ?", names).Find(&tags).Error; err != nil {
		return nil, fmt.Errorf("get tags: %w", err)
	}

	found := make(map[string]struct{}, len(tags))
	for _, tag := range tags {
		found[tag.Name] = struct{}{}
	}
	for _, name := range names {
		if _, ok := found[name]; !ok {
			return nil, fmt.Errorf("tag %q does not exist; create it first", name)
		}
	}
	return tags, nil
}

func (h *Handlers) validatePaymentUsers(payment apimodels.Payment) (uint, []dbmodels.UserDBEntry, error) {
	payerID, err := h.Db.GetUserIdFromName(payment.Payer)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to obtain payer ID from DB: %w", err)
	}

	debtors, err := h.Db.GetUsersFromNames(payment.Debtors)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to obtain debtor data from DB: %w", err)
	}

	debtorIDsByName := make(map[string]uint, len(debtors))
	for _, debtor := range debtors {
		debtorIDsByName[debtor.Name] = debtor.ID
	}

	for _, debtorName := range payment.Debtors {
		if strings.EqualFold(strings.TrimSpace(debtorName), payment.Payer) {
			return 0, nil, fmt.Errorf("user %s cannot be indebted to themselves", debtorName)
		}
		if _, ok := debtorIDsByName[debtorName]; !ok {
			return 0, nil, fmt.Errorf("no user with name %s could be found", debtorName)
		}
	}

	return payerID, debtors, nil
}

func applyPaymentDebts(tx *gorm.DB, payment dbmodels.PaymentDBEntry) error {
	if len(payment.Debtors) == 0 {
		return nil
	}

	parsedNumDebtors := decimal.NewFromInt(int64(len(payment.Debtors)))
	amountOwedPerDebtor := payment.Amount.Div(parsedNumDebtors.Add(decimal.NewFromInt(1)))
	for _, debtor := range payment.Debtors {
		if debtor.ID == payment.PayerID {
			return fmt.Errorf("user cannot be indebted to themselves")
		}
		if err := debts.ApplyNetDebt(tx, debtor.ID, payment.PayerID, amountOwedPerDebtor, payment.Currency); err != nil {
			return err
		}
	}
	return nil
}

func reversePaymentDebts(tx *gorm.DB, payment dbmodels.PaymentDBEntry) error {
	if len(payment.Debtors) == 0 {
		return nil
	}

	parsedNumDebtors := decimal.NewFromInt(int64(len(payment.Debtors)))
	amountOwedPerDebtor := payment.Amount.Div(parsedNumDebtors.Add(decimal.NewFromInt(1)))
	for _, debtor := range payment.Debtors {
		if err := debts.ApplyNetDebt(tx, payment.PayerID, debtor.ID, amountOwedPerDebtor, payment.Currency); err != nil {
			return err
		}
	}
	return nil
}

func preloadPayment(db *gorm.DB) *gorm.DB {
	return db.Preload("Payer").Preload("Debtors").Preload("Tags")
}

func sharedCSVQueryValues(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.Split(value, ",")
}
