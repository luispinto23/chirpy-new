package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"sync"
)

type Chirp struct {
	ID   int    `json:"id,omitempty"`
	Body string `json:"body,omitempty"`
}

type DB struct {
	path string
	mux  *sync.RWMutex
}

type DBStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

var ErrNotFound = errors.New("record not found")

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
func (db *DB) CreateChirp(body string) (Chirp, error) {
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
		ID:   id,
		Body: body,
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

// GetChirp returns the chirp of the given ID from the database
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

// ensureDB creates a new database file if it doesn't exist
func (db *DB) ensureDB() error {
	dbStruct := DBStructure{
		Chirps: make(map[int]Chirp),
	}
	_, err := os.Stat(db.path)
	if errors.Is(err, os.ErrNotExist) {
		return db.writeDB(dbStruct)
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
