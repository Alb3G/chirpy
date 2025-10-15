package main

import (
	"sync/atomic"
	"time"

	"github.com/Alb3G/chirpy/internal/database"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
}

type UserRequestData struct {
	ExpiresIn int    `json:"expires_in_seconds"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserId    uuid.UUID `json:"user_id"`
}

type apiConfig struct {
	fileserverhits atomic.Int32
	Queries        *database.Queries
	Env            string
	TokenScret     string
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Body string `json:"cleaned_body"`
}

type ChirpBody struct {
	Body   string    `json:"body"`
	UserId uuid.UUID `json:"user_id"`
}
