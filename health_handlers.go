package main

import (
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func fileServerHandler(toStrip, filepathRoot string) http.Handler {
	return http.StripPrefix(toStrip, http.FileServer(http.Dir(filepathRoot)))
}
