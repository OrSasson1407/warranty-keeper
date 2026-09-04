package handlers_test

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"warrantykeeper/server/internal/models"
	"warrantykeeper/server/internal/ocr"
	"warrantykeeper/server/internal/warranty"
)

type receiptDraft struct {
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

func TestUploadReceipt_HighConfidenceResolvesCategoryAndWarranty(t *testing.T) {
	s := newTestSetup(t)
	s.seedRule(t, "מזגן", "", 24)

	purchaseDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	amount := 3200.0
	s.ocrProvider.result = ocr.ParsedReceipt{
		Vendor:     "חנות מזגנים בע\"מ",
		RawText:    "קבלה עבור מזגן טורנדו",
		Date:       &purchaseDate,
		Amount:     &amount,
		Confidence: 0.92,
	}

	rec := doMultipartAs(t, s.router, http.MethodPost, "/receipts", s.token, "image", "receipt.jpg", []byte("fake-bytes"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (body: %s)", rec.Code, http.StatusOK, rec.Body.String())
	}

	var draft receiptDraft
	decodeJSON(t, rec, &draft)

	if draft.Status != "processed" {
		t.Errorf("Status = %q, want %q", draft.Status, "processed")
	}
	if draft.SuggestedCategory != "מזגן" {
		t.Errorf("SuggestedCategory = %q, want %q", draft.SuggestedCategory, "מזגן")
	}
	if draft.WarrantyUncertain {
		t.Error("WarrantyUncertain = true, want false when a rule matched the guessed category")
	}
	wantExpiry := purchaseDate.AddDate(0, 24, 0).Format("2006-01-02")
	if draft.WarrantyExpiresAt != wantExpiry {
		t.Errorf("WarrantyExpiresAt = %q, want %q", draft.WarrantyExpiresAt, wantExpiry)
	}
	if draft.ParsedAmount == nil || *draft.ParsedAmount != amount {
		t.Errorf("ParsedAmount = %v, want %v", draft.ParsedAmount, amount)
	}
	if draft.ImageURL == "" {
		t.Error("expected a non-empty image URL from storage")
	}

	var stored models.Receipt
	if err := s.db.First(&stored, "id = ?", draft.ReceiptID).Error; err != nil {
		t.Fatalf("expected a receipt row to be persisted: %v", err)
	}
	if stored.Status != models.ReceiptStatusProcessed {
		t.Errorf("stored receipt status = %q, want %q", stored.Status, models.ReceiptStatusProcessed)
	}
	if stored.HouseholdID != s.householdID {
		t.Errorf("stored receipt household_id = %v, want %v", stored.HouseholdID, s.householdID)
	}
}

func TestUploadReceipt_NoOCRMatchFallsBackAndUncertain(t *testing.T) {
	s := newTestSetup(t)
	// No rules seeded; OCR returns nothing recognizable (mirrors the stub provider).
	s.ocrProvider.result = ocr.ParsedReceipt{Confidence: 0}

	rec := doMultipartAs(t, s.router, http.MethodPost, "/receipts", s.token, "image", "receipt.jpg", []byte("fake-bytes"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var draft receiptDraft
	decodeJSON(t, rec, &draft)

	if draft.SuggestedCategory != "" {
		t.Errorf("SuggestedCategory = %q, want empty (nothing for the client to prefill)", draft.SuggestedCategory)
	}
	if !draft.WarrantyUncertain {
		t.Error("WarrantyUncertain = false, want true")
	}
	wantExpiry := time.Now().AddDate(0, warranty.DefaultFallbackMonths, 0).Format("2006-01-02")
	if draft.WarrantyExpiresAt != wantExpiry {
		t.Errorf("WarrantyExpiresAt = %q, want %q (today + fallback months, since OCR gave no date)", draft.WarrantyExpiresAt, wantExpiry)
	}
}

func TestUploadReceipt_OCRErrorStillReturnsUsableDraft(t *testing.T) {
	s := newTestSetup(t)
	s.ocrProvider.err = errors.New("ocr provider unavailable")

	rec := doMultipartAs(t, s.router, http.MethodPost, "/receipts", s.token, "image", "receipt.jpg", []byte("fake-bytes"))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d — an OCR failure should still hand back a draft for manual entry", rec.Code, http.StatusOK)
	}
	var draft receiptDraft
	decodeJSON(t, rec, &draft)
	if draft.Status != "failed" {
		t.Errorf("Status = %q, want %q", draft.Status, "failed")
	}

	var stored models.Receipt
	if err := s.db.First(&stored, "id = ?", draft.ReceiptID).Error; err != nil {
		t.Fatalf("expected the receipt row to still exist: %v", err)
	}
	if stored.Status != models.ReceiptStatusFailed {
		t.Errorf("stored receipt status = %q, want %q", stored.Status, models.ReceiptStatusFailed)
	}
}

func TestUploadReceipt_StorageFailureReturns500(t *testing.T) {
	s := newTestSetup(t)
	s.storage.err = errors.New("bucket unavailable")

	rec := doMultipartAs(t, s.router, http.MethodPost, "/receipts", s.token, "image", "receipt.jpg", []byte("fake-bytes"))
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestUploadReceipt_MissingFileReturns400(t *testing.T) {
	s := newTestSetup(t)
	rec := doMultipartNoFileAs(t, s.router, http.MethodPost, "/receipts", s.token)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestUploadReceipt_RequiresAuth(t *testing.T) {
	s := newTestSetup(t)
	rec := doMultipartAs(t, s.router, http.MethodPost, "/receipts", "", "image", "receipt.jpg", []byte("fake-bytes"))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestGetReceipt_ScopedToHousehold(t *testing.T) {
	s := newTestSetup(t)
	receipt := models.Receipt{HouseholdID: s.householdID, ImageURL: "https://fake-storage.test/x.jpg"}
	if err := s.db.Create(&receipt).Error; err != nil {
		t.Fatalf("failed to seed receipt: %v", err)
	}

	rec := doJSONAs(t, s.router, http.MethodGet, "/receipts/"+receipt.ID.String(), s.token, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	otherToken, _ := s.createOtherHousehold(t)
	rec2 := doJSONAs(t, s.router, http.MethodGet, "/receipts/"+receipt.ID.String(), otherToken, nil)
	if rec2.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d for a different household", rec2.Code, http.StatusNotFound)
	}
}
