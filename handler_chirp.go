package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/fernando8franco/http-server-golang/internal/auth"
	"github.com/fernando8franco/http-server-golang/internal/database"
	"github.com/google/uuid"
)

type Chirp struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

func (ac *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	const maxChirpLength = 140
	type parameters struct {
		Body string `json:"body"`
	}
	type response struct {
		Chirp
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "The Authorization header don't exist", err)
		return
	}

	userId, err := auth.ValidateJWT(token, ac.secret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't validate the token", err)
		return
	}

	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	cleanBody := replaceWords(params.Body)

	chirp, err := ac.db.CreateChirp(
		r.Context(),
		database.CreateChirpParams{
			Body:   cleanBody,
			UserID: userId,
		},
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create the chirp", err)
		return
	}

	resp := response{
		Chirp: Chirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		},
	}

	respondWithJSON(w, http.StatusCreated, resp)
}

func (ac *apiConfig) getAllChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := ac.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps", err)
		return
	}

	resp := []Chirp{}

	for _, chirp := range chirps {
		resp = append(
			resp,
			Chirp{
				Id:        chirp.ID,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
				Body:      chirp.Body,
				UserId:    chirp.UserID,
			},
		)
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (ac *apiConfig) getOneChirp(w http.ResponseWriter, r *http.Request) {
	chirpIdStr := r.PathValue("chirpId")
	chirpId, err := uuid.Parse(chirpIdStr)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "The chirp UUID is not valid", err)
		return
	}

	type response struct {
		Chirp
	}

	chirp, err := ac.db.GetOneChirp(r.Context(), chirpId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "The chirp doesn't exist", err)
			return
		}

		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirp", err)
		return
	}

	resp := response{
		Chirp: Chirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserId:    chirp.UserID,
		},
	}

	respondWithJSON(w, http.StatusOK, resp)
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
