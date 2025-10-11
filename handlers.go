package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverhits atomic.Int32
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Valid bool `json:"valid"`
}

type ChirpBody struct {
	Body string `json:"body"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	w.WriteHeader(http.StatusOK)

	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Something went wrong"})
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var body ChirpBody
	decoder.Decode(&body)

	if len(body.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Error: "Chirp is too long"})
		return
	}

	json.NewEncoder(w).Encode(SuccessResponse{Valid: true})
}

func (ac *apiConfig) hitsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html; charset=utf-8")

	w.WriteHeader(http.StatusOK)

	template := `<html>
  		<body>
    		<h1>Welcome, Chirpy Admin</h1>
    		<p>Chirpy has been visited %d times!</p>
  		</body>
	</html>`

	fmt.Fprintf(w, template, ac.fileserverhits.Load())
}

func (ac *apiConfig) metricsCountMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ac.fileserverhits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (ac *apiConfig) resetHitsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	ac.fileserverhits.Store(0)

	w.WriteHeader(http.StatusOK)

	w.Write([]byte("Hits to file have been reset!"))
}
