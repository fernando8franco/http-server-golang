package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	serverMux := http.NewServeMux()
	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	apiCfg := &apiConfig{}

	serverMux.Handle("GET /app", apiCfg.middlewareMetricsInc(handler))
	serverMux.Handle("GET /app/assets/logo.png", apiCfg.middlewareMetricsInc(handler))
	serverMux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	serverMux.HandleFunc("GET /metrics", apiCfg.metrics)
	serverMux.HandleFunc("POST /reset", apiCfg.reset)

	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}

func (ac *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ac.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (ac *apiConfig) metrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Hits: %v", ac.fileserverHits.Load())
}

func (ac *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	ac.fileserverHits.Store(0)
}
