package proxy

import (
	"net/http"
	"strings"
)

var hopByHopHeaders = []string{
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

func removeHopByHopHeaders(h http.Header) {
	for _, key := range hopByHopHeaders {
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

func setForwardedHeaders(dst *http.Request, client *http.Request, externalProto string) {
	clientIP := clientIPFromRequest(client)
	if prior := client.Header.Get("X-Forwarded-For"); prior != "" {
		dst.Header.Set("X-Forwarded-For", prior+", "+clientIP)
	} else {
		dst.Header.Set("X-Forwarded-For", clientIP)
	}
	dst.Header.Set("X-Forwarded-Host", client.Host)
	dst.Header.Set("X-Forwarded-Proto", externalProto)
}

func clientIPFromRequest(r *http.Request) string {
	if host, _, ok := strings.Cut(r.RemoteAddr, ":"); ok {
		return host
	}
	return r.RemoteAddr
}
