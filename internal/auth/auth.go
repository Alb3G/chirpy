package auth

import "github.com/alexedwards/argon2id"

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
