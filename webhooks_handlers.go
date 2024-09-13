package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/luispinto23/chirpy-new/internal/database"
)

type polkaDto struct {
	Event string `json:"event,omitempty"`
	Data  struct {
		UserID int `json:"user_id,omitempty"`
	} `json:"data,omitempty"`
}

func (cfg *apiConfig) polka(w http.ResponseWriter, r *http.Request) {
	apiKeyHeader := r.Header.Get("Authorization")

	if apiKeyHeader == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	apiKeyStr := strings.Split(apiKeyHeader, " ")[1]
	if cfg.polkaApiKey != apiKeyStr {
		respondWithError(w, http.StatusUnauthorized, "")
		return
	}

	var polka polkaDto

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&polka)
	if err != nil {
		log.Printf("Error decoding body: %s", err)

		respondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	if polka.Event != "user.upgraded" {
		respondWithError(w, http.StatusNoContent, "")
		return
	}

	err = cfg.db.UpgradeUser(polka.Data.UserID)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
