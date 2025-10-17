package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

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

func (ac *apiConfig) resetHitsHandler(w http.ResponseWriter, r *http.Request) {
	if ac.Env != "dev" {
		respondWithError(w, 403, "This endpoint is only available in dev environment")
		return
	}

	err := ac.Queries.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 200, struct {
		Result string `json:"result"`
	}{
		Result: "Database has been reset succesfully",
	})
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
		respondWithError(w, 400, "Account already exists")
		return
	}

	domain_user, err := toUser(dbUser, nil)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	respondWithJSON(w, 201, domain_user)
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

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	user_uuid, err := auth.ValidateJWT(token, ac.TokenScret)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	if len(reqData.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Something went wrong")
		return
	}

	validatedBody := validateChirpBody(reqData.Body, []string{"kerfuffle", "sharbert", "fornax"})

	chirp, err := ac.Queries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   validatedBody,
		UserID: user_uuid,
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

/*
Implement cache to store the logged user in order to avoid multiple creations
of jwt and refresh tokens in the db for the same user.
*/
func (ac *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var userReqdata UserRequestData
	decoder.Decode(&userReqdata)

	dbUser, err := ac.Queries.GetUserByEmail(r.Context(), userReqdata.Email)
	if err != nil {
		respondWithError(w, 401, "No user found for the email provided")
		return
	}

	match, err := auth.CheckPasswordHash(userReqdata.Password, dbUser.HashedPass)
	if err != nil {
		log.Printf("Password verification error for user %s: %v", userReqdata.Email, err)
		respondWithError(w, 500, "Internal server error")
		return
	}

	if !match {
		respondWithError(w, 401, "Wrong credentials")
		return
	}

	token, err := auth.MakeJWT(dbUser.ID, ac.TokenScret, time.Second*3600)
	if err != nil {
		respondWithError(w, 500, "Error while creating JWT")
		return
	}

	domainUser, err := toUser(dbUser, &token)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	refresh_token := auth.MakeRefreshToken()
	refresh_token_params := database.CreateRefreshTokenParams{
		Token:     refresh_token,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    domainUser.ID,
		ExpiresAt: time.Now().Add(DAY * 60),
	}
	_, err = ac.Queries.CreateRefreshToken(r.Context(), refresh_token_params)
	if err != nil {
		respondWithError(w, 500, err.Error())
		return
	}

	domainUser.RefreshToken = refresh_token

	respondWithJSON(w, 200, domainUser)
}

// Endpoint to refresh access tokens
func (ac *apiConfig) refreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Invalid Bearer format")
		return
	}

	db_ref_Token, err := ac.Queries.GetToken(r.Context(), refresh_token)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	if db_ref_Token.RevokedAt.Valid {
		respondWithError(w, 401, "Token is revoked")
		return
	}

	if db_ref_Token.ExpiresAt.Before(time.Now()) {
		respondWithError(w, 401, "Token is expired")
		return
	}

	user, err := ac.Queries.GetUserByToken(r.Context(), db_ref_Token.Token)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	newJwt, err := auth.MakeJWT(user.ID, ac.TokenScret, time.Hour)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	respondWithJSON(w, 200, struct {
		Token string `json:"token"`
	}{
		Token: newJwt,
	})
}

func (ac *apiConfig) revokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	err = ac.Queries.RevokeToken(r.Context(), token)
	if err != nil {
		respondWithError(w, 401, err.Error())
		return
	}

	w.WriteHeader(204)
}
