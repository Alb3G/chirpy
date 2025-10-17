package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Alb3G/chirpy/internal/database"
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

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatal("Error setting up the database")
	}

	queries := database.New(db)

	mux := http.NewServeMux()

	apiCfg := &apiConfig{
		Queries:    queries,
		Env:        env,
		TokenScret: secret,
	}

	// GETs
	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.getChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.getChirpById)
	mux.HandleFunc("GET /admin/metrics", apiCfg.hitsHandler)
	// POSTs
	mux.HandleFunc("POST /api/users", apiCfg.usersHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.chirpsHandler)
	mux.HandleFunc("POST /api/login", apiCfg.loginHandler)
	mux.HandleFunc("POST /api/refresh", apiCfg.refreshTokenHandler)
	mux.HandleFunc("POST /api/revoke", apiCfg.revokeTokenHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetHitsHandler)

	fileServer := http.FileServer(http.Dir(FILE_PATH_ROOT))
	mux.Handle("/app/", http.StripPrefix("/app", apiCfg.metricsCountMiddleware(fileServer)))

	s := &http.Server{
		Addr:    ":" + PORT,
		Handler: mux,
	}

	fmt.Printf("Server running on port: 8080\n")

	log.Fatal(s.ListenAndServe())
}
