package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
	"github.com/asdf57/bsw/internal/models/mappers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateUser godoc
// @Summary Create a new user
// @Tags user
// @Accept json
// @Produce json
// @Param user body api.User true "User payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/user [post]
func (h *Handlers) CreateUser(c *gin.Context) {
	var user apimodels.User

	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to coerce user request to DB model"})
		return
	}

	record := dbmodels.UserDBEntry{Name: user.Name}
	if err := h.Db.DB.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create user in the DB: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"info": "created user in DB"})
}

// GetUsers godoc
// @Summary Get all users
// @Tags user
// @Produce json
// @Success 200 {array} api.UserSummary
// @Failure 500 {object} map[string]string
// @Router /api/v1/user [get]
func (h *Handlers) GetUsers(c *gin.Context) {
	var users []dbmodels.UserDBEntry

	if err := h.Db.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch users: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, mappers.UserSummariesFromDB(users))
}

// DeleteUser godoc
// @Summary Delete a user by ID
// @Tags user
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/user/{id} [delete]
func (h *Handlers) DeleteUser(c *gin.Context) {
	userIDParam := c.Param("id")
	userID64, err := strconv.ParseUint(userIDParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	userID := uint(userID64)

	// First see if that user even exists
	var user dbmodels.UserDBEntry
	if err := h.Db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user with the provided ID does not exist"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query user"})
		return
	}

	var payerPaymentCount int64
	if err := h.Db.DB.Model(&dbmodels.PaymentDBEntry{}).
		Where("payer_id = ?", userID).
		Count(&payerPaymentCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search for payments for user"})
		return
	}

	if payerPaymentCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "user still has active payments as payer"})
		return
	}

	var debtorPaymentCount int64
	if err := h.Db.DB.Table("debtors").
		Joins("JOIN payments ON payments.id = debtors.payment_db_entry_id").
		Where("debtors.user_db_entry_id = ?", userID).
		Where("payments.deleted_at IS NULL").
		Count(&debtorPaymentCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search for payments for user as debtor"})
		return
	}

	if debtorPaymentCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "user still has active payments as debtor"})
		return
	}

	var activeDebtCount int64
	if err := h.Db.DB.Model(&dbmodels.DebtDBEntry{}).
		Where("user_low_id = ? OR user_high_id = ?", userID, userID).
		Count(&activeDebtCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search for debts for user"})
		return
	}

	if activeDebtCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "user still has active debts"})
		return
	}

	if err := h.Db.DB.Unscoped().Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"info": "user deleted"})
}

// func (crtl *Controller) GetUserById(c *gin.Context) {
// 	id := c.
// }

// func (crtl *Controller) GetUserByName(c *gin.Context) {

// }
