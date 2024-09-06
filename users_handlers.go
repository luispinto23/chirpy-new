package main

import (
	"encoding/json"
	"log"
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

type loginReq struct {
	Email            *string `json:"email,omitempty"`
	Password         *string `json:"password,omitempty"`
	ExpiresInSeconds int     `json:"expires_in_seconds,omitempty"`
}

type userDto struct {
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
	ID       int     `json:"id,omitempty"`
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
		ID:       dbUser.ID,
		Email:    &dbUser.Email,
		Password: nil,
	}
	respondWithJSON(w, http.StatusCreated, response)
}

func (cfg *apiConfig) login(w http.ResponseWriter, r *http.Request) {
	var req loginReq

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		log.Printf("Error decoding body: %s", err)

		respondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	if req.Email == nil || req.Password == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	dbUser, err := cfg.db.GetUser(*req.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(*req.Password))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	response := userDto{
		ID:       dbUser.ID,
		Email:    &dbUser.Email,
		Password: nil,
	}

	respondWithJSON(w, http.StatusOK, response)
}
