package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/luispinto23/chirpy-new/internal/database"
)

type apiConfig struct {
	db             database.DB
	jwtSecret      string
	fileServerHits int
}

func main() {
	godotenv.Load()
	mux := http.NewServeMux()

	db, err := database.NewDB("database.json")
	if err != nil {
		log.Fatal(err)
	}
	jwtSecret := os.Getenv("JWT_SECRET")

	apicfg := apiConfig{
		fileServerHits: 0,
		db:             *db,
		jwtSecret:      jwtSecret,
	}

	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	rootFilePath := "."
	appPath := "/app/"
	fileHandler := fileServerHandler("/app", rootFilePath)
	mux.Handle(appPath, apicfg.middlewareMetricsInc(fileHandler))

	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("GET /admin/metrics", apicfg.metricsHandler)
	mux.HandleFunc("GET /api/reset", apicfg.resetMetrics)
	mux.HandleFunc("POST /api/chirps", apicfg.createChirp)
	mux.HandleFunc("GET /api/chirps", apicfg.getChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apicfg.getChirp)
	mux.HandleFunc("POST /api/users", apicfg.createUser)
	mux.HandleFunc("PUT /api/users", apicfg.updateUser)
	mux.HandleFunc("POST /api/login", apicfg.login)
	mux.HandleFunc("POST /api/refresh", apicfg.refreshToken)
	mux.HandleFunc("POST /api/revoke", apicfg.revokeToken)

	log.Fatal(srv.ListenAndServe())
}
