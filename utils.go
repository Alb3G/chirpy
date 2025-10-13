package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

func respondWithError(w http.ResponseWriter, statusCode int, errMsg string) {
	if statusCode > 499 {
		log.Println("Internal server Error")
	}

	respondWithJSON(w, statusCode, ErrorResponse{Error: errMsg})
}

func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-type", "application/json")

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error parsing json object: %v", err)
		w.WriteHeader(500)
		return
	}

	w.WriteHeader(statusCode)

	w.Write(data)
}

func validateChirpBody(body string, bannedWords []string) string {
	words := strings.Fields(body)

	for i, word := range words {
		for _, bannedWord := range bannedWords {
			if strings.EqualFold(word, bannedWord) {
				words[i] = "****"
			}
		}
	}
	return strings.Join(words, " ")
}
