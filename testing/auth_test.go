package testing

import (
	"testing"

	auth "github.com/Alb3G/chirpy/internal/auth"
)

func TestPasswordHashing(t *testing.T) {
	password := "password1234."
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Test case failed with error: %v", err)
	}

	if hash == "" {
		t.Fatalf("Expected non-empty hash")
	}

	if hash == password {
		t.Fatalf("Hash should be different to original password")
	}
}

func TestCheckHash(t *testing.T) {
	password := "password1234."

	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Test case failed with error: %v", err)
	}

	match, err := auth.CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("Test case failed with error: %v", err)
	}

	if !match {
		t.Fatalf("Hash and password should match")
	}
}
