package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/fernando8franco/http-server-golang/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
	expirationTime time.Duration
}

type ErrorMessage struct {
	Error string `json:"error"`
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	secret := os.Getenv("SECRET")
	if secret == "" {
		log.Fatal("SECRET must be set")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	dbQueries := database.New(db)

	apiCfg := &apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		secret:         secret,
		expirationTime: time.Hour,
	}

	serverMux := http.NewServeMux()

	handler := http.StripPrefix("/app", http.FileServer(http.Dir(".")))
	serverMux.Handle("GET /app/", apiCfg.middlewareMetricsInc(handler))

	serverMux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	serverMux.HandleFunc("POST /api/users", apiCfg.createUser)
	serverMux.HandleFunc("PUT /api/users", apiCfg.updateUser)
	serverMux.HandleFunc("POST /api/login", apiCfg.loginUser)

	serverMux.HandleFunc("POST /api/refresh", apiCfg.refreshToken)
	serverMux.HandleFunc("POST /api/revoke", apiCfg.revokeToken)

	serverMux.HandleFunc("POST /api/chirps", apiCfg.createChirp)
	serverMux.HandleFunc("GET /api/chirps", apiCfg.getAllChirps)
	serverMux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.getOneChirp)
	serverMux.HandleFunc("DELETE /api/chirps/{chirpId}", apiCfg.deleteChirp)

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
	if ac.platform != "dev" {
		respondWithError(w, http.StatusForbidden, "Reset is only allowed in dev environment", nil)
		return
	}

	ac.fileserverHits.Store(0)
	err := ac.db.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't delete the users", err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
