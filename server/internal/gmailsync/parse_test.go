package gmailsync_test

import (
	"strings"
	"testing"
	"time"

	"warrantykeeper/server/internal/gmailsync"
)

func TestMatchVendor_AllowlistedDomains(t *testing.T) {
	cases := []struct {
		from       string
		wantVendor string
		wantOK     bool
	}{
		{"Amazon.com <auto-confirm@amazon.com>", "Amazon", true},
		{"Amazon <shipment-tracking@amazon.co.il>", "Amazon", true},
		{"KSP <noreply@ksp.co.il>", "KSP", true},
		{"IKEA <no-reply@ikea.com>", "איקאה", true},
		{"IKEA Israel <orders@ikea.co.il>", "איקאה", true},
		{"Someone Else <hello@totally-unrelated-store.com>", "", false},
		{"not even an email address", "", false},
	}
	for _, tc := range cases {
		vendor, ok := gmailsync.MatchVendor(tc.from)
		if ok != tc.wantOK || vendor != tc.wantVendor {
			t.Errorf("MatchVendor(%q) = (%q, %v), want (%q, %v)", tc.from, vendor, ok, tc.wantVendor, tc.wantOK)
		}
	}
}

func TestGmailSearchQuery_IncludesAllAllowlistedDomainsAndLookback(t *testing.T) {
	query := gmailsync.GmailSearchQuery(14)
	if !strings.Contains(query, "newer_than:14d") {
		t.Errorf("query = %q, want it to include the lookback window", query)
	}
	for _, v := range gmailsync.AllowedVendors {
		for _, d := range v.SenderDomains {
			if !strings.Contains(query, "from:"+d) {
				t.Errorf("query = %q, want it to include from:%s", query, d)
			}
		}
	}
}

func TestParseOrderEmail_ExtractsShekelAmount(t *testing.T) {
	date := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	result := gmailsync.ParseOrderEmail(
		"Amazon.com <auto-confirm@amazon.com>",
		"ההזמנה שלך אושרה",
		date,
		"תודה על ההזמנה. סכום לתשלום: ₪349.90",
	)
	if result.Vendor != "Amazon" {
		t.Errorf("Vendor = %q, want %q", result.Vendor, "Amazon")
	}
	if result.Amount == nil || *result.Amount != 349.90 {
		t.Errorf("Amount = %v, want %v", result.Amount, 349.90)
	}
	if result.Date == nil || !result.Date.Equal(date) {
		t.Errorf("Date = %v, want %v", result.Date, date)
	}
	if result.Confidence <= 0.3 {
		t.Errorf("Confidence = %v, want it boosted above the no-amount baseline once an amount is found", result.Confidence)
	}
}

func TestParseOrderEmail_ExtractsDollarAmount(t *testing.T) {
	result := gmailsync.ParseOrderEmail(
		"orders@ikea.com",
		"Your order confirmation",
		time.Now(),
		"Order total: $129.99",
	)
	if result.Amount == nil || *result.Amount != 129.99 {
		t.Errorf("Amount = %v, want %v", result.Amount, 129.99)
	}
}

func TestParseOrderEmail_NoAmountFoundStaysLowConfidence(t *testing.T) {
	result := gmailsync.ParseOrderEmail(
		"noreply@ksp.co.il",
		"ההזמנה שלך יצאה למשלוח",
		time.Now(),
		"ההזמנה שלך בדרך אליך.",
	)
	if result.Amount != nil {
		t.Errorf("Amount = %v, want nil (no currency amount in the body)", result.Amount)
	}
	if result.Confidence != 0.3 {
		t.Errorf("Confidence = %v, want the baseline 0.3 when nothing beyond a vendor match was found", result.Confidence)
	}
}

func TestParseOrderEmail_UnmatchedSenderYieldsEmptyVendor(t *testing.T) {
	result := gmailsync.ParseOrderEmail("hello@totally-unrelated-store.com", "Order", time.Now(), "")
	if result.Vendor != "" {
		t.Errorf("Vendor = %q, want empty for a non-allowlisted sender", result.Vendor)
	}
}
