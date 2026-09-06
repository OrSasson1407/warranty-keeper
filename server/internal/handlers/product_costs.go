package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"warrantykeeper/server/internal/models"
)

type createProductCostRequest struct {
	Amount      float64 `json:"amount" binding:"required"`
	Description string  `json:"description"`
	IncurredAt  string  `json:"incurred_at"` // YYYY-MM-DD, defaults to today if omitted
}

// CreateProductCost logs a post-purchase cost (repair, accessory, ...)
// against a product, for the basic TCO view alongside its purchase price.
func (h *Handler) CreateProductCost(c *gin.Context) {
	product, ok := h.authorizeProduct(c, c.Param("id"))
	if !ok {
		return
	}

	var req createProductCostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	incurredAt := time.Now()
	if req.IncurredAt != "" {
		parsed, err := time.Parse(dateLayout, req.IncurredAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "incurred_at must be YYYY-MM-DD"})
			return
		}
		incurredAt = parsed
	}

	cost := models.ProductCost{
		ProductID:   product.ID,
		Amount:      req.Amount,
		Description: req.Description,
		IncurredAt:  incurredAt,
	}
	if err := h.DB.Create(&cost).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save cost"})
		return
	}

	c.JSON(http.StatusCreated, cost)
}

// ListProductCosts returns every logged cost for a product, soonest-incurred
// first, so the client can sum them with the purchase price for a running
// total-cost-of-ownership figure.
func (h *Handler) ListProductCosts(c *gin.Context) {
	product, ok := h.authorizeProduct(c, c.Param("id"))
	if !ok {
		return
	}

	var costs []models.ProductCost
	if err := h.DB.Where("product_id = ?", product.ID).Order("incurred_at DESC").Find(&costs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load costs"})
		return
	}

	c.JSON(http.StatusOK, costs)
}
