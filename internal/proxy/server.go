package proxy

import (
	"crypto/tls"
	"net/http"
	"net/url"
)

// NewTLSServer builds an HTTPS server that terminates TLS and proxies to target.
func NewTLSServer(addr string, target *url.URL, certFile, keyFile string, opts Options) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: NewHandler(target, opts),
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
			NextProtos: []string{"h2", "http/1.1"},
		},
	}
}
