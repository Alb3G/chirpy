package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Alb3G/chirpy/internal/database"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	w.WriteHeader(http.StatusOK)

	w.Write([]byte(http.StatusText(http.StatusOK)))
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

// Modificar response para devolver un json en lugar de texto plano.
func (ac *apiConfig) resetHitsHandler(w http.ResponseWriter, r *http.Request) {
	if ac.Env != "dev" {
		respondWithError(w, 403, "This endpoint is only available in dev environment")
	}

	ac.Queries.DeleteUsers(r.Context())

	w.WriteHeader(http.StatusOK)

	w.Write([]byte("Users db has been reset."))
}

func (ac *apiConfig) usersHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var userReqdata UserRequestData
	decoder.Decode(&userReqdata)

	dbUser, err := ac.Queries.CreateUser(r.Context(), userReqdata.Email)
	if err != nil {
		respondWithError(w, 500, err.Error())
	}

	user := User{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}

	respondWithJSON(w, 201, user)
}

func (ac *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Header.Get("Content-Type") != "application/json" {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var reqData ChirpBody
	decoder.Decode(&reqData)

	if len(reqData.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	validatedBody := validateChirpBody(reqData.Body, []string{"kerfuffle", "sharbert", "fornax"})

	chirp, err := ac.Queries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   validatedBody,
		UserID: reqData.UserId,
	})
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 201, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserId:    chirp.UserID,
	})
}
