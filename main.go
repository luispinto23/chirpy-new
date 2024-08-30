package main

import (
	"log"
	"net/http"
)

type apiConfig struct {
	fileServerHits int
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

	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("GET /admin/metrics", apicfg.metricsHandler)
	mux.HandleFunc("GET /api/reset", apicfg.resetMetrics)

	log.Fatal(srv.ListenAndServe())
}
