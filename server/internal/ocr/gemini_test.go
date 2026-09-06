package ocr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestGeminiProvider(t *testing.T, handler http.HandlerFunc) *GeminiProvider {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	p := NewGeminiProvider("test-key", "gemini-2.0-flash")
	p.baseURL = server.URL
	return p
}

func geminiRespondWithText(t *testing.T, text string) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-goog-api-key") != "test-key" {
			t.Errorf("expected x-goog-api-key header to be set")
		}
		if !strings.Contains(r.URL.Path, "gemini-2.0-flash") {
			t.Errorf("expected model to appear in the URL path, got %q", r.URL.Path)
		}

		var reqBody geminiGenerateRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if len(reqBody.Contents) == 0 || len(reqBody.Contents[0].Parts) != 2 {
			t.Fatalf("expected one content block with an image part and a text part, got %+v", reqBody.Contents)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"candidates": []map[string]any{
				{"content": map[string]any{"parts": []map[string]any{{"text": text}}}},
			},
		})
	}
}

func TestGeminiParse_ExtractsFieldsFromWellFormedJSON(t *testing.T) {
	p := newTestGeminiProvider(t, geminiRespondWithText(t, `{"vendor":"KSP","date":"2026-03-15","amount":499.9,"item_description":"אוזניות בלוטות'","confidence":0.85}`))

	result, err := p.Parse(context.Background(), []byte{0xFF, 0xD8, 0xFF})
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
	if result.Confidence != 0.85 {
		t.Errorf("Confidence = %v, want 0.85", result.Confidence)
	}
}

func TestGeminiParse_StripsMarkdownFencesAroundJSON(t *testing.T) {
	p := newTestGeminiProvider(t, geminiRespondWithText(t, "```json\n{\"vendor\":\"KSP\",\"date\":null,\"amount\":null,\"item_description\":\"\",\"confidence\":0.3}\n```"))

	result, err := p.Parse(context.Background(), []byte{0xFF, 0xD8, 0xFF})
	if err != nil {
		t.Fatalf("Parse returned an error: %v", err)
	}
	if result.Vendor != "KSP" {
		t.Errorf("Vendor = %q, want %q", result.Vendor, "KSP")
	}
}

func TestGeminiParse_ReturnsErrorOnUnparseableModelOutput(t *testing.T) {
	p := newTestGeminiProvider(t, geminiRespondWithText(t, "sorry, I can't read this receipt"))

	_, err := p.Parse(context.Background(), []byte{0xFF, 0xD8, 0xFF})
	if err == nil {
		t.Fatal("expected an error when the model doesn't return JSON")
	}
}

func TestGeminiParse_ReturnsErrorOnAPIErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"message": "API key not valid"},
		})
	}))
	defer server.Close()

	p := NewGeminiProvider("bad-key", "gemini-2.0-flash")
	p.baseURL = server.URL

	_, err := p.Parse(context.Background(), []byte{0xFF, 0xD8, 0xFF})
	if err == nil {
		t.Fatal("expected an error when the API responds with an error field")
	}
}

func TestGeminiParse_ReturnsErrorOnEmptyCandidates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"candidates": []map[string]any{}})
	}))
	defer server.Close()

	p := NewGeminiProvider("test-key", "gemini-2.0-flash")
	p.baseURL = server.URL

	_, err := p.Parse(context.Background(), []byte{0xFF, 0xD8, 0xFF})
	if err == nil {
		t.Fatal("expected an error when there are no candidates")
	}
}

func TestGeminiParse_ReturnsErrorWhenServerUnreachable(t *testing.T) {
	p := NewGeminiProvider("test-key", "gemini-2.0-flash")
	p.baseURL = "http://127.0.0.1:1"

	_, err := p.Parse(context.Background(), []byte{0xFF, 0xD8, 0xFF})
	if err == nil {
		t.Fatal("expected an error when the request itself fails")
	}
}
