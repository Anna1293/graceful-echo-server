package config

import (
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Setenv("BACKEND_ADDR", "")
	t.Setenv("PROXY_ADDR", "")
	t.Setenv("TLS_CERT_FILE", "")
	t.Setenv("TLS_KEY_FILE", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BackendAddr != ":8080" {
		t.Fatalf("BackendAddr = %q, want :8080", cfg.BackendAddr)
	}
	if cfg.ProxyAddr != ":8443" {
		t.Fatalf("ProxyAddr = %q, want :8443", cfg.ProxyAddr)
	}
	if cfg.CertFile != "certs/server.crt" {
		t.Fatalf("CertFile = %q", cfg.CertFile)
	}
}

func TestUpstreamURL(t *testing.T) {
	cfg := Config{BackendAddr: ":9090"}
	u, err := cfg.UpstreamURL()
	if err != nil {
		t.Fatalf("UpstreamURL: %v", err)
	}
	if got, want := u.String(), "http://127.0.0.1:9090"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
