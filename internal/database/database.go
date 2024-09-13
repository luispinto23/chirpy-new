package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

type Chirp struct {
	Body     string `json:"body,omitempty"`
	ID       int    `json:"id,omitempty"`
	AuthorID int    `json:"author_id,omitempty"`
}

type Token struct {
	TokenExpirationDate time.Time `json:"token_expiration_date,omitempty"`
	RefreshToken        string    `json:"refresh_token,omitempty"`
	UserID              int       `json:"user_id,omitempty"`
}

type User struct {
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`
	ID       int    `json:"id,omitempty"`
}

type DB struct {
	mux  *sync.RWMutex
	path string
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
	Users  map[int]User  `json:"users"`
	Tokens map[int]Token `json:"tokens"`
}

var (
	ErrNotFound     = errors.New("record not found")
	ErrUnauthorized = errors.New("can't do that")
)

// NewDB creates a new database connection
// and creates the database file if it doesn't exist
func NewDB(path string) (*DB, error) {
	db := &DB{
		path: path,
		mux:  &sync.RWMutex{},
	}
	err := db.ensureDB()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// CreateChirp creates a new chirp and saves it to disk
func (db *DB) CreateChirp(body string, userID int) (Chirp, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStructure, err := db.loadDB()
	if err != nil {
		return Chirp{}, err
	}

	if dbStructure.Chirps == nil {
		dbStructure.Chirps = make(map[int]Chirp)
	}

	id := len(dbStructure.Chirps) + 1
	chirp := Chirp{
		ID:       id,
		Body:     body,
		AuthorID: userID,
	}
	dbStructure.Chirps[id] = chirp

	err = db.writeDB(dbStructure)
	if err != nil {
		return Chirp{}, err
	}

	return chirp, nil
}

// GetChirps returns all chirps in the database
func (db *DB) GetChirps() ([]Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	dbStructure, err := db.loadDB()
	if err != nil {
		return nil, err
	}

	// Create a slice to hold the chirps
	chirps := make([]Chirp, 0, len(dbStructure.Chirps))

	// Extract chirps from the map into the slice
	for _, chirp := range dbStructure.Chirps {
		chirps = append(chirps, chirp)
	}

	// Sort the slice based on the ID field
	sort.Slice(chirps, func(i, j int) bool {
		return chirps[i].ID < chirps[j].ID
	})

	return chirps, nil
}

// GetChirpByID returns the chirp of the given ID from the database
func (db *DB) GetChirpByID(ID int) (Chirp, error) {
	db.mux.RLock()
	defer db.mux.RUnlock()

	var chirp Chirp

	dbStructure, err := db.loadDB()
	if err != nil {
		return chirp, err
	}

	chirp, ok := dbStructure.Chirps[ID]
	if !ok {
		return chirp, ErrNotFound
	}

	return chirp, nil
}

// DeleteChirpByID returns the chirp of the given ID from the database
func (db *DB) DeleteChirpByID(ID, userID int) error {
	db.mux.RLock()
	defer db.mux.RUnlock()

	var chirp Chirp

	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	chirp, ok := dbStructure.Chirps[ID]
	if !ok {
		return ErrNotFound
	}

	if chirp.AuthorID != userID {
		return ErrUnauthorized
	}

	return nil
}

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	dbStruct := new(DBStructure)
	_, err := os.Stat(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.writeDB(*dbStruct)
	}
	return err
}

// loadDB reads the database file into memory
func (db *DB) loadDB() (DBStructure, error) {
	file, err := os.ReadFile(db.path)
	if err != nil {
		fmt.Println("ERROR ON READ")
		return DBStructure{}, err
	}

	var dbStruct DBStructure
	if len(file) == 0 {
		return dbStruct, nil
	}

	err = json.Unmarshal(file, &dbStruct)
	if err != nil {
		fmt.Println("ERROR UNMARSHALLING")

		return DBStructure{}, err
	}

	return dbStruct, nil
}

// writeDB writes the database file to disk
func (db *DB) writeDB(dbStructure DBStructure) error {
	file, err := json.Marshal(dbStructure)
	if err != nil {
		return err
	}

	return os.WriteFile(db.path, file, 0644)
}

// CreateUser creates a new user and saves it to disk
func (db *DB) CreateUser(email string, password string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if dbStructure.Users == nil {
		dbStructure.Users = make(map[int]User)
	}

	for _, user := range dbStructure.Users {
		if user.Email == email {
			return User{}, errors.New("somethig went wrong")
		}
	}

	id := len(dbStructure.Users) + 1
	user := User{
		ID:       id,
		Email:    email,
		Password: password,
	}

	dbStructure.Users[id] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

// GetUser retrieves the user for a given email
func (db *DB) GetUser(email string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if dbStructure.Users == nil {
		return User{}, ErrNotFound
	}

	for _, user := range dbStructure.Users {
		if user.Email == email {
			return user, nil
		}
	}

	return User{}, ErrNotFound
}

// UpdateUser updates a given user
func (db *DB) UpdateUser(ID int, email, password string) (User, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStructure, err := db.loadDB()
	if err != nil {
		return User{}, err
	}

	if dbStructure.Users == nil {
		return User{}, ErrNotFound
	}

	user, exists := dbStructure.Users[ID]
	if !exists {
		return User{}, ErrNotFound
	}

	// Update the user
	user.Email = email
	user.Password = password

	// Update the map
	dbStructure.Users[ID] = user

	err = db.writeDB(dbStructure)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

// GetToken get's a token based on it's value
func (db *DB) GetToken(refreshToken string) (Token, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStructure, err := db.loadDB()
	if err != nil {
		return Token{}, err
	}

	if dbStructure.Tokens == nil {
		return Token{}, ErrNotFound
	}

	for _, token := range dbStructure.Tokens {
		if token.RefreshToken == refreshToken {
			return token, nil
		}
	}

	return Token{}, ErrNotFound
}

// UpdateUserRefreshToken updates a given user refreshToken
func (db *DB) UpdateUserRefreshToken(userID int, refreshToken string, tokenExpDate time.Time) (Token, error) {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStructure, err := db.loadDB()
	if err != nil {
		return Token{}, err
	}

	if dbStructure.Tokens == nil {
		token := Token{
			UserID:              userID,
			RefreshToken:        refreshToken,
			TokenExpirationDate: tokenExpDate,
		}

		dbStructure.Tokens = make(map[int]Token)
		dbStructure.Tokens[userID] = token

		err = db.writeDB(dbStructure)
		if err != nil {
			return Token{}, err
		}

		return token, nil
	}

	dbToken := Token{}

	for _, token := range dbStructure.Tokens {
		if token.UserID == userID {
			dbToken = token
		}
	}

	if dbToken.UserID == 0 {
		token := Token{
			UserID:              userID,
			RefreshToken:        refreshToken,
			TokenExpirationDate: tokenExpDate,
		}

		dbStructure.Tokens[userID] = token

		err = db.writeDB(dbStructure)
		if err != nil {
			return Token{}, err
		}

		return token, nil
	}

	dbToken.RefreshToken = refreshToken
	dbToken.TokenExpirationDate = tokenExpDate

	dbStructure.Tokens[userID] = dbToken

	err = db.writeDB(dbStructure)
	if err != nil {
		return Token{}, err
	}

	return dbToken, nil
}

// RevokeToken revokes a given refresh token
func (db *DB) RevokeToken(refreshToken string) error {
	db.mux.Lock()
	defer db.mux.Unlock()

	dbStructure, err := db.loadDB()
	if err != nil {
		return err
	}

	if dbStructure.Tokens == nil {
		return ErrNotFound
	}

	for i, token := range dbStructure.Tokens {
		if token.RefreshToken == refreshToken {
			delete(dbStructure.Tokens, i)
			err = db.writeDB(dbStructure)
			if err != nil {
				return err
			}
			return nil
		}
	}

	return ErrNotFound
}
