package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
)

// SupabaseStore uploads files to a Supabase Storage bucket over its REST
// API, using the service_role key so uploads work regardless of the
// bucket's row-level-security policies. The bucket itself must be created
// as public ahead of time, so the returned URL is directly usable by the
// mobile app without any auth header.
type SupabaseStore struct {
	baseURL    string
	bucket     string
	serviceKey string
	httpClient *http.Client
}

func NewSupabaseStore(projectURL, bucket, serviceRoleKey string) *SupabaseStore {
	return &SupabaseStore{
		baseURL:    projectURL,
		bucket:     bucket,
		serviceKey: serviceRoleKey,
		httpClient: &http.Client{},
	}
}

func (s *SupabaseStore) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	// Only the base filename is ever used, same safety rule as LocalStore:
	// a caller-supplied key must not be interpreted as a path.
	name := filepath.Base(key)
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, s.bucket, url.PathEscape(name))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("build upload request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("supabase storage upload failed (status %d): %s", resp.StatusCode, body)
	}

	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.baseURL, s.bucket, url.PathEscape(name)), nil
}
