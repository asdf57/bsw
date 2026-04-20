package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/asdf57/bsw/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// GetDebts godoc
// @Summary Fetch all debts as From-To user debt pairs
// @Tags debts
// @Produce json
// @Success 200 {array} models.DebtDBEntry
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/debts [get]
func (h *Handlers) GetDebts(c *gin.Context) {
	var debts []models.DebtDBEntry

	if err := h.Db.DB.Order("id").Find(&debts).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("error: could not find debt record in database: %s", err)
			c.JSON(http.StatusNotFound, gin.H{"error": "debt record not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "unhandled error while retrieving debts"})
		return
	}

	c.JSON(http.StatusOK, debts)
}

// GetAllUserDebts godoc
// @Summary Fetch a map of users to their debts
// @Tags debts
// @Produce json
// @Success 200 {object} map[string][]models.DebtDBEntry
// @Failure 500 {object} map[string]string
// @Router /api/v1/debts/debts/users [get]
func (h *Handlers) GetAllUserDebts(c *gin.Context) {
	var debts []models.DebtDBEntry
	debtsMap := make(map[string][]models.DebtDBEntry)

	if err := h.Db.DB.Find(&debts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unhandled error while retrieving debts"})
		return
	}

	// Get set of ids for uniqueness
	idSet := make(map[uint]struct{})
	for _, d := range debts {
		idSet[d.OwedByUserId] = struct{}{}
	}

	// Store all unique ids in a list
	ids := make([]uint, 0, len(idSet))
	for id, _ := range idSet {
		ids = append(ids, id)
	}

	var users []models.UserDBEntry
	if err := h.Db.DB.Where("id IN ?", ids).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unhandled error while querying for users"})
		return
	}

	for _, user := range users {
		for _, debt := range debts {
			if debt.OwedByUserId == user.ID {
				debtsMap[user.Name] = append(debtsMap[user.Name], debt)
			}
		}
	}

	c.JSON(http.StatusOK, debts)
}

func (h *Handlers) AddDebt(c *gin.Context) {
	var debt models.DebtEntry

	if err := c.ShouldBind(&debt); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Users cannot owe debts to themselves
	if debt.OwedByUser == debt.OwedToUser {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("user %s cannot be indebted to themselves", debt.OwedByUser)})
		return
	}

	// optimization -- can probably batch all user id calls into a single query!
	owedByUserId, err := h.Db.GetUserIdFromName(debt.OwedByUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("no user with name %s could be found", debt.OwedByUser)})
		return
	}

	owedToUserId, err := h.Db.GetUserIdFromName(debt.OwedToUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("no user with name %s could be found", debt.OwedToUser)})
		return
	}

	parsedAmount := decimal.NewFromFloat(debt.Amount)

	debtEntry := models.DebtDBEntry{
		OwedByUserId: owedByUserId,
		OwedToUserId: owedToUserId,
		Amount:       parsedAmount,
		Currency:     debt.Currency, // debts are always stored in the
	}

	if err := h.Db.DB.Create(&debtEntry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create debt entry in the db: %s", err.Error())})
		return
	}
}
