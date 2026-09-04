package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// LocalStore writes uploads to disk and serves them back via a URL prefix
// the API router exposes as a static route (see internal/api/router.go).
// Stands in for the real S3-compatible bucket until credentials are wired up.
type LocalStore struct {
	dir       string
	publicURL string
}

func NewLocalStore(dir, publicURL string) (*LocalStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &LocalStore{dir: dir, publicURL: publicURL}, nil
}

func (s *LocalStore) Upload(_ context.Context, key string, data []byte, _ string) (string, error) {
	path := filepath.Join(s.dir, filepath.Base(key))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return fmt.Sprintf("%s/%s", s.publicURL, filepath.Base(key)), nil
}
