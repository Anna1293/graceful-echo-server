package echo

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// NewMux registers the echo HTTP handlers.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/echo", echoHandler)
	return mux
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "use GET", http.StatusMethodNotAllowed)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok\n"))
}

func echoHandler(w http.ResponseWriter, r *http.Request) {
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

	select {
	case <-time.After(time.Duration(delaySec) * time.Second):
		_, _ = fmt.Fprintf(w, "echo: %s\n", msg)
	case <-r.Context().Done():
		http.Error(w, "request canceled", http.StatusRequestTimeout)
	}
}
