package handlers

import (
	"fmt"
	"net/http"
	"strings"

	apimodels "github.com/asdf57/bsw/internal/models/api"
	dbmodels "github.com/asdf57/bsw/internal/models/db"
	"github.com/gin-gonic/gin"
)

// GetTags godoc
// @Summary Get all payment tags
// @Tags tag
// @Produce json
// @Success 200 {array} api.TagResponse
// @Failure 500 {object} map[string]string
// @Router /api/v1/tags [get]
func (h *Handlers) GetTags(c *gin.Context) {
	var tags []dbmodels.TagDBEntry
	if err := h.Db.DB.Order("name ASC").Find(&tags).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tags"})
		return
	}

	response := make([]gin.H, 0, len(tags))
	for _, tag := range tags {
		response = append(response, gin.H{"id": tag.ID, "tag": tag.Name})
	}
	c.JSON(http.StatusOK, response)
}

// CreateTag godoc
// @Summary Create a payment tag
// @Tags tag
// @Accept json
// @Produce json
// @Param tag body api.Tag true "Tag payload"
// @Success 200 {object} api.TagResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/tags [post]
func (h *Handlers) CreateTag(c *gin.Context) {
	var tag apimodels.Tag
	if err := c.ShouldBind(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse tag request"})
		return
	}

	name := strings.ToLower(strings.TrimSpace(tag.Name))
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tag is required"})
		return
	}

	record := dbmodels.TagDBEntry{Name: name}
	if err := h.Db.DB.Where("name = ?", name).FirstOrCreate(&record, dbmodels.TagDBEntry{Name: name}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create tag: %s", err.Error())})
		return
	}

	c.JSON(http.StatusOK, apimodels.TagResponse{ID: record.ID, Name: record.Name})
}
