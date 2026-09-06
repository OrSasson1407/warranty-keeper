package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"warrantykeeper/server/internal/models"
)

type createWarrantyRuleReportRequest struct {
	Note string `json:"note"`
}

// ReportWarrantyRule lets a user flag that a product's resolved warranty
// period looks wrong -- the community-correction half of expanding
// warranty_rules coverage. No automated action: this just queues a row for
// a maintainer to review and fix the underlying rule data manually.
func (h *Handler) ReportWarrantyRule(c *gin.Context) {
	product, ok := h.authorizeProduct(c, c.Param("id"))
	if !ok {
		return
	}

	var req createWarrantyRuleReportRequest
	_ = c.ShouldBindJSON(&req) // note is optional; an empty/missing body is fine

	report := models.WarrantyRuleReport{
		ProductID: product.ID,
		UserID:    userID(c),
		Note:      req.Note,
	}
	if err := h.DB.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save report"})
		return
	}

	c.JSON(http.StatusCreated, report)
}
