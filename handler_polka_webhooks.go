package main

import (
	"encoding/json"
	"net/http"

	"github.com/fernando8franco/http-server-golang/internal/auth"
	"github.com/google/uuid"
)

func (ac *apiConfig) polkaWebhook(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Event string `json:"event"`
		Data  struct {
			UserId string `json:"user_id"`
		} `json:"data"`
	}

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find the api key", err)
		return
	}

	if apiKey != ac.polkaKey {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate the api key", nil)
		return
	}

	params := parameters{}
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode the parameters", err)
		return
	}

	userId, err := uuid.Parse(params.Data.UserId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Invalid user", err)
		return
	}

	if params.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_, err = ac.db.UpdateUserToChirpyRed(r.Context(), userId)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't validate the user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
