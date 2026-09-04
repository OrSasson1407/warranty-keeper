package warranty

import "strings"

// categoryKeywords is a simple keyword -> category lookup used to guess a
// product's category from OCR text (architecture doc step 5). This is
// deliberately basic substring matching for MVP, not NLP/ML classification
// (full categorization is explicitly out of scope for V1).
var categoryKeywords = map[string]string{
	"מזגן":     "מזגן",
	"מקרר":     "מקרר",
	"מקפיא":    "מקפיא",
	"כביסה":    "מכונת כביסה",
	"מייבש":    "מייבש כביסה",
	"מדיח":     "מדיח כלים",
	"תנור":     "תנור בנוי",
	"כיריים":   "כיריים",
	"דוד שמש":  "דוד שמש",
	"מיקרוגל":  "מיקרוגל",
	"שואב אבק": "שואב אבק",
	"קפה":      "מכונת קפה",
	"טוסטר":    "טוסטר אובן",
	"קומקום":   "קומקום חשמלי",
	"מאוורר":   "מאוורר",
	"מטהר אוויר": "מטהר אוויר",
	"טלוויזיה": "טלוויזיה",
	"מחשב נייד": "מחשב נייד",
	"מחשב נייח": "מחשב נייח",
	"טאבלט":    "טאבלט",
	"סמארטפון": "סמארטפון",
	"אוזניות":  "אוזניות",
	"רמקול":    "רמקול בלוטות'",
	"מדפסת":    "מדפסת",
	"מצלמה":    "מצלמה",
	"שעון חכם": "שעון חכם",
	"קונסולה":  "קונסולת משחקים",
	"נתב":      "נתב אינטרנט",
	"ספה":      "ספה",
	"מזרן":     "מזרן",
	"כיסא":     "כיסא משרדי",
	"ארון":     "ארון בגדים",
}

// GuessCategory does a case-insensitive keyword search over free text
// (typically OCR raw text + vendor name). Returns "" when nothing matches,
// leaving the field for manual entry on the confirmation screen.
func GuessCategory(text string) string {
	lower := strings.ToLower(text)
	for keyword, category := range categoryKeywords {
		if strings.Contains(lower, strings.ToLower(keyword)) {
			return category
		}
	}
	return ""
}
