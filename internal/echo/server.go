package echo

import "net/http"

// NewServer returns an HTTP server serving the echo API.
func NewServer(addr string) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: NewMux(),
	}
}
