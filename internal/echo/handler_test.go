package echo

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	healthHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got := rec.Body.String(); got != "ok\n" {
		t.Fatalf("body = %q", got)
	}
}

func TestEchoDefaults(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/echo", nil)
	rec := httptest.NewRecorder()

	echoHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if got := strings.TrimSpace(string(body)); got != "echo: echo" {
		t.Fatalf("body = %q", got)
	}
}

func TestEchoCustomMessage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/echo?msg=hello", nil)
	rec := httptest.NewRecorder()

	echoHandler(rec, req)

	body, _ := io.ReadAll(rec.Body)
	if got := strings.TrimSpace(string(body)); got != "echo: hello" {
		t.Fatalf("body = %q", got)
	}
}

func TestEchoInvalidDelay(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/echo?delay=-1", nil)
	rec := httptest.NewRecorder()

	echoHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
}

func TestEchoMethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/echo", nil)
	rec := httptest.NewRecorder()

	echoHandler(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
}
