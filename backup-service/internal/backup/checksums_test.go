package backup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChecksumsRoundTrip(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "data"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data", "a.bin"), []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data", "b.bin"), []byte("world"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := WriteChecksums(dir); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := VerifyChecksums(dir); err != nil {
		t.Fatalf("verify clean: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "data", "a.bin"), []byte("tampered"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := VerifyChecksums(dir); err == nil {
		t.Fatal("expected verify to fail after tampering")
	}
}

func TestVerifyNoChecksums(t *testing.T) {
	dir := t.TempDir()
	if err := VerifyChecksums(dir); err != ErrNoChecksums {
		t.Fatalf("expected ErrNoChecksums, got %v", err)
	}
}
