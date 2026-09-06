package ocr

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultGeminiBaseURL = "https://generativelanguage.googleapis.com"

// GeminiProvider extracts receipt fields using a Gemini vision model. Exists
// as a free alternative to AnthropicProvider: Google AI Studio issues API
// keys for Gemini's free tier without requiring a credit card, unlike the
// Anthropic API. See internal/config for GEMINI_API_KEY / GEMINI_OCR_MODEL.
type GeminiProvider struct {
	apiKey     string
	model      string
	httpClient *http.Client
	baseURL    string // overridable in tests; defaults to the real API
}

func NewGeminiProvider(apiKey, model string) *GeminiProvider {
	return &GeminiProvider{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    defaultGeminiBaseURL,
	}
}

type geminiGenerateRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text       string            `json:"text,omitempty"`
	InlineData *geminiInlineData `json:"inlineData,omitempty"`
}

type geminiInlineData struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *GeminiProvider) Parse(ctx context.Context, imageBytes []byte) (ParsedReceipt, error) {
	reqBody := geminiGenerateRequest{
		Contents: []geminiContent{
			{
				Parts: []geminiPart{
					{InlineData: &geminiInlineData{
						MimeType: detectImageMediaType(imageBytes),
						Data:     base64.StdEncoding.EncodeToString(imageBytes),
					}},
					{Text: extractionPrompt},
				},
			},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return ParsedReceipt{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent", p.baseURL, p.model)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return ParsedReceipt{}, fmt.Errorf("failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return ParsedReceipt{}, fmt.Errorf("request to Gemini API failed: %w", err)
	}
	defer resp.Body.Close()

	var parsed geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return ParsedReceipt{}, fmt.Errorf("failed to decode Gemini API response: %w", err)
	}
	if parsed.Error != nil {
		return ParsedReceipt{}, fmt.Errorf("Gemini API error: %s", parsed.Error.Message)
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return ParsedReceipt{}, fmt.Errorf("Gemini API returned no content")
	}

	var extraction receiptExtraction
	text := extractJSONObject(parsed.Candidates[0].Content.Parts[0].Text)
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
