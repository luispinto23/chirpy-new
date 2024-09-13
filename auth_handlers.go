package main

import (
	"net/http"
	"strings"

	"github.com/luispinto23/chirpy-new/internal/auth"
)

func (cfg *apiConfig) refreshToken(w http.ResponseWriter, r *http.Request) {
	authReqHeader := r.Header.Get("Authorization")

	if authReqHeader == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenStr := strings.Split(authReqHeader, " ")[1]

	dbToken, err := cfg.db.GetToken(tokenStr)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	signedToken, err := auth.IssueJWT(dbToken.UserID, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := tokenDto{
		Token: signedToken,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) revokeToken(w http.ResponseWriter, r *http.Request) {
	authReqHeader := r.Header.Get("Authorization")

	if authReqHeader == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenStr := strings.Split(authReqHeader, " ")[1]

	err := cfg.db.RevokeToken(tokenStr)
	if err != nil {
		respondWithJSON(w, http.StatusNoContent, nil)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
