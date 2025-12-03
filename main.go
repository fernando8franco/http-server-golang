package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/fernando8franco/http-server-golang/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
}

type ErrorMessage struct {
	Error string `json:"error"`
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(db)

	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
	}

	serverMux := http.NewServeMux()

	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serverMux.Handle("GET /app/", apiCfg.middlewareMetricsInc(handler))

	serverMux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	serverMux.HandleFunc("POST /api/validate_chirp", validateChirp)

	serverMux.HandleFunc("GET /admin/metrics", apiCfg.metrics)
	serverMux.HandleFunc("POST /admin/reset", apiCfg.reset)

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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", ac.fileserverHits.Load())
}

func (ac *apiConfig) reset(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	ac.fileserverHits.Store(0)
}

func validateChirp(w http.ResponseWriter, r *http.Request) {
	const maxChirpLength = 140
	chirp := struct {
		Body string `json:"body"`
	}{}

	if err := json.NewDecoder(r.Body).Decode(&chirp); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	if len(chirp.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	cleanBody := replaceWords(chirp.Body)

	cleanChip := struct {
		CleanedBody string `json:"cleaned_body"`
	}{
		CleanedBody: cleanBody,
	}
	respondWithJSON(w, http.StatusOK, cleanChip)
}

func replaceWords(msg string) (newMsg string) {
	rWords := map[string]bool{
		"kerfuffle": true,
		"sharbert":  true,
		"fornax":    true,
	}

	words := strings.Split(msg, " ")
	for i, word := range words {
		lower := strings.ToLower(word)
		if rWords[lower] {
			words[i] = "****"
		}
	}

	newMsg = strings.Join(words, " ")
	return newMsg
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	if err != nil {
		log.Println(err)
	}
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(code)
	w.Write(data)
}
