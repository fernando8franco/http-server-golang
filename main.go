package main

import (
	"net/http"
)

func main() {
	serverMux := http.NewServeMux()
	serverMux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	serverMux.Handle("/app/assets/logo.png", http.StripPrefix("/app", http.FileServer(http.Dir("."))))
	serverMux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	server := http.Server{
		Handler: serverMux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}
