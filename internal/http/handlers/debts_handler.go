package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/asdf57/bsw/internal/currency"
	ds "github.com/asdf57/bsw/internal/debts"
	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
	"github.com/asdf57/bsw/internal/models/mappers"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// GetDebts godoc
// @Summary Fetch all debts as From-To user debt pairs
// @Tags debts
// @Produce json
// @Param currency query string false "Currency for generated debts" default(USD)
// @Success 200 {array} api.DebtResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/debts [get]
func (h *Handlers) GetDebts(c *gin.Context) {
	var debts []dbmodels.DebtDBEntry
	targetCurrency := currency.NormalizeCurrencyCode(c.DefaultQuery("currency", "USD"))

	if err := h.Db.DB.Order("id").Find(&debts).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("error: could not find debt record in database: %s", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "debt record not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "unhandled error while retrieving debts"})
		return
	}

	usersByID, err := h.getDebtUsersByID(debts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unhandled error while querying for users"})
		return
	}

	responses, err := h.convertDebtResponses(debts, usersByID, targetCurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to convert debts: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, responses)
}

// GetAllUserDebts godoc
// @Summary Fetch a map of users to their debts
// @Tags debts
// @Produce json
// @Param currency query string false "Currency for generated debts" default(USD)
// @Success 200 {object} map[string][]api.DebtResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/debts/users [get]
func (h *Handlers) GetAllUserDebts(c *gin.Context) {
	var debts []dbmodels.DebtDBEntry
	debtsMap := make(map[string][]apimodels.DebtResponse)
	targetCurrency := currency.NormalizeCurrencyCode(c.DefaultQuery("currency", "USD"))

	if err := h.Db.DB.Find(&debts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unhandled error while retrieving debts"})
		return
	}

	usersByID, err := h.getDebtUsersByID(debts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unhandled error while querying for users"})
		return
	}

	responses, err := h.convertDebtResponses(debts, usersByID, targetCurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to convert debts: %s", err.Error())})
		return
	}

	for _, debt := range responses {
		owedByUser := debt.OwedByUser
		if owedByUser.ID == 0 {
			continue
		}

		debtsMap[owedByUser.Name] = append(debtsMap[owedByUser.Name], debt)
	}

	c.JSON(http.StatusOK, debtsMap)
}

// SettleDebts godoc
// @Summary Settle debts for a user
// @Tags debts
// @Accept json
// @Produce json
// @Param settle body api.SettleDebtsRequest true "Settle debts payload"
// @Success 200 {object} api.SettleDebtsResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/debts/settle [post]
func (h *Handlers) SettleDebts(c *gin.Context) {
	var settlePayload apimodels.SettleDebtsRequest

	if err := c.ShouldBind(&settlePayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	owedByName := strings.TrimSpace(settlePayload.OwedBy)
	owedToName := strings.TrimSpace(settlePayload.OwedTo)
	if owedByName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owedBy is required"})
		return
	}

	owedByUser, err := h.Db.GetUserFromName(owedByName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	owedByID := owedByUser.ID

	var owedToID *uint
	var owedToUser *dbmodels.UserDBEntry
	if owedToName != "" {
		user, err := h.Db.GetUserFromName(owedToName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		owedToUser = user
		owedToID = &user.ID
	}
	if settlePayload.Amount != nil && owedToID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "owedTo is required when amount is provided"})
		return
	}
	if settlePayload.Amount != nil && settlePayload.Amount.LessThan(decimal.NewFromFloat(0.01)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "settlement amount must be at least 0.01"})
		return
	}

	var settlements []dbmodels.SettlementDBEntry
	err = h.Db.DB.Transaction(func(tx *gorm.DB) error {
		debtsToSettle, err := ds.DebtsToSettle(tx, owedByID, owedToID)
		if err != nil {
			return err
		}
		if settlePayload.Amount != nil && len(debtsToSettle) != 1 {
			return fmt.Errorf("no debt from %s to %s to settle", owedByName, owedToName)
		}

		settlements = make([]dbmodels.SettlementDBEntry, 0, len(debtsToSettle))
		for _, debt := range debtsToSettle {
			owedToIDForDebt := ds.GetOwedToUser(debt)
			owedToForDebt := owedToUser
			if owedToForDebt == nil || owedToForDebt.ID != owedToIDForDebt {
				user := dbmodels.UserDBEntry{}
				if err := tx.First(&user, owedToIDForDebt).Error; err != nil {
					return err
				}
				owedToForDebt = &user
			}
			amount := debt.NetAmount.Abs()
			if settlePayload.Amount != nil {
				if settlePayload.Amount.GreaterThan(amount) {
					return fmt.Errorf("settlement amount cannot exceed %s", amount.StringFixed(2))
				}
				amount = *settlePayload.Amount
			}

			settlements = append(settlements, dbmodels.SettlementDBEntry{
				OwedByID: owedByID,
				OwedToID: owedToIDForDebt,
				Amount:   amount,
				Currency: debt.Currency,
				Date:     time.Now().UTC(),
			})

			settlements[len(settlements)-1].OwedBy = *owedByUser
			settlements[len(settlements)-1].OwedTo = *owedToForDebt
		}

		if len(settlements) > 0 {
			rows := make([]dbmodels.SettlementDBEntry, 0, len(settlements))
			for _, settlement := range settlements {
				rows = append(rows, dbmodels.SettlementDBEntry{
					OwedByID: settlement.OwedByID,
					OwedToID: settlement.OwedToID,
					Amount:   settlement.Amount,
					Currency: settlement.Currency,
					Date:     settlement.Date,
				})
			}
			if err := tx.Create(&rows).Error; err != nil {
				return fmt.Errorf("create settlements: %w", err)
			}
			for idx := range settlements {
				settlements[idx].Model = rows[idx].Model
			}
		}

		if settlePayload.Amount != nil {
			return ds.SettleDebtAmount(tx, debtsToSettle[0], *settlePayload.Amount)
		}

		_, err = ds.SettleDebts(tx, owedByID, owedToID)
		return err
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, apimodels.SettleDebtsResponse{
		OwedBy:       owedByName,
		OwedTo:       owedToName,
		SettledCount: int64(len(settlements)),
		Settlements:  mappers.SettlementResponsesFromDB(settlements),
	})
}

// GetSettlements godoc
// @Summary Fetch settlement history
// @Tags debts
// @Produce json
// @Success 200 {array} api.SettlementResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/debts/settlements [get]
func (h *Handlers) GetSettlements(c *gin.Context) {
	var settlements []dbmodels.SettlementDBEntry
	if err := h.Db.DB.Preload("OwedBy").Preload("OwedTo").Order("date DESC, id DESC").Find(&settlements).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch settlements"})
		return
	}

	c.JSON(http.StatusOK, mappers.SettlementResponsesFromDB(settlements))
}

// ReverseSettlement godoc
// @Summary Reverse a settlement by ID
// @Tags debts
// @Produce json
// @Param id path int true "Settlement ID"
// @Success 200 {object} api.SettlementResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/debts/settlements/{id}/reverse [post]
func (h *Handlers) ReverseSettlement(c *gin.Context) {
	id := c.Param("id")

	var settlement dbmodels.SettlementDBEntry
	err := h.Db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Preload("OwedBy").Preload("OwedTo").First(&settlement, id).Error; err != nil {
			return err
		}
		if settlement.ReversedAt != nil {
			return fmt.Errorf("settlement already reversed")
		}

		if err := ds.ApplyNetDebt(tx, settlement.OwedByID, settlement.OwedToID, settlement.Amount, settlement.Currency); err != nil {
			return err
		}

		now := time.Now().UTC()
		settlement.ReversedAt = &now
		if err := tx.Model(&settlement).Update("reversed_at", now).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "settlement not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, mappers.SettlementResponseFromDB(settlement))
}

func (h *Handlers) getDebtUsersByID(debts []dbmodels.DebtDBEntry) (map[uint]dbmodels.UserDBEntry, error) {
	idSet := make(map[uint]struct{})
	for _, debt := range debts {
		idSet[debt.UserLowId] = struct{}{}
		idSet[debt.UserHighId] = struct{}{}
	}

	ids := make([]uint, 0, len(idSet))
	for id := range idSet {
		ids = append(ids, id)
	}

	usersByID := make(map[uint]dbmodels.UserDBEntry, len(ids))
	if len(ids) == 0 {
		return usersByID, nil
	}

	var users []dbmodels.UserDBEntry
	if err := h.Db.DB.Where("id IN ?", ids).Find(&users).Error; err != nil {
		return nil, err
	}

	for _, user := range users {
		usersByID[user.ID] = user
	}

	return usersByID, nil
}

func (h *Handlers) convertDebtResponses(debts []dbmodels.DebtDBEntry, usersByID map[uint]dbmodels.UserDBEntry, targetCurrency string) ([]apimodels.DebtResponse, error) {
	type pairKey struct {
		low  uint
		high uint
	}

	netByPair := make(map[pairKey]decimal.Decimal)
	for _, debt := range debts {
		converted, err := currency.ConvertCurrencyWithCache(h.Db.DB, debt.NetAmount.Abs(), debt.Currency, targetCurrency, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		if debt.NetAmount.IsNegative() {
			converted = converted.Neg()
		}
		key := pairKey{low: debt.UserLowId, high: debt.UserHighId}
		netByPair[key] = netByPair[key].Add(converted)
	}

	responses := make([]apimodels.DebtResponse, 0, len(netByPair))
	for key, amount := range netByPair {
		if amount.IsZero() {
			continue
		}
		debt := dbmodels.DebtDBEntry{
			UserLowId:  key.low,
			UserHighId: key.high,
			NetAmount:  amount,
			Currency:   targetCurrency,
		}
		responses = append(responses, mappers.DebtResponseFromDB(debt, usersByID))
	}

	return responses, nil
}
