package certs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureSelfSignedCreatesFiles(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureSelfSigned(certPath, keyPath); err != nil {
		t.Fatalf("EnsureSelfSigned: %v", err)
	}
	for _, path := range []string{certPath, keyPath} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("stat %s: %v", path, err)
		}
	}
}

func TestEnsureSelfSignedIdempotent(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "server.crt")
	keyPath := filepath.Join(dir, "server.key")

	if err := EnsureSelfSigned(certPath, keyPath); err != nil {
		t.Fatalf("first EnsureSelfSigned: %v", err)
	}
	info, err := os.Stat(certPath)
	if err != nil {
		t.Fatalf("stat cert: %v", err)
	}
	modTime := info.ModTime()

	if err := EnsureSelfSigned(certPath, keyPath); err != nil {
		t.Fatalf("second EnsureSelfSigned: %v", err)
	}
	info2, err := os.Stat(certPath)
	if err != nil {
		t.Fatalf("stat cert again: %v", err)
	}
	if !info2.ModTime().Equal(modTime) {
		t.Fatal("expected existing cert not to be regenerated")
	}
}
