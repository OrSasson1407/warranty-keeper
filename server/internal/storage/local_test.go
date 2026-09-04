package storage_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"warrantykeeper/server/internal/storage"
)

func TestNewLocalStore_CreatesTheUploadDirIfMissing(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "does", "not", "exist", "yet")

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Fatalf("test setup: %q should not exist yet", dir)
	}

	if _, err := storage.NewLocalStore(dir, "https://example.test/uploads"); err != nil {
		t.Fatalf("NewLocalStore returned error: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("expected the upload dir to be created: %v", err)
	}
	if !info.IsDir() {
		t.Errorf("%q was created but is not a directory", dir)
	}
}

func TestLocalStore_UploadWritesFileAndReturnsURL(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocalStore(dir, "https://example.test/uploads")
	if err != nil {
		t.Fatalf("NewLocalStore returned error: %v", err)
	}

	content := []byte("fake-jpeg-bytes")
	url, err := store.Upload(context.Background(), "receipt-123.jpg", content, "image/jpeg")
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}

	wantURL := "https://example.test/uploads/receipt-123.jpg"
	if url != wantURL {
		t.Errorf("url = %q, want %q", url, wantURL)
	}

	got, err := os.ReadFile(filepath.Join(dir, "receipt-123.jpg"))
	if err != nil {
		t.Fatalf("expected the file to exist on disk: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("file content = %q, want %q", got, content)
	}
}

func TestLocalStore_UploadStripsDirectoryComponentsFromTheKey(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocalStore(dir, "https://example.test/uploads")
	if err != nil {
		t.Fatalf("NewLocalStore returned error: %v", err)
	}

	// A key containing path traversal must not escape the upload directory —
	// only the base filename should ever be used.
	url, err := store.Upload(context.Background(), "../../etc/passwd", []byte("data"), "text/plain")
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if url != "https://example.test/uploads/passwd" {
		t.Errorf("url = %q, want the traversal stripped down to the base filename", url)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("failed to read upload dir: %v", err)
	}
	if len(entries) != 1 || entries[0].Name() != "passwd" {
		t.Errorf("upload dir contents = %v, want exactly one file named \"passwd\"", entries)
	}

	if _, err := os.Stat(filepath.Join(dir, "..", "..", "etc", "passwd")); !os.IsNotExist(err) {
		t.Error("the file must not have been written outside the upload directory")
	}
}

func TestLocalStore_MultipleUploadsProduceSeparateFiles(t *testing.T) {
	dir := t.TempDir()
	store, err := storage.NewLocalStore(dir, "https://example.test/uploads")
	if err != nil {
		t.Fatalf("NewLocalStore returned error: %v", err)
	}

	if _, err := store.Upload(context.Background(), "a.jpg", []byte("A"), "image/jpeg"); err != nil {
		t.Fatalf("Upload a.jpg returned error: %v", err)
	}
	if _, err := store.Upload(context.Background(), "b.jpg", []byte("B"), "image/jpeg"); err != nil {
		t.Fatalf("Upload b.jpg returned error: %v", err)
	}

	a, err := os.ReadFile(filepath.Join(dir, "a.jpg"))
	if err != nil || string(a) != "A" {
		t.Errorf("a.jpg content = %q, err = %v, want \"A\"", a, err)
	}
	b, err := os.ReadFile(filepath.Join(dir, "b.jpg"))
	if err != nil || string(b) != "B" {
		t.Errorf("b.jpg content = %q, err = %v, want \"B\"", b, err)
	}
}
