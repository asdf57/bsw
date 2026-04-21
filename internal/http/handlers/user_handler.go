package handlers

import (
	"fmt"
	"net/http"

	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
	"github.com/asdf57/bsw/internal/models/mappers"
	"github.com/gin-gonic/gin"
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

func (h *Handlers) DeleteUser(c *gin.Context) {

}

// func (crtl *Controller) GetUserById(c *gin.Context) {
// 	id := c.
// }

// func (crtl *Controller) GetUserByName(c *gin.Context) {

// }
