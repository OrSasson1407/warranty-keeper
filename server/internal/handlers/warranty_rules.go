package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"warrantykeeper/server/internal/warranty"
)

// ResolveWarranty lets the mobile confirmation screen recompute the
// suggested expiry date live as the user edits category/brand, without
// re-uploading the receipt image.
func (h *Handler) ResolveWarranty(c *gin.Context) {
	category := c.Query("category")
	brand := c.Query("brand")
	purchaseDate := time.Now()
	if raw := c.Query("purchase_date"); raw != "" {
		parsed, err := time.Parse("2006-01-02", raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "purchase_date must be YYYY-MM-DD"})
			return
		}
		purchaseDate = parsed
	}

	resolution, err := warranty.Resolve(h.DB, category, brand, purchaseDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve warranty period"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"warranty_expires_at": resolution.ExpiresAt.Format("2006-01-02"),
		"duration_months":     resolution.DurationMonths,
		"uncertain":           resolution.Uncertain,
		"source":              resolution.Source,
	})
}
