package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Alb3G/chirpy/internal/database"
)

const DAY = time.Hour * 24

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

func toChirp(dbc database.Chirp) Chirp {
	return Chirp{
		ID:        dbc.ID,
		CreatedAt: dbc.CreatedAt,
		UpdatedAt: dbc.UpdatedAt,
		Body:      dbc.Body,
		UserId:    dbc.UserID,
	}
}

func toUser(dbu database.User, token *string) (User, error) {
	tokenValue := ""

	if token != nil {
		tokenValue = *token
	}

	return User{
		ID:           dbu.ID,
		CreatedAt:    dbu.CreatedAt,
		UpdatedAt:    dbu.UpdatedAt,
		Email:        dbu.Email,
		Token:        tokenValue,
		RefreshToken: "",
		IsChirpyRed:  dbu.IsChirpyRed.Bool,
	}, nil
}
