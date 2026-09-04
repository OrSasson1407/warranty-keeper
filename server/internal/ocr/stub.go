package ocr

import (
	"context"
	"time"
)

// StubProvider stands in for a real OCR vendor (Google Vision / AWS Textract)
// until credentials are wired up. It always reports low confidence so the
// receipt-processing flow correctly falls back to manual entry (per the
// architecture doc's "confidence too low" case), while still exercising the
// full upload -> draft -> confirm pipeline end-to-end.
type StubProvider struct{}

func NewStubProvider() *StubProvider {
	return &StubProvider{}
}

func (p *StubProvider) Parse(_ context.Context, _ []byte) (ParsedReceipt, error) {
	now := time.Now()
	return ParsedReceipt{
		Vendor:     "",
		Date:       &now,
		Amount:     nil,
		RawText:    "[OCR stub: no provider configured — confirm all fields manually]",
		Confidence: 0,
	}, nil
}
