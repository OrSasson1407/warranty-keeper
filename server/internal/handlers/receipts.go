package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"warrantykeeper/server/internal/models"
	"warrantykeeper/server/internal/warranty"
)

// ocrTimeout bounds how long a single OCR call can block the upload request.
// The stub returns instantly, but a real provider (see internal/ocr.AnthropicProvider)
// makes a real network call -- without a bound, a slow or hanging provider
// would hang the HTTP request indefinitely instead of falling back to manual
// entry the way a fast provider error already does.
const ocrTimeout = 20 * time.Second

type receiptDraftResponse struct {
	ReceiptID         string   `json:"receipt_id"`
	ImageURL          string   `json:"image_url"`
	Status            string   `json:"status"`
	ParsedVendor      string   `json:"parsed_vendor"`
	ParsedDate        *string  `json:"parsed_date"`
	ParsedAmount      *float64 `json:"parsed_amount"`
	RawOCRText        string   `json:"raw_ocr_text"`
	Confidence        float64  `json:"confidence"`
	SuggestedCategory string   `json:"suggested_category"`
	WarrantyExpiresAt string   `json:"warranty_expires_at"`
	WarrantyUncertain bool     `json:"warranty_uncertain"`
}

// UploadReceipt implements the receipt processing flow from the architecture
// doc: save the image, run it through OCR (stubbed for now), guess a
// category by keyword, resolve a warranty expiry, and return a draft for
// the mobile confirmation screen. Nothing is persisted as a Product yet —
// that happens on explicit confirm via POST /products.
func (h *Handler) UploadReceipt(c *gin.Context) {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'image' file"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read uploaded file"})
		return
	}
	defer file.Close()

	imageBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read uploaded file"})
		return
	}

	hid := householdID(c)
	receipt := models.Receipt{HouseholdID: hid, Status: models.ReceiptStatusPending}
	if err := h.DB.Create(&receipt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create receipt record"})
		return
	}

	key := fmt.Sprintf("%s%s", receipt.ID.String(), fileExtension(fileHeader.Filename))
	imageURL, err := h.Storage.Upload(c.Request.Context(), key, imageBytes, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		h.DB.Model(&receipt).Update("status", models.ReceiptStatusFailed)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to store receipt image"})
		return
	}

	ocrCtx, cancel := context.WithTimeout(c.Request.Context(), ocrTimeout)
	defer cancel()
	parsed, err := h.OCR.Parse(ocrCtx, imageBytes)
	status := models.ReceiptStatusProcessed
	if err != nil {
		status = models.ReceiptStatusFailed
	}

	receipt.ImageURL = imageURL
	receipt.RawOCRText = parsed.RawText
	receipt.ParsedVendor = parsed.Vendor
	receipt.ParsedDate = parsed.Date
	receipt.ParsedAmount = parsed.Amount
	receipt.Status = status
	if err := h.DB.Save(&receipt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save OCR results"})
		return
	}

	category := warranty.GuessCategory(parsed.Vendor + " " + parsed.RawText)
	purchaseDate := time.Now()
	if parsed.Date != nil {
		purchaseDate = *parsed.Date
	}

	resolution, err := warranty.Resolve(h.DB, category, "", purchaseDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve warranty period"})
		return
	}

	var parsedDateStr *string
	if parsed.Date != nil {
		s := parsed.Date.Format("2006-01-02")
		parsedDateStr = &s
	}

	c.JSON(http.StatusOK, receiptDraftResponse{
		ReceiptID:         receipt.ID.String(),
		ImageURL:          imageURL,
		Status:            string(status),
		ParsedVendor:      parsed.Vendor,
		ParsedDate:        parsedDateStr,
		ParsedAmount:      parsed.Amount,
		RawOCRText:        parsed.RawText,
		Confidence:        parsed.Confidence,
		SuggestedCategory: category,
		WarrantyExpiresAt: resolution.ExpiresAt.Format("2006-01-02"),
		WarrantyUncertain: resolution.Uncertain,
	})
}

func (h *Handler) GetReceipt(c *gin.Context) {
	var receipt models.Receipt
	if err := h.DB.Where("id = ? AND household_id = ?", c.Param("id"), householdID(c)).First(&receipt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "receipt not found"})
		return
	}
	c.JSON(http.StatusOK, receipt)
}

func fileExtension(filename string) string {
	for i := len(filename) - 1; i >= 0 && i > len(filename)-6; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ""
}
