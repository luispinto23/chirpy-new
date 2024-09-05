package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"strings"
)

type chirp struct {
	Body *string `json:"body,omitempty"`
}

type resp struct {
	Error       string `json:"error,omitempty"`
	CleanedBody string `json:"cleaned_body,omitempty"`
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

func respondWithError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	respBody := resp{
		Error: msg,
	}

	data, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (cfg *apiConfig) validateChirp(w http.ResponseWriter, r *http.Request) {
	var chirp chirp

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

	respBody := resp{
		CleanedBody: cleanBody,
	}

	respondWithJSON(w, http.StatusOK, respBody)
}
