package main

import (
	"net/http"

	"github.com/fernando8franco/http-server-golang/internal/auth"
)

func (ac *apiConfig) refreshToken(w http.ResponseWriter, r *http.Request) {
	type response struct {
		AccessToken string `json:"token"`
	}

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find the refresh token", err)
		return
	}

	userId, err := ac.db.GetUserIdFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't retrieve the user id from the refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(userId, ac.secret, ac.expirationTime)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create the access token", err)
		return
	}

	resp := response{
		AccessToken: accessToken,
	}
	respondWithJSON(w, http.StatusOK, resp)
}

func (ac *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't find the refresh token", err)
		return
	}

	err = ac.db.SetRevokedAt(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't revoke the session", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
