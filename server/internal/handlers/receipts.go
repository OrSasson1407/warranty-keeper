package handlers

import (
	"context"
	"fmt"
	"io"
	"log"
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
		log.Printf("OCR parse failed for receipt %s: %v", receipt.ID, err)
	}

	receipt.ImageURL = imageURL
	receipt.RawOCRText = parsed.RawText
	receipt.ParsedVendor = parsed.Vendor
	receipt.ParsedDate = parsed.Date
	receipt.ParsedAmount = parsed.Amount
	receipt.Confidence = parsed.Confidence
	receipt.Status = status
	if err := h.DB.Save(&receipt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save OCR results"})
		return
	}

	draft, err := h.buildDraftResponse(&receipt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve warranty period"})
		return
	}
	c.JSON(http.StatusOK, draft)
}

// buildDraftResponse guesses a category and resolves a warranty expiry from
// a receipt's already-parsed fields, producing the same draft shape whether
// the receipt came from a photo upload (UploadReceipt) or a Gmail-matched
// order email (ListReceipts) -- both route into the same mobile confirm
// screen either way.
func (h *Handler) buildDraftResponse(receipt *models.Receipt) (receiptDraftResponse, error) {
	category := warranty.GuessCategory(receipt.ParsedVendor + " " + receipt.RawOCRText)
	purchaseDate := receipt.CreatedAt
	if receipt.ParsedDate != nil {
		purchaseDate = *receipt.ParsedDate
	}

	resolution, err := warranty.Resolve(h.DB, category, "", purchaseDate)
	if err != nil {
		return receiptDraftResponse{}, err
	}

	var parsedDateStr *string
	if receipt.ParsedDate != nil {
		s := receipt.ParsedDate.Format("2006-01-02")
		parsedDateStr = &s
	}

	return receiptDraftResponse{
		ReceiptID:         receipt.ID.String(),
		ImageURL:          receipt.ImageURL,
		Status:            string(receipt.Status),
		ParsedVendor:      receipt.ParsedVendor,
		ParsedDate:        parsedDateStr,
		ParsedAmount:      receipt.ParsedAmount,
		RawOCRText:        receipt.RawOCRText,
		Confidence:        receipt.Confidence,
		SuggestedCategory: category,
		WarrantyExpiresAt: resolution.ExpiresAt.Format("2006-01-02"),
		WarrantyUncertain: resolution.Uncertain,
	}, nil
}

func (h *Handler) GetReceipt(c *gin.Context) {
	var receipt models.Receipt
	if err := h.DB.Where("id = ? AND household_id = ?", c.Param("id"), householdID(c)).First(&receipt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "receipt not found"})
		return
	}
	c.JSON(http.StatusOK, receipt)
}

// ListReceipts supports the Gmail-import review flow: since Gmail-sourced
// receipts are created asynchronously by the background scan (there's no
// request/response cycle to hand a draft back on, unlike a photo upload),
// the mobile app polls this list to discover ones awaiting confirmation.
// Optional status/source query params narrow the list; both flows share it.
func (h *Handler) ListReceipts(c *gin.Context) {
	query := h.DB.Where("household_id = ?", householdID(c))
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if source := c.Query("source"); source != "" {
		query = query.Where("source = ?", source)
	}

	var receipts []models.Receipt
	if err := query.Order("created_at desc").Find(&receipts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list receipts"})
		return
	}

	drafts := make([]receiptDraftResponse, 0, len(receipts))
	for i := range receipts {
		draft, err := h.buildDraftResponse(&receipts[i])
		if err != nil {
			log.Printf("failed to build draft for receipt %s: %v", receipts[i].ID, err)
			continue
		}
		drafts = append(drafts, draft)
	}
	c.JSON(http.StatusOK, drafts)
}

func fileExtension(filename string) string {
	for i := len(filename) - 1; i >= 0 && i > len(filename)-6; i-- {
		if filename[i] == '.' {
			return filename[i:]
		}
	}
	return ""
}
