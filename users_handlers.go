package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type userDto struct {
	Email *string `json:"email,omitempty"`
}

func (cfg *apiConfig) createUser(w http.ResponseWriter, r *http.Request) {
	var user userDto

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		log.Printf("Error decoding body: %s", err)

		respondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	if user.Email == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dbChirp, err := cfg.db.CreateUser(*user.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, dbChirp)
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {}
