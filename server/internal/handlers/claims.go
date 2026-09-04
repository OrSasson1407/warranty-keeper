package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"warrantykeeper/server/internal/models"
)

type createClaimRequest struct {
	IssueDescription string `json:"issue_description" binding:"required"`
}

// authorizeProduct loads a product and checks it belongs to the caller's
// household, writing an error response and returning ok=false if not.
func (h *Handler) authorizeProduct(c *gin.Context, productID string) (models.Product, bool) {
	var product models.Product
	if err := h.DB.Where("id = ? AND household_id = ?", productID, householdID(c)).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return product, false
	}
	return product, true
}

func (h *Handler) CreateClaim(c *gin.Context) {
	product, ok := h.authorizeProduct(c, c.Param("id"))
	if !ok {
		return
	}

	var req createClaimRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	claim := models.WarrantyClaim{
		ProductID:        product.ID,
		IssueDescription: req.IssueDescription,
		Status:           models.ClaimStatusOpen,
	}
	if err := h.DB.Create(&claim).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save claim"})
		return
	}

	c.JSON(http.StatusCreated, claim)
}

func (h *Handler) ListClaims(c *gin.Context) {
	product, ok := h.authorizeProduct(c, c.Param("id"))
	if !ok {
		return
	}

	var claims []models.WarrantyClaim
	if err := h.DB.Where("product_id = ?", product.ID).Order("created_at DESC").Find(&claims).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load claims"})
		return
	}

	c.JSON(http.StatusOK, claims)
}
