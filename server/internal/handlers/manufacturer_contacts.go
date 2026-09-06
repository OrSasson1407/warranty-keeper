package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"warrantykeeper/server/internal/models"
)

// ListManufacturerContacts returns every known manufacturer contact. The
// dataset is small (dozens of brands, not thousands), so the app fetches
// and caches the whole list rather than looking up one brand at a time.
func (h *Handler) ListManufacturerContacts(c *gin.Context) {
	var contacts []models.ManufacturerContact
	if err := h.DB.Order("brand").Find(&contacts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load manufacturer contacts"})
		return
	}
	c.JSON(http.StatusOK, contacts)
}
