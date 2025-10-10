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

	mux.HandleFunc("GET /healthz", healthHandler)
	mux.HandleFunc("GET /metrics", apiCfg.hitsHandler)
	mux.HandleFunc("POST /reset", apiCfg.resetHitsHandler)

	fileServer := http.FileServer(http.Dir(FILE_PATH_ROOT))
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.metricsCountMiddleware(fileServer)))

	s := &http.Server{
		Addr:    ":" + PORT,
		Handler: mux,
	}

	fmt.Printf("Server running on port: 8080\n")

	log.Fatal(s.ListenAndServe())
}
