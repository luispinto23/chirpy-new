package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/luispinto23/chirpy-new/internal/auth"
)

type loginReq struct {
	Email    *string `json:"email,omitempty"`
	Password *string `json:"password,omitempty"`
}

type tokenDto struct {
	Token string `json:"token,omitempty"`
}

type userDto struct {
	Email        *string `json:"email,omitempty"`
	Password     *string `json:"password,omitempty"`
	Token        string  `json:"token,omitempty"`
	RefreshToken string  `json:"refresh_token,omitempty"`
	ID           int     `json:"id,omitempty"`
	IsChirpyRed  bool    `json:"is_chirpy_red"`
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

	pass, err := auth.GenerateHashedPass(*user.Password)
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
		ID:          dbUser.ID,
		Email:       &dbUser.Email,
		Password:    nil,
		IsChirpyRed: dbUser.IsChirpyRed,
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

	err = auth.ValidateToken([]byte(dbUser.Password), []byte(*req.Password))
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	signedToken, err := auth.IssueJWT(dbUser.ID, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	refreshTokenDb, err := cfg.db.UpdateUserRefreshToken(dbUser.ID, refreshToken.Token, refreshToken.TokenExpDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := userDto{
		ID:           dbUser.ID,
		Email:        &dbUser.Email,
		Password:     nil,
		IsChirpyRed:  dbUser.IsChirpyRed,
		Token:        signedToken,
		RefreshToken: refreshTokenDb.RefreshToken,
	}

	respondWithJSON(w, http.StatusOK, response)
}

func (cfg *apiConfig) updateUser(w http.ResponseWriter, r *http.Request) {
	authReqHeader := r.Header.Get("Authorization")

	if authReqHeader == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tokenStr := strings.Split(authReqHeader, " ")[1]

	token, err := auth.ValidateJWTToken(tokenStr, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		respondWithError(w, http.StatusInternalServerError, "Couldn't parse claims")
		return
	}

	userID, err := claims.GetSubject()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	var user userDto

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&user)
	if err != nil {
		log.Printf("Error decoding body: %s", err)

		respondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	if user.Email == nil || user.Password == nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pass, err := auth.GenerateHashedPass(*user.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// userID to int
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	updatedUser, err := cfg.db.UpdateUser(userIDInt, *user.Email, string(pass))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := userDto{
		ID:          updatedUser.ID,
		Email:       &updatedUser.Email,
		IsChirpyRed: updatedUser.IsChirpyRed,
		Password:    nil,
	}
	respondWithJSON(w, http.StatusOK, response)
}
