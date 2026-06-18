package proxy

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRemoveHopByHopHeaders(t *testing.T) {
	h := http.Header{}
	h.Set("Connection", "close")
	h.Set("Content-Type", "text/plain")

	removeHopByHopHeaders(h)

	if h.Get("Connection") != "" {
		t.Fatal("Connection header should be removed")
	}
	if h.Get("Content-Type") != "text/plain" {
		t.Fatal("Content-Type should remain")
	}
}

func TestSetForwardedHeaders(t *testing.T) {
	client := httptest.NewRequest(http.MethodGet, "https://example.com/echo", nil)
	client.Host = "example.com"
	client.RemoteAddr = "203.0.113.5:12345"
	client.Header.Set("X-Forwarded-For", "198.51.100.1")

	dst := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080/echo", nil)
	setForwardedHeaders(dst, client, "https")

	if got := dst.Header.Get("X-Forwarded-For"); got != "198.51.100.1, 203.0.113.5" {
		t.Fatalf("X-Forwarded-For = %q", got)
	}
	if got := dst.Header.Get("X-Forwarded-Host"); got != "example.com" {
		t.Fatalf("X-Forwarded-Host = %q", got)
	}
	if got := dst.Header.Get("X-Forwarded-Proto"); got != "https" {
		t.Fatalf("X-Forwarded-Proto = %q", got)
	}
}

func TestClientIPFromRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "10.0.0.1:4321"
	if got := clientIPFromRequest(r); got != "10.0.0.1" {
		t.Fatalf("got %q", got)
	}
}
