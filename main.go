package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	FILE_PATH_ROOT = "."
	PORT           = "8080"
)

func main() {
	mux := http.NewServeMux()

	apiCfg := &apiConfig{}

	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.hitsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHitsHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	fileServer := http.FileServer(http.Dir(FILE_PATH_ROOT))
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.metricsCountMiddleware(fileServer)))

	s := &http.Server{
		Addr:    ":" + PORT,
		Handler: mux,
	}

	fmt.Printf("Server running on port: 8080\n")

	log.Fatal(s.ListenAndServe())
}
