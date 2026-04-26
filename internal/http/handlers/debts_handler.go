package handlers

import (
	"errors"
	"log"
	"net/http"

	ds "github.com/asdf57/bsw/internal/debts"
	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
	"github.com/asdf57/bsw/internal/models/mappers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetDebts godoc
// @Summary Fetch all debts as From-To user debt pairs
// @Tags debts
// @Produce json
// @Success 200 {array} api.DebtResponse
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/debts [get]
func (h *Handlers) GetDebts(c *gin.Context) {
	var debts []dbmodels.DebtDBEntry

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

	c.JSON(http.StatusOK, mappers.DebtResponsesFromDB(debts, usersByID))
}

// GetAllUserDebts godoc
// @Summary Fetch a map of users to their debts
// @Tags debts
// @Produce json
// @Success 200 {object} map[string][]api.DebtResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/debts/debts/users [get]
func (h *Handlers) GetAllUserDebts(c *gin.Context) {
	var debts []dbmodels.DebtDBEntry
	debtsMap := make(map[string][]apimodels.DebtResponse)

	if err := h.Db.DB.Find(&debts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unhandled error while retrieving debts"})
		return
	}

	usersByID, err := h.getDebtUsersByID(debts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unhandled error while querying for users"})
		return
	}

	for _, debt := range debts {
		owedByUser := usersByID[ds.GetOwedByUser(debt)]
		if owedByUser.ID == 0 {
			continue
		}

		debtsMap[owedByUser.Name] = append(debtsMap[owedByUser.Name], mappers.DebtResponseFromDB(debt, usersByID))
	}

	c.JSON(http.StatusOK, debtsMap)
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
