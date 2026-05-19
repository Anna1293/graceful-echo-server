package proxy

import (
	"io"
	"net/http"
	"net/url"
	"time"
)

// Options configures the reverse proxy handler.
type Options struct {
	UpstreamTimeout time.Duration
	ExternalProto   string // e.g. "https" for TLS front door
}

// NewHandler returns an http.Handler that forwards requests to target.
func NewHandler(target *url.URL, opts Options) http.Handler {
	if opts.UpstreamTimeout <= 0 {
		opts.UpstreamTimeout = 30 * time.Second
	}
	if opts.ExternalProto == "" {
		opts.ExternalProto = "https"
	}

	client := &http.Client{
		Timeout: opts.UpstreamTimeout,
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
		setForwardedHeaders(req, r, opts.ExternalProto)
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
