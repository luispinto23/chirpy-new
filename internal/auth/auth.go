package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const JwtExpirationSeconds = 360

type RefreshToken struct {
	TokenExpDate time.Time
	Token        string
}

func ValidateToken(password, reqPassword []byte) error {
	err := bcrypt.CompareHashAndPassword(password, reqPassword)
	if err != nil {
		return err
	}
	return nil
}

func IssueJWT(userID int, jwtSecret string) (string, error) {
	var token string
	now := time.Now().UTC()
	// Create a NumericDate from the current time
	numericNow := jwt.NewNumericDate(now)

	expirationDate := now.Add(time.Duration(JwtExpirationSeconds) * time.Second)
	numericExp := jwt.NewNumericDate(expirationDate)

	jwt := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  numericNow,
		ExpiresAt: numericExp,
		Subject:   strconv.Itoa(userID),
	})

	token, err := jwt.SignedString([]byte(jwtSecret))
	if err != nil {
		return token, err
	}
	return token, nil
}

func GenerateRefreshToken() (RefreshToken, error) {
	var token RefreshToken
	now := time.Now().UTC()

	c := 32
	randB := make([]byte, c)
	_, err := rand.Read(randB)
	if err != nil {
		return token, err
	}

	token.Token = hex.EncodeToString(randB)
	token.TokenExpDate = now.Add(time.Duration(60*24) * time.Hour)

	return token, nil
}

func GenerateHashedPass(password string) ([]byte, error) {
	pass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return pass, err
}

func ValidateJWTToken(tokenStr, jwtSecret string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &jwt.RegisteredClaims{
		Issuer: "chirpy",
	}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return token, nil
}
