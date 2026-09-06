package ocr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestProvider(t *testing.T, handler http.HandlerFunc) *AnthropicProvider {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	p := NewAnthropicProvider("test-key", "claude-haiku-4-5-20251001")
	p.baseURL = server.URL
	return p
}

func respondWithText(t *testing.T, text string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			t.Errorf("expected x-api-key header to be set")
		}
		if r.Header.Get("anthropic-version") == "" {
			t.Errorf("expected anthropic-version header to be set")
		}

		var reqBody anthropicMessageRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if reqBody.Model != "claude-haiku-4-5-20251001" {
			t.Errorf("expected model to be passed through, got %q", reqBody.Model)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(anthropicResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{{Type: "text", Text: text}},
		})
	}
}

func TestParse_ExtractsFieldsFromWellFormedJSON(t *testing.T) {
	p := newTestProvider(t, respondWithText(t, `{"vendor":"KSP","date":"2026-03-15","amount":499.9,"item_description":"אוזניות בלוטות'","confidence":0.92}`))

	result, err := p.Parse(context.Background(), []byte{0xFF, 0xD8, 0xFF}) // JPEG magic bytes
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if result.Vendor != "KSP" {
		t.Errorf("Vendor = %q, want %q", result.Vendor, "KSP")
	}
	if result.Date == nil || result.Date.Format("2006-01-02") != "2026-03-15" {
		t.Errorf("Date = %v, want 2026-03-15", result.Date)
	}
	if result.Amount == nil || *result.Amount != 499.9 {
		t.Errorf("Amount = %v, want 499.9", result.Amount)
	}
	if result.RawText != "אוזניות בלוטות'" {
		t.Errorf("RawText = %q, want the item description", result.RawText)
	}
	if result.Confidence != 0.92 {
		t.Errorf("Confidence = %v, want 0.92", result.Confidence)
	}
}

func TestParse_StripsMarkdownFencesAroundJSON(t *testing.T) {
	p := newTestProvider(t, respondWithText(t, "```json\n{\"vendor\":\"KSP\",\"date\":null,\"amount\":null,\"item_description\":\"\",\"confidence\":0.4}\n```"))

	result, err := p.Parse(context.Background(), []byte{0xFF, 0xD8, 0xFF})
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if result.Vendor != "KSP" {
		t.Errorf("Vendor = %q, want %q", result.Vendor, "KSP")
	}
	if result.Date != nil {
		t.Errorf("Date = %v, want nil", result.Date)
	}
}

func TestParse_ReturnsErrorOnUnparseableModelOutput(t *testing.T) {
	p := newTestProvider(t, respondWithText(t, "sorry, I can't read this receipt"))

	_, err := p.Parse(context.Background(), []byte{0xFF, 0xD8, 0xFF})
	if err == nil {
		t.Fatal("expected an error when the model doesn't return JSON")
	}
}

func TestParse_ReturnsErrorOnAPIErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(anthropicResponse{
			Error: &struct {
				Message string `json:"message"`
			}{Message: "invalid x-api-key"},
		})
	}))
	defer server.Close()

	p := NewAnthropicProvider("bad-key", "claude-haiku-4-5-20251001")
	p.baseURL = server.URL

	_, err := p.Parse(context.Background(), []byte{0xFF, 0xD8, 0xFF})
	if err == nil {
		t.Fatal("expected an error when the API responds with an error field")
	}
}

func TestParse_ReturnsErrorWhenServerUnreachable(t *testing.T) {
	p := NewAnthropicProvider("test-key", "claude-haiku-4-5-20251001")
	p.baseURL = "http://127.0.0.1:1" // nothing listens here

	_, err := p.Parse(context.Background(), []byte{0xFF, 0xD8, 0xFF})
	if err == nil {
		t.Fatal("expected an error when the request itself fails")
	}
}

func TestDetectImageMediaType(t *testing.T) {
	cases := []struct {
		name  string
		bytes []byte
		want  string
	}{
		{"JPEG", []byte{0xFF, 0xD8, 0xFF, 0xE0}, "image/jpeg"},
		{"PNG", []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, "image/png"},
		{"GIF", []byte("GIF89a"), "image/gif"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := detectImageMediaType(tc.bytes); got != tc.want {
				t.Errorf("detectImageMediaType(%s) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}
