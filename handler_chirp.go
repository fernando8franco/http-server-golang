package main

import (
	"encoding/json"
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

	accessToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find the refresh token", err)
		return
	}

	userId, err := auth.ValidateJWT(accessToken, ac.secret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate the token", err)
		return
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode the parameters", err)
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
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve the chirps", err)
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
		respondWithError(w, http.StatusInternalServerError, "Invalid chirp Id", err)
		return
	}

	type response struct {
		Chirp
	}

	chirp, err := ac.db.GetOneChirp(r.Context(), chirpId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't retrieve the chirp", err)
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
	rWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}

	words := strings.Split(msg, " ")
	for i, word := range words {
		lower := strings.ToLower(word)
		if _, ok := rWords[lower]; ok {
			words[i] = "****"
		}
	}

	newMsg = strings.Join(words, " ")
	return newMsg
}
