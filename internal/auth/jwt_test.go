package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// TestMakeJWT creates a new token
func TestMakeJWT(t *testing.T) {
	token, err := MakeJWT(uuid.New(), "AllYourBase", time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT() returned an error: %v", err)
	}
	if token == "" {
		t.Error("MakeJWT() returned an empty token")
	}
}

// TestValidateJWT validates a token
func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "AllYourBase"
	token, err := MakeJWT(userID, tokenSecret, 20+time.Second)
	if err != nil {
		t.Fatalf("MakeJWT() failed: %v", err)
	}

	validatedID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Fatalf("ValidateJWT() returned an error: %v", err)
	}
	if userID != validatedID {
		t.Error("ValidateJWT() failed to validate userID")
	}
}

// TestRejectJWT rejects expired tokens
func TestRejectJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "AllYourBase"
	token, err := MakeJWT(userID, tokenSecret, 0)
	if err != nil {
		t.Fatalf("MakeJWT() failed: %v", err)
	}

	_, err = ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Fatal("expected validation to fail for expired token")
	}
	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got %v", err)
	}
}

// TestWrongSecret rejects tokens with the incorect secret
func TestWrongSecret(t *testing.T) {
	userID := uuid.New()
	token, err := MakeJWT(userID, "WrongSecret", 0)
	if err != nil {
		t.Fatalf("MakeJWT() failed: %v", err)
	}

	_, err = ValidateJWT(token, "AllYourBase")
	if err == nil {
		t.Fatal("expected validation to fail for incorrect secret")
	}
	if !errors.Is(err, jwt.ErrSignatureInvalid) {
		t.Fatalf("expected ErrSignatureInvalid, got %v", err)
	}
}
