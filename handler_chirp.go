package main

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

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
		Body   string    `json:"body"`
		UserId uuid.UUID `json:"user_id"`
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

	if len(params.Body) > maxChirpLength {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long", nil)
		return
	}
	cleanBody := replaceWords(params.Body)

	chirp, err := ac.db.CreateChirp(
		r.Context(),
		database.CreateChirpParams{
			Body:   cleanBody,
			UserID: params.UserId,
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
