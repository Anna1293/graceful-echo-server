package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"graceful-echo-server/internal/certs"
	"graceful-echo-server/internal/config"
	"graceful-echo-server/internal/echo"
	"graceful-echo-server/internal/proxy"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if err := certs.EnsureSelfSigned(cfg.CertFile, cfg.KeyFile); err != nil {
		log.Fatalf("tls certs: %v", err)
	}

	targetURL, err := cfg.UpstreamURL()
	if err != nil {
		log.Fatalf("upstream URL: %v", err)
	}

	backendServer := echo.NewServer(cfg.BackendAddr)
	proxyServer := proxy.NewTLSServer(
		cfg.ProxyAddr,
		targetURL,
		cfg.CertFile,
		cfg.KeyFile,
		proxy.Options{UpstreamTimeout: cfg.UpstreamTimeout},
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	backendErrCh := make(chan error, 1)
	go func() {
		log.Printf("backend started on %s", cfg.BackendAddr)
		if err := backendServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			backendErrCh <- err
		}
	}()

	proxyErrCh := make(chan error, 1)
	go func() {
		log.Printf("reverse proxy started on %s (TLS + HTTP/2)", cfg.ProxyAddr)
		if err := proxyServer.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile); err != nil && err != http.ErrServerClosed {
			proxyErrCh <- err
		}
	}()

	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
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
