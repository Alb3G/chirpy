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

	mux.HandleFunc("/healthz", healthHandler)
	mux.HandleFunc("/metrics", apiCfg.hitsHandler)
	mux.HandleFunc("/reset", apiCfg.resetHitsHandler)

	fileServer := http.FileServer(http.Dir(FILE_PATH_ROOT))
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.metricsCountMiddleware(fileServer)))

	s := &http.Server{
		Addr:    ":" + PORT,
		Handler: mux,
	}

	fmt.Printf("Server running on port: 8080\n")

	log.Fatal(s.ListenAndServe())
}
