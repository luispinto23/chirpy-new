package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
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
}

const jwtExpirationSeconds = 360

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
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	now := time.Now().UTC()

	// Create a NumericDate from the current time
	numericNow := jwt.NewNumericDate(now)

	expirationDate := now.Add(time.Duration(jwtExpirationSeconds) * time.Second)
	numericExp := jwt.NewNumericDate(expirationDate)

	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  numericNow,
		ExpiresAt: numericExp,
		Subject:   strconv.Itoa(dbUser.ID),
	})

	c := 32
	randB := make([]byte, c)
	_, err = rand.Read(randB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "")
		return
	}

	refreshToken := hex.EncodeToString(randB)
	refreshExpDate := now.Add(time.Duration(60*24) * time.Hour)

	refreshTokenDb, err := cfg.db.UpdateUserRefreshToken(dbUser.ID, refreshToken, refreshExpDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	signedToken, err := jwt.SignedString([]byte(cfg.jwtSecret))
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := userDto{
		ID:           dbUser.ID,
		Email:        &dbUser.Email,
		Password:     nil,
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

	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{
		Issuer: "chirpy",
	}, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.jwtSecret), nil
	})
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	if !token.Valid {
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

	pass, err := bcrypt.GenerateFromPassword([]byte(*user.Password), bcrypt.DefaultCost)
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
		ID:       updatedUser.ID,
		Email:    &updatedUser.Email,
		Password: nil,
	}
	respondWithJSON(w, http.StatusOK, response)
}

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

	c := 32
	randB := make([]byte, c)
	_, err = rand.Read(randB)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "")
		return
	}

	now := time.Now().UTC()

	refreshToken := hex.EncodeToString(randB)
	refreshExpDate := now.Add(time.Duration(60*24) * time.Hour)

	_, err = cfg.db.UpdateUserRefreshToken(dbToken.UserID, refreshToken, refreshExpDate)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response := tokenDto{
		Token: refreshToken,
	}

	respondWithJSON(w, http.StatusOK, response)
}
