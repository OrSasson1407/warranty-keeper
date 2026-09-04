package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"warrantykeeper/server/internal/models"
	"warrantykeeper/server/internal/warranty"
)

const dateLayout = "2006-01-02"

type createProductRequest struct {
	Name              string   `json:"name" binding:"required"`
	Category          string   `json:"category" binding:"required"`
	Brand             string   `json:"brand"`
	PurchaseDate      string   `json:"purchase_date" binding:"required"`
	Price             *float64 `json:"price"`
	Room              string   `json:"room"`
	PhotoURL          string   `json:"photo_url"`
	ReceiptID         *string  `json:"receipt_id"`
	WarrantyExpiresAt *string  `json:"warranty_expires_at"` // manual override, e.g. extended warranty
}

// CreateProduct confirms a receipt draft (or a fully manual entry) into a
// saved Product, resolving the warranty expiry via the rules engine unless
// the caller supplies an explicit override.
func (h *Handler) CreateProduct(c *gin.Context) {
	var req createProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	purchaseDate, err := time.Parse(dateLayout, req.PurchaseDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "purchase_date must be YYYY-MM-DD"})
		return
	}

	hid := householdID(c)

	product := models.Product{
		HouseholdID:  hid,
		Name:         req.Name,
		Category:     req.Category,
		Brand:        req.Brand,
		PurchaseDate: purchaseDate,
		Price:        req.Price,
		Room:         req.Room,
		PhotoURL:     req.PhotoURL,
	}

	if req.ReceiptID != nil && *req.ReceiptID != "" {
		receiptID, err := uuid.Parse(*req.ReceiptID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid receipt_id"})
			return
		}
		var receipt models.Receipt
		if err := h.DB.Where("id = ? AND household_id = ?", receiptID, hid).First(&receipt).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "receipt not found for this household"})
			return
		}
		if product.PhotoURL == "" {
			product.PhotoURL = receipt.ImageURL
		}
		product.ReceiptID = &receiptID
	}

	if req.WarrantyExpiresAt != nil && *req.WarrantyExpiresAt != "" {
		expiresAt, err := time.Parse(dateLayout, *req.WarrantyExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "warranty_expires_at must be YYYY-MM-DD"})
			return
		}
		product.WarrantyExpiresAt = expiresAt
		product.WarrantyUncertain = false
	} else {
		resolution, err := warranty.Resolve(h.DB, req.Category, req.Brand, purchaseDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve warranty period"})
			return
		}
		product.WarrantyExpiresAt = resolution.ExpiresAt
		product.WarrantyUncertain = resolution.Uncertain
	}

	if err := h.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

// ListProducts returns the household's products sorted soonest-to-expire
// first (Dashboard default per the UX doc), with optional free-text search.
func (h *Handler) ListProducts(c *gin.Context) {
	var products []models.Product
	query := h.DB.Where("household_id = ?", householdID(c))

	if q := c.Query("q"); q != "" {
		query = query.Where("LOWER(name) LIKE LOWER(?)", "%"+q+"%")
	}

	if err := query.Order("warranty_expires_at ASC").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func (h *Handler) GetProduct(c *gin.Context) {
	var product models.Product
	if err := h.DB.Where("id = ? AND household_id = ?", c.Param("id"), householdID(c)).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	c.JSON(http.StatusOK, product)
}

type updateProductRequest struct {
	Name              *string  `json:"name"`
	Category          *string  `json:"category"`
	Brand             *string  `json:"brand"`
	Price             *float64 `json:"price"`
	Room              *string  `json:"room"`
	PhotoURL          *string  `json:"photo_url"`
	WarrantyExpiresAt *string  `json:"warranty_expires_at"`
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	var product models.Product
	if err := h.DB.Where("id = ? AND household_id = ?", c.Param("id"), householdID(c)).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	var req updateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Category != nil {
		product.Category = *req.Category
	}
	if req.Brand != nil {
		product.Brand = *req.Brand
	}
	if req.Price != nil {
		product.Price = req.Price
	}
	if req.Room != nil {
		product.Room = *req.Room
	}
	if req.PhotoURL != nil {
		product.PhotoURL = *req.PhotoURL
	}
	if req.WarrantyExpiresAt != nil {
		expiresAt, err := time.Parse(dateLayout, *req.WarrantyExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "warranty_expires_at must be YYYY-MM-DD"})
			return
		}
		product.WarrantyExpiresAt = expiresAt
		product.WarrantyUncertain = false
	}

	if err := h.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update product"})
		return
	}

	c.JSON(http.StatusOK, product)
}
