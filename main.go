package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func newEchoMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Пример запроса: /echo?msg=hello&delay=3
	mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "use GET", http.StatusMethodNotAllowed)
			return
		}

		msg := r.URL.Query().Get("msg")
		if msg == "" {
			msg = "echo"
		}

		delaySec := 5
		if raw := r.URL.Query().Get("delay"); raw != "" {
			sec, err := strconv.Atoi(raw)
			if err != nil || sec < 0 {
				http.Error(w, "delay must be a non-negative integer", http.StatusBadRequest)
				return
			}
			delaySec = sec
		}

		// Имитируем долгую работу и прерываем ее при отмене запроса клиентом.
		select {
		case <-time.After(time.Duration(delaySec) * time.Second):
			_, _ = fmt.Fprintf(w, "echo: %s\n", msg)
		case <-r.Context().Done():
			http.Error(w, "request canceled", http.StatusRequestTimeout)
		}
	})

	return mux
}

func newServer(addr string) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: newEchoMux(),
	}
}

func removeHopByHopHeaders(h http.Header) {
	hopByHop := []string{
		"Connection",
		"Proxy-Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailer",
		"Transfer-Encoding",
		"Upgrade",
	}
	for _, key := range hopByHop {
		h.Del(key)
	}
}

func copyHeaders(dst, src http.Header) {
	for key, values := range src {
		for _, v := range values {
			dst.Add(key, v)
		}
	}
}

func newReverseProxyHandler(target *url.URL) http.Handler {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			ForceAttemptHTTP2:   true,
			MaxIdleConns:        100,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstreamURL := *target
		upstreamURL.Path = r.URL.Path
		upstreamURL.RawPath = r.URL.RawPath
		upstreamURL.RawQuery = r.URL.RawQuery

		req, err := http.NewRequestWithContext(r.Context(), r.Method, upstreamURL.String(), r.Body)
		if err != nil {
			http.Error(w, "failed to build upstream request", http.StatusBadGateway)
			return
		}

		copyHeaders(req.Header, r.Header)
		removeHopByHopHeaders(req.Header)

		if host, _, ok := strings.Cut(r.RemoteAddr, ":"); ok {
			req.Header.Set("X-Forwarded-For", host)
		}
		req.Header.Set("X-Forwarded-Host", r.Host)
		req.Header.Set("X-Forwarded-Proto", "https")
		req.Host = target.Host

		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "upstream unavailable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		removeHopByHopHeaders(resp.Header)
		copyHeaders(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	})
}

func newProxyServer(addr string, target *url.URL) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: newReverseProxyHandler(target),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			NextProtos: []string{"h2", "http/1.1"},
		},
	}
}

func main() {
	backendAddr := ":8080"
	proxyAddr := ":8443"
	certFile := "certs/server.crt"
	keyFile := "certs/server.key"

	backendServer := newServer(backendAddr)
	targetURL, err := url.Parse("http://127.0.0.1" + backendAddr)
	if err != nil {
		log.Fatalf("invalid upstream URL: %v", err)
	}
	proxyServer := newProxyServer(proxyAddr, targetURL)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	backendErrCh := make(chan error, 1)
	go func() {
		log.Printf("backend started on %s", backendAddr)
		if err := backendServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			backendErrCh <- err
		}
	}()

	proxyErrCh := make(chan error, 1)
	go func() {
		log.Printf("reverse proxy started on %s (TLS + HTTP/2)", proxyAddr)
		if err := proxyServer.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
			proxyErrCh <- err
		}
	}()

	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_ = proxyServer.Shutdown(ctx)
		_ = backendServer.Shutdown(ctx)
	}()

	select {
	case err := <-backendErrCh:
		log.Fatalf("backend server error: %v", err)
	case err := <-proxyErrCh:
		log.Fatalf("proxy server error: %v", err)
	case <-stop:
	}

	log.Println("servers stopped")
}
