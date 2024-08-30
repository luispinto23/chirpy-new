package main

import (
	"fmt"
	"log"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	resp := fmt.Sprintf("Hits: %d", cfg.fileServerHits)
	w.Write([]byte(resp))
}

func (cfg *apiConfig) resetMetrics(w http.ResponseWriter, r *http.Request) {
	cfg.fileServerHits = 0
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

type apiConfig struct {
	fileServerHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits++
		next.ServeHTTP(w, r)
	})
}

func fileServerHandler(toStrip, filepathRoot string) http.Handler {
	return http.StripPrefix(toStrip, http.FileServer(http.Dir(filepathRoot)))
}

func main() {
	mux := http.NewServeMux()

	apicfg := apiConfig{
		fileServerHits: 0,
	}

	srv := http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	rootFilePath := "."
	appPath := "/app/"
	fileHandler := fileServerHandler("/app", rootFilePath)
	mux.Handle(appPath, apicfg.middlewareMetricsInc(fileHandler))

	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/metrics", apicfg.metricsHandler)
	mux.HandleFunc("/reset", apicfg.resetMetrics)

	log.Fatal(srv.ListenAndServe())
}
