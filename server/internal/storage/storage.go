package storage

import "context"

// Store is the seam between receipt/product image uploads and whichever
// object storage backend is behind them. A real S3-compatible
// implementation can be swapped in later without touching handler code.
type Store interface {
	Upload(ctx context.Context, key string, data []byte, contentType string) (url string, err error)
}
