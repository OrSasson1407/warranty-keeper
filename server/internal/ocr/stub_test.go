package ocr_test

import (
	"context"
	"testing"
	"time"

	"warrantykeeper/server/internal/ocr"
)

func TestStubProvider_ReportsZeroConfidenceSoCallersFallBackToManualEntry(t *testing.T) {
	p := ocr.NewStubProvider()
	result, err := p.Parse(context.Background(), []byte("irrelevant"))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if result.Confidence != 0 {
		t.Errorf("Confidence = %v, want 0 (never confident enough to skip manual review)", result.Confidence)
	}
}

func TestStubProvider_LeavesVendorAndAmountEmpty(t *testing.T) {
	p := ocr.NewStubProvider()
	result, err := p.Parse(context.Background(), []byte("irrelevant"))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if result.Vendor != "" {
		t.Errorf("Vendor = %q, want empty (nothing was really recognized)", result.Vendor)
	}
	if result.Amount != nil {
		t.Errorf("Amount = %v, want nil", result.Amount)
	}
}

func TestStubProvider_RawTextExplainsItIsAStub(t *testing.T) {
	p := ocr.NewStubProvider()
	result, err := p.Parse(context.Background(), []byte("irrelevant"))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if result.RawText == "" {
		t.Error("expected a non-empty explanatory raw text")
	}
}

func TestStubProvider_DateDefaultsToApproximatelyNow(t *testing.T) {
	p := ocr.NewStubProvider()
	before := time.Now()
	result, err := p.Parse(context.Background(), []byte("irrelevant"))
	after := time.Now()
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if result.Date == nil {
		t.Fatal("expected a non-nil Date so callers have something to resolve a warranty against")
	}
	if result.Date.Before(before) || result.Date.After(after) {
		t.Errorf("Date = %v, want it between %v and %v", result.Date, before, after)
	}
}

func TestStubProvider_IgnoresInputBytes(t *testing.T) {
	p := ocr.NewStubProvider()
	resultEmpty, err := p.Parse(context.Background(), []byte{})
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	resultLarge, err := p.Parse(context.Background(), make([]byte, 1024))
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if resultEmpty.RawText != resultLarge.RawText || resultEmpty.Confidence != resultLarge.Confidence {
		t.Error("the stub is documented as ignoring its input; results should be equivalent regardless of image bytes")
	}
}
