package ocr

import (
	"net/http"
	"strings"
)

// extractionPrompt is shared by every vision-model-backed Provider (Anthropic,
// Gemini, ...) -- the extraction contract doesn't depend on which model reads
// the image.
const extractionPrompt = `You are reading a photo of a store receipt. Respond with ONLY a single JSON object (no markdown fences, no other text) with exactly these fields:
{
  "vendor": string (the store/vendor name as printed, or "" if unreadable),
  "date": string in YYYY-MM-DD format (the purchase date), or null if unreadable,
  "amount": number (the total amount paid), or null if unreadable,
  "item_description": string (a short description of the main item(s) purchased, in the receipt's own language, or "" if unreadable),
  "confidence": number from 0 to 1 (your confidence that the above fields are accurate)
}`

// receiptExtraction is the JSON shape every provider prompts the model to
// return, per extractionPrompt above.
type receiptExtraction struct {
	Vendor          string   `json:"vendor"`
	Date            *string  `json:"date"`
	Amount          *float64 `json:"amount"`
	ItemDescription string   `json:"item_description"`
	Confidence      float64  `json:"confidence"`
}

// extractJSONObject strips any leading/trailing text or markdown code fences
// around a JSON object, in case the model doesn't follow the "only JSON"
// instruction exactly.
func extractJSONObject(text string) string {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start == -1 || end == -1 || end < start {
		return text
	}
	return text[start : end+1]
}

func detectImageMediaType(imageBytes []byte) string {
	contentType := http.DetectContentType(imageBytes)
	switch {
	case strings.Contains(contentType, "png"):
		return "image/png"
	case strings.Contains(contentType, "gif"):
		return "image/gif"
	case strings.Contains(contentType, "webp"):
		return "image/webp"
	default:
		return "image/jpeg"
	}
}
