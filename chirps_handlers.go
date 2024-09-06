package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/luispinto23/chirpy-new/internal/database"
)

type chirpDto struct {
	Body *string `json:"body,omitempty"`
}

func cleanUpBody(body string) string {
	var cleanBody []string
	forbiddenWords := []string{"kerfuffle", "sharbert", "fornax"}

	words := strings.Split(body, " ")

	for _, word := range words {
		if slices.Contains(forbiddenWords, strings.ToLower(word)) {
			cleanBody = append(cleanBody, "****")
		} else {
			cleanBody = append(cleanBody, word)
		}
	}

	return strings.Join(cleanBody, " ")
}

func (cfg *apiConfig) createChirp(w http.ResponseWriter, r *http.Request) {
	var chirp chirpDto

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&chirp)
	if err != nil {
		log.Printf("Error decoding body: %s", err)

		respondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	if chirp.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(*chirp.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanBody := cleanUpBody(*chirp.Body)

	dbChirp, err := cfg.db.CreateChirp(cleanBody)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, dbChirp)
}

func (cfg *apiConfig) getChirps(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := cfg.db.GetChirps()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, dbChirps)
}

func (cfg *apiConfig) getChirp(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	dbChirp, err := cfg.db.GetChirpByID(id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, dbChirp)
}
