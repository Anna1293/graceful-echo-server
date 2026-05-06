package main

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"
)

func startTestServer(t *testing.T) (baseURL string, srv *http.Server, done <-chan error) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen failed: %v", err)
	}

	srv = newServer(listener.Addr().String())
	serverDone := make(chan error, 1)
	go func() {
		err := srv.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverDone <- err
			return
		}
		serverDone <- nil
	}()

	return "http://" + listener.Addr().String(), srv, serverDone
}

func TestGracefulShutdownWaitsForInFlightRequest(t *testing.T) {
	baseURL, srv, serverDone := startTestServer(t)

	var (
		wg       sync.WaitGroup
		reqErr   error
		respBody string
		respCode int
	)

	wg.Add(1)
	go func() {
		defer wg.Done()

		resp, err := http.Get(baseURL + "/echo?msg=grace&delay=1")
		if err != nil {
			reqErr = err
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			reqErr = err
			return
		}
		respCode = resp.StatusCode
		respBody = string(body)
	}()

	// Даем обработчику зайти в долгую работу до вызова Shutdown.
	time.Sleep(150 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	wg.Wait()
	if reqErr != nil {
		t.Fatalf("request failed: %v", reqErr)
	}
	if respCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", respCode)
	}
	if got, want := strings.TrimSpace(respBody), "echo: grace"; got != want {
		t.Fatalf("unexpected response body: got %q want %q", got, want)
	}

	if err := <-serverDone; err != nil {
		t.Fatalf("server exited with error: %v", err)
	}
}

func TestGracefulShutdownRejectsNewRequests(t *testing.T) {
	baseURL, srv, serverDone := startTestServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	_, err := http.Get(baseURL + "/echo?msg=after")
	if err == nil {
		t.Fatal("expected request to fail after shutdown")
	}

	if err := <-serverDone; err != nil {
		t.Fatalf("server exited with error: %v", err)
	}
}

func TestConcurrentShutdownCallsAreSafe(t *testing.T) {
	_, srv, serverDone := startTestServer(t)

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	shutdown := func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		errCh <- srv.Shutdown(ctx)
	}

	wg.Add(2)
	go shutdown()
	go shutdown()
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent shutdown returned error: %v", err)
		}
	}

	if err := <-serverDone; err != nil {
		t.Fatalf("server exited with error: %v", err)
	}
}

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
	proxy := httptest.NewTLSServer(newReverseProxyHandler(targetURL))
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
