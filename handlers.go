package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Alb3G/chirpy/internal/auth"
	"github.com/Alb3G/chirpy/internal/database"
	"github.com/google/uuid"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respondWithError(w, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed))
		return
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
		return
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

	if userReqdata.Email == "" || userReqdata.Password == "" {
		respondWithError(w, 400, "Email and password required")
		return
	}

	hash, err := auth.HashPassword(userReqdata.Password)
	if err != nil {
		respondWithError(w, 500, "Error hashing user password")
		return
	}

	userParams := database.CreateUserParams{
		Email:      userReqdata.Email,
		HashedPass: hash,
	}

	dbUser, err := ac.Queries.CreateUser(r.Context(), userParams)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 201, toUser(dbUser))
}

func (ac *apiConfig) chirpsHandler(w http.ResponseWriter, r *http.Request) {
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

	respondWithJSON(w, 201, toChirp(chirp))
}

func (ac *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	dbChirps, err := ac.Queries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	resChirpsArr := make([]Chirp, 0)
	for _, dbChirp := range dbChirps {
		resChirpsArr = append(resChirpsArr, toChirp(dbChirp))
	}

	respondWithJSON(w, 200, resChirpsArr)
}

func (ac *apiConfig) getChirpById(w http.ResponseWriter, r *http.Request) {
	chirpId := r.PathValue("chirpId")
	dbChirp, err := ac.Queries.GetChirpById(r.Context(), uuid.MustParse(chirpId))
	if err != nil {
		respondWithError(w, 404, err.Error())
		return
	}

	respondWithJSON(w, 200, toChirp(dbChirp))
}

func (ac *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var userReqdata UserRequestData
	decoder.Decode(&userReqdata)

	dbUser, err := ac.Queries.GetUserByEmail(r.Context(), userReqdata.Email)
	if err != nil {
		respondWithError(w, 401, "Failed to fetch user data by email")
		return
	}

	match, err := auth.CheckPasswordHash(userReqdata.Password, dbUser.HashedPass)
	if err != nil {
		log.Printf("Password verification error for user %s: %v", userReqdata.Email, err)
		respondWithError(w, 500, "Internal server error")
		return
	}

	if !match {
		respondWithError(w, 401, "Invalid credentials to login")
	}

	respondWithJSON(w, 200, toUser(dbUser))
}
