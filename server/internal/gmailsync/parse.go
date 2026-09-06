package gmailsync

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Vendor is one retailer this integration is allowed to scan for. Per the
// v2 scope doc's trust/privacy mitigation, scanning is limited to a small,
// named allowlist rather than reading every email in the inbox -- adding a
// retailer means adding an entry here, not loosening a general filter.
type Vendor struct {
	Name          string
	SenderDomains []string
}

var AllowedVendors = []Vendor{
	{Name: "Amazon", SenderDomains: []string{"amazon.com", "amazon.co.il"}},
	{Name: "KSP", SenderDomains: []string{"ksp.co.il"}},
	{Name: "איקאה", SenderDomains: []string{"ikea.com", "ikea.co.il"}},
}

var emailAddrRe = regexp.MustCompile(`[\w.+-]+@([\w-]+(?:\.[\w-]+)+)`)

func extractDomain(fromHeader string) string {
	m := emailAddrRe.FindStringSubmatch(strings.ToLower(fromHeader))
	if len(m) < 2 {
		return ""
	}
	return m[1]
}

// MatchVendor reports whether a message's From header address falls under
// an allowlisted retailer's domain, and if so which one.
func MatchVendor(fromHeader string) (name string, ok bool) {
	domain := extractDomain(fromHeader)
	if domain == "" {
		return "", false
	}
	for _, v := range AllowedVendors {
		for _, d := range v.SenderDomains {
			if domain == d || strings.HasSuffix(domain, "."+d) {
				return v.Name, true
			}
		}
	}
	return "", false
}

// GmailSearchQuery builds the Gmail search restricting a scan to allowlisted
// senders within a lookback window, so a scan is cheap, predictable, and
// never touches mail outside the declared allowlist.
func GmailSearchQuery(lookbackDays int) string {
	domains := make([]string, 0)
	for _, v := range AllowedVendors {
		for _, d := range v.SenderDomains {
			domains = append(domains, "from:"+d)
		}
	}
	return fmt.Sprintf("newer_than:%dd (%s)", lookbackDays, strings.Join(domains, " OR "))
}

// amountRe matches a currency amount tagged with a shekel/ILS or dollar
// symbol, in either order (symbol-then-number or number-then-symbol) since
// retailer templates vary.
var amountRe = regexp.MustCompile(`(?:₪|ILS|NIS)\s?([\d,]+(?:\.\d{1,2})?)|([\d,]+(?:\.\d{1,2})?)\s?(?:₪|ILS|NIS)|\$\s?([\d,]+(?:\.\d{1,2})?)`)

func extractAmount(text string) *float64 {
	m := amountRe.FindStringSubmatch(text)
	if m == nil {
		return nil
	}
	for _, g := range m[1:] {
		if g == "" {
			continue
		}
		clean := strings.ReplaceAll(g, ",", "")
		if v, err := strconv.ParseFloat(clean, 64); err == nil {
			return &v
		}
	}
	return nil
}

func firstLine(s string, max int) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexAny(s, "\r\n"); i >= 0 {
		s = s[:i]
	}
	if len(s) > max {
		return s[:max]
	}
	return s
}

// ParsedOrderEmail is the Gmail-integration equivalent of ocr.ParsedReceipt:
// a best-effort extraction from an order-confirmation email.
type ParsedOrderEmail struct {
	Vendor     string
	Date       *time.Time
	Amount     *float64
	Snippet    string
	Confidence float64
}

// ParseOrderEmail extracts vendor/date/amount from a matched email.
// Deliberately crude: there is no per-vendor template parser, just a
// currency-amount regex over the subject and body -- so confidence is
// always capped low, which routes every result through the same
// low-confidence manual-review banner as uncertain photo OCR rather than
// silently trusting a guess. Improving parsing accuracy needs real sample
// data from actual retailer emails, which isn't available in development.
func ParseOrderEmail(fromHeader, subject string, date time.Time, bodyText string) ParsedOrderEmail {
	vendor, _ := MatchVendor(fromHeader)
	amount := extractAmount(subject + "\n" + bodyText)

	confidence := 0.3
	if amount != nil {
		confidence = 0.45
	}

	snippet := subject
	if bodyText != "" {
		snippet = subject + " — " + firstLine(bodyText, 200)
	}

	d := date
	return ParsedOrderEmail{
		Vendor:     vendor,
		Date:       &d,
		Amount:     amount,
		Snippet:    snippet,
		Confidence: confidence,
	}
}
