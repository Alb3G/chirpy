package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Alb3G/chirpy/internal/database"
	"github.com/didip/tollbooth/v7"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const (
	FILE_PATH_ROOT = "."
	PORT           = "8080"
)

func main() {
	godotenv.Load()

	dbUrl := os.Getenv("DB_URL")
	env := os.Getenv("ENV")
	secret := os.Getenv("TOKEN_SECRET")
	polka_key := os.Getenv("POLKA_KEY")

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Error setting up the database")
	}

	queries := database.New(db)

	mux := http.NewServeMux()

	apiCfg := &apiConfig{
		Queries:     queries,
		Env:         env,
		TokenSecret: secret,
		Key:         polka_key,
	}

	limiter := tollbooth.NewLimiter(5, nil)

	// GETs
	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.getChirpById)
	mux.HandleFunc("GET /admin/metrics", apiCfg.hitsHandler)
	// POSTs
	mux.HandleFunc("POST /api/users", apiCfg.usersHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.chirpsHandler)
	mux.Handle("POST /api/login", tollbooth.LimitFuncHandler(limiter, apiCfg.loginHandler))
	mux.HandleFunc("POST /api/refresh", apiCfg.refreshTokenHandler)
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeTokenHandler)
	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.upgradeUser)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHitsHandler)
	// PUTs
	mux.HandleFunc("PUT /api/users", apiCfg.updateUserHandler)
	// DELETEs
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.deleteChirp)

	fileServer := http.FileServer(http.Dir(FILE_PATH_ROOT))
	mux.Handle("/app/", http.StripPrefix("/app", fileServer))

	s := &http.Server{
		Addr:    ":" + PORT,
		Handler: loggingMiddleware(mux),
	}

	fmt.Printf("Server running on port: 8080\n")

	log.Fatal(s.ListenAndServe())
}
