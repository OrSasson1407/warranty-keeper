package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultAnthropicBaseURL = "https://api.anthropic.com"

// AnthropicProvider extracts receipt fields using a vision-capable Claude
// model, replacing the always-empty OCR stub (see stub.go) once an API key
// is configured. See internal/config for ANTHROPIC_API_KEY / ANTHROPIC_OCR_MODEL.
type AnthropicProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string // overridable in tests; defaults to the real API
}

func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    defaultAnthropicBaseURL,
	}
}

const extractionPrompt = `You are reading a photo of a store receipt. Respond with ONLY a single JSON object (no markdown fences, no other text) with exactly these fields:
{
  "vendor": string (the store/vendor name as printed, or "" if unreadable),
  "date": string in YYYY-MM-DD format (the purchase date), or null if unreadable,
  "amount": number (the total amount paid), or null if unreadable,
  "item_description": string (a short description of the main item(s) purchased, in the receipt's own language, or "" if unreadable),
  "confidence": number from 0 to 1 (your confidence that the above fields are accurate)
}`

type anthropicExtraction struct {
	Vendor          string   `json:"vendor"`
	Date            *string  `json:"date"`
	Amount          *float64 `json:"amount"`
	ItemDescription string   `json:"item_description"`
	Confidence      float64  `json:"confidence"`
}

type anthropicMessageRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string             `json:"role"`
	Content []anthropicContent `json:"content"`
}

type anthropicContent struct {
	Type   string          `json:"type"`
	Text   string          `json:"text,omitempty"`
	Source *anthropicImage `json:"source,omitempty"`
}

type anthropicImage struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *AnthropicProvider) Parse(ctx context.Context, imageBytes []byte) (ParsedReceipt, error) {
	reqBody := anthropicMessageRequest{
		Model:     p.model,
		MaxTokens: 500,
		Messages: []anthropicMessage{
			{
				Role: "user",
				Content: []anthropicContent{
					{
						Type: "image",
						Source: &anthropicImage{
							Type:      "base64",
							MediaType: detectImageMediaType(imageBytes),
							Data:      base64.StdEncoding.EncodeToString(imageBytes),
						},
					},
					{Type: "text", Text: extractionPrompt},
				},
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return ParsedReceipt{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return ParsedReceipt{}, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return ParsedReceipt{}, fmt.Errorf("request to Anthropic API failed: %w", err)
	}
	defer resp.Body.Close()

	var parsed anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return ParsedReceipt{}, fmt.Errorf("failed to decode Anthropic API response: %w", err)
	}
	if parsed.Error != nil {
		return ParsedReceipt{}, fmt.Errorf("Anthropic API error: %s", parsed.Error.Message)
	}
	if len(parsed.Content) == 0 {
		return ParsedReceipt{}, fmt.Errorf("Anthropic API returned no content")
	}

	var extraction anthropicExtraction
	text := extractJSONObject(parsed.Content[0].Text)
	if err := json.Unmarshal([]byte(text), &extraction); err != nil {
		return ParsedReceipt{}, fmt.Errorf("failed to parse model output as JSON: %w", err)
	}

	var date *time.Time
	if extraction.Date != nil {
		if d, err := time.Parse("2006-01-02", *extraction.Date); err == nil {
			date = &d
		}
	}

	return ParsedReceipt{
		Vendor:     extraction.Vendor,
		Date:       date,
		Amount:     extraction.Amount,
		RawText:    extraction.ItemDescription,
		Confidence: extraction.Confidence,
	}, nil
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
