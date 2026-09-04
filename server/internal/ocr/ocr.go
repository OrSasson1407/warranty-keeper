package ocr

import (
	"context"
	"time"
)

// ParsedReceipt is what any OCR provider returns after reading a receipt image.
type ParsedReceipt struct {
	Vendor     string
	Date       *time.Time
	Amount     *float64
	RawText    string
	Confidence float64 // 0..1 — callers fall back to manual entry below a threshold
}

// Provider is the seam between the receipt-processing flow and whichever OCR
// vendor is behind it. Swapping Google Vision / AWS Textract in later only
// means adding a new Provider implementation and changing the wiring in
// cmd/api/main.go — no business logic changes.
type Provider interface {
	Parse(ctx context.Context, imageBytes []byte) (ParsedReceipt, error)
}
