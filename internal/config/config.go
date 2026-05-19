package config

import (
	"fmt"
	"net/url"
	"os"
	"time"
)

// Config holds runtime settings (env vars override defaults).
type Config struct {
	BackendAddr      string
	ProxyAddr        string
	CertFile         string
	KeyFile          string
	ShutdownTimeout  time.Duration
	UpstreamTimeout  time.Duration
}

// Load reads configuration from environment variables.
func Load() (Config, error) {
	cfg := Config{
		BackendAddr:     envOr("BACKEND_ADDR", ":8080"),
		ProxyAddr:       envOr("PROXY_ADDR", ":8443"),
		CertFile:        envOr("TLS_CERT_FILE", "certs/server.crt"),
		KeyFile:         envOr("TLS_KEY_FILE", "certs/server.key"),
		ShutdownTimeout: 10 * time.Second,
		UpstreamTimeout: 30 * time.Second,
	}

	if cfg.BackendAddr == "" {
		return Config{}, fmt.Errorf("BACKEND_ADDR must not be empty")
	}
	if cfg.ProxyAddr == "" {
		return Config{}, fmt.Errorf("PROXY_ADDR must not be empty")
	}

	return cfg, nil
}

// UpstreamURL builds the backend URL for the reverse proxy.
func (c Config) UpstreamURL() (*url.URL, error) {
	host := c.BackendAddr
	if host[0] == ':' {
		host = "127.0.0.1" + host
	}
	return url.Parse("http://" + host)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
