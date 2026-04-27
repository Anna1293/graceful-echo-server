package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

func main() {
	// Пример запроса: /echo?msg=hello&delay=3
	http.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) {
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

	server := &http.Server{Addr: ":8080"}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Graceful shutdown: перестаем принимать новые запросы и ждем текущие.
	go func() {
		<-stop
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = server.Shutdown(ctx)
	}()

	log.Println("server started on :8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
	log.Println("server stopped")
}
