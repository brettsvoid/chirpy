package auth

import (
	"errors"
	"net/http"
	"testing"
)

// TestGetBearerToken extracts the token out of the headers
func TestGetBearerToken(t *testing.T) {
	headers := http.Header{}
	headers.Add("Authorization", "Bearer abc123")
	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("GetBearerToken() returned an error: %v", err)
	}
	if token == "" {
		t.Error("GetBearerToken() returned an empty token")
	}
}

// TestInvalidAuhtorizationHeader fails for invalid Authorization headers
func TestInvalidAuhtorizationHeader(t *testing.T) {
	headers := http.Header{}
	_, err := GetBearerToken(headers)
	if !errors.Is(err, ErrNoAuthHeaderIncluded) {
		t.Fatalf("expected ErrNoAuthHeaderIncluded, got %v", err)
	}

	headers.Set("Authorization", "Bearer")
	_, err = GetBearerToken(headers)
	if err == nil {
		t.Fatal("expected validation to fail for missing token")
	}

	headers.Set("Authorization", "token")
	_, err = GetBearerToken(headers)
	if err == nil {
		t.Fatal("expected validation to fail for malformed token")
	}
}
