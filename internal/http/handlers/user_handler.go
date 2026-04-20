package handlers

import (
	"fmt"
	"net/http"

	"github.com/asdf57/bsw/internal/models"
	"github.com/gin-gonic/gin"
)

// CreateUser godoc
// @Summary Create a new user
// @Tags user
// @Accept json
// @Produce json
// @Param user body models.User true "User payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/user [post]
func (h *Handlers) CreateUser(c *gin.Context) {
	var user models.User

	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to coerce user request to DB model"})
		return
	}

	if err := h.Db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create user in the DB: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, gin.H{"info": "created user in DB"})
}

// GetUsers godoc
// @Summary Get all users
// @Tags user
// @Produce json
// @Success 200 {array} models.User
// @Failure 500 {object} map[string]string
// @Router /api/v1/user [get]
func (h *Handlers) GetUsers(c *gin.Context) {
	var users []models.UserDBEntry

	if err := h.Db.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to fetch users: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, users)
}

func (h *Handlers) DeleteUser(c *gin.Context) {

}

// func (crtl *Controller) GetUserById(c *gin.Context) {
// 	id := c.
// }

// func (crtl *Controller) GetUserByName(c *gin.Context) {

// }
