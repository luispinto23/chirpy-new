package main

import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type userDto struct {
	ID       int     `json:"id,omitempty"`
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
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

	if user.Email == nil || user.Password == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(*user.Password), bcrypt.DefaultCost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	dbUser, err := cfg.db.CreateUser(*user.Email, string(pass))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := userDto{
		dbUser.ID,
		&dbUser.Email,
		nil,
	}
	respondWithJSON(w, http.StatusCreated, response)
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	var user userDto

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		log.Printf("Error decoding body: %s", err)

		respondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	if user.Email == nil || user.Password == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dbUser, err := cfg.db.GetUser(*user.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(*user.Password))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	response := userDto{
		dbUser.ID,
		&dbUser.Email,
		nil,
	}

	respondWithJSON(w, http.StatusOK, response)
}
