package main

import (
	"encoding/json"
	"net/http"
	"strings"
)

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
