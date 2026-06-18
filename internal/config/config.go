package config

import (
	"fmt"
	"net/url"
	"os"
	"time"
)

// Config — параметры запуска (переменные окружения переопределяют значения по умолчанию).
type Config struct {
	BackendAddr      string
	ProxyAddr        string
	CertFile         string
	KeyFile          string
	ShutdownTimeout  time.Duration
	UpstreamTimeout  time.Duration
}

// Load читает конфигурацию из переменных окружения.
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
		return Config{}, fmt.Errorf("BACKEND_ADDR не должен быть пустым")
	}
	if cfg.ProxyAddr == "" {
		return Config{}, fmt.Errorf("PROXY_ADDR не должен быть пустым")
	}

	return cfg, nil
}

// UpstreamURL формирует URL backend для reverse proxy.
func (c Config) UpstreamURL() (*url.URL, error) {
	host := c.BackendAddr
	if host[0] == ':' {
		host = "127.0.0.1" + host
	}
	return url.Parse("http://" + host)
}

// envOr возвращает значение переменной окружения key или fallback, если она не задана.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
