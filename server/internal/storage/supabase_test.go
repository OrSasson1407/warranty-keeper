package storage_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"warrantykeeper/server/internal/storage"
)

func TestSupabaseStore_UploadReturnsPublicURL(t *testing.T) {
	var gotMethod, gotPath, gotAuth, gotAPIKey, gotContentType, gotBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		gotAPIKey = r.Header.Get("apikey")
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := storage.NewSupabaseStore(server.URL, "receipts", "test-service-role-key")
	url, err := store.Upload(context.Background(), "receipt-123.jpg", []byte("fake-jpeg-bytes"), "image/jpeg")
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}

	wantURL := server.URL + "/storage/v1/object/public/receipts/receipt-123.jpg"
	if url != wantURL {
		t.Errorf("url = %q, want %q", url, wantURL)
	}
	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/storage/v1/object/receipts/receipt-123.jpg" {
		t.Errorf("request path = %q, want the bucket+key upload path", gotPath)
	}
	if gotAuth != "Bearer test-service-role-key" {
		t.Errorf("Authorization header = %q, want the service role bearer token", gotAuth)
	}
	if gotAPIKey != "test-service-role-key" {
		t.Errorf("apikey header = %q, want the service role key", gotAPIKey)
	}
	if gotContentType != "image/jpeg" {
		t.Errorf("Content-Type = %q, want %q", gotContentType, "image/jpeg")
	}
	if gotBody != "fake-jpeg-bytes" {
		t.Errorf("uploaded body = %q, want the raw file bytes", gotBody)
	}
}

func TestSupabaseStore_UploadStripsDirectoryComponentsFromTheKey(t *testing.T) {
	var gotPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	store := storage.NewSupabaseStore(server.URL, "receipts", "key")
	url, err := store.Upload(context.Background(), "../../etc/passwd", []byte("data"), "text/plain")
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}

	if gotPath != "/storage/v1/object/receipts/passwd" {
		t.Errorf("request path = %q, want the traversal stripped down to the base filename", gotPath)
	}
	if url != server.URL+"/storage/v1/object/public/receipts/passwd" {
		t.Errorf("url = %q, want the traversal stripped from the returned URL too", url)
	}
}

func TestSupabaseStore_UploadReturnsErrorOnNonSuccessStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid api key"}`))
	}))
	defer server.Close()

	store := storage.NewSupabaseStore(server.URL, "receipts", "bad-key")
	if _, err := store.Upload(context.Background(), "receipt.jpg", []byte("data"), "image/jpeg"); err == nil {
		t.Error("expected an error when the storage API rejects the upload")
	}
}
