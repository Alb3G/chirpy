package testing

import (
	"testing"
	"time"

	auth "github.com/Alb3G/chirpy/internal/auth"
	"github.com/google/uuid"
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

func TestMakeAndValidateJWT(t *testing.T) {
	// Creation of JWT
	userId := uuid.MustParse("fb68025f-be8f-4649-aa15-0c2b6b1c6409")
	secret := "test-secret-key"

	token, err := auth.MakeJWT(userId, secret, time.Minute*2)
	if err != nil {
		t.Fatalf("Failed signing JWT: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token value")
	}

	// Validation of JWT
	validatedUserId, err := auth.ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("Internal error validation of the JWT, err: %v", err)
	}

	if validatedUserId.String() != userId.String() {
		t.Error("Error expecting token values to be equals")
	}
}
