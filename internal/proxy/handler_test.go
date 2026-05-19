package proxy

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"graceful-echo-server/internal/echo"
)

func TestReverseProxyForwardsRequest(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/echo" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("msg"); got != "proxy" {
			t.Fatalf("unexpected query msg: %s", got)
		}
		if got := r.Header.Get("X-Forwarded-Proto"); got != "https" {
			t.Fatalf("unexpected X-Forwarded-Proto: %s", got)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("proxied"))
	}))
	defer upstream.Close()

	targetURL, err := url.Parse(upstream.URL)
	if err != nil {
		t.Fatalf("parse upstream URL failed: %v", err)
	}
	proxy := httptest.NewTLSServer(NewHandler(targetURL, Options{}))
	defer proxy.Close()

	client := proxy.Client()
	resp, err := client.Get(proxy.URL + "/echo?msg=proxy")
	if err != nil {
		t.Fatalf("proxy request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response failed: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	if got, want := string(body), "proxied"; got != want {
		t.Fatalf("unexpected response body: got %q want %q", got, want)
	}
}

func TestReverseProxyUpstreamUnavailable(t *testing.T) {
	targetURL, err := url.Parse("http://127.0.0.1:1") // nothing listens
	if err != nil {
		t.Fatalf("parse URL: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/echo", nil)
	NewHandler(targetURL, Options{UpstreamTimeout: 500 * time.Millisecond}).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
}

func TestEchoEndToEndViaTLSProxy(t *testing.T) {
	backend := httptest.NewServer(echo.NewMux())
	defer backend.Close()

	targetURL, err := url.Parse(backend.URL)
	if err != nil {
		t.Fatalf("parse backend URL: %v", err)
	}

	tlsProxy := httptest.NewTLSServer(NewHandler(targetURL, Options{}))
	defer tlsProxy.Close()

	resp, err := tlsProxy.Client().Get(tlsProxy.URL + "/echo?msg=via-proxy&delay=0")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if got, want := string(body), "echo: via-proxy\n"; got != want {
		t.Fatalf("body = %q want %q", got, want)
	}
}
