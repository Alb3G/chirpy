package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	hashed_pass, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", err
	}

	return hashed_pass, nil
}

func CheckPasswordHash(password, hash string) (bool, error) {
	match, _, err := argon2id.CheckHash(password, hash)
	return match, err
}

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  &jwt.NumericDate{Time: time.Now().UTC()},
		ExpiresAt: &jwt.NumericDate{Time: time.Now().Add(expiresIn)},
		Subject:   userId.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	userId, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	return uuid.MustParse(userId), nil
}

func GetBearerToken(headers http.Header) (string, error) {
	bearer := headers.Get("Authorization")
	if bearer == "" {
		return "", errors.New("no authorization header provided")
	}

	bearerParts := strings.Split(bearer, " ")
	if len(bearerParts) != 2 {
		return "", errors.New("wrong bearer format provided")
	}

	return bearerParts[1], nil
}

func MakeRefreshToken() string {
	randTokenBytes := make([]byte, 32)

	rand.Read(randTokenBytes)

	token := hex.EncodeToString(randTokenBytes)

	return token
}
