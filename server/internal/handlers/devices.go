package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"warrantykeeper/server/internal/models"
)

type registerDeviceRequest struct {
	ExpoPushToken string `json:"expo_push_token" binding:"required"`
}

// RegisterDevice upserts a push token for the calling user, used by the
// notify-expiring job to reach them.
func (h *Handler) RegisterDevice(c *gin.Context) {
	var req registerDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device := models.DeviceToken{UserID: userID(c), ExpoPushToken: req.ExpoPushToken}
	err := h.DB.Where("expo_push_token = ?", req.ExpoPushToken).
		Assign(models.DeviceToken{UserID: userID(c)}).
		FirstOrCreate(&device).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "registered"})
}
