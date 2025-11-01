package auth

import (
	"testing"
)

// TestHashPassword calls auth.HashPassword with a password, returning a hashed
// password.
func TestHashPassword(t *testing.T) {
	password := "my_password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() returned an error: %v", err)
	}
	if hash == "" {
		t.Error("HashPassword() returned an empty hash")
	}
	if hash == password {
		t.Error("HashPassword() returned the password unhashed")
	}
}

// TestCheckPasswordHash verifies that a correct password matches the hash
// and an incorrect password does not match.
func TestCheckPasswordHash(t *testing.T) {
	password := "my_password"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed: %v", err)
	}

	// Test with correct password
	match, err := CheckPasswordHash(password, hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash() returned an error: %v", err)
	}
	if !match {
		t.Error("CheckPasswordHash() failed to match correct password")
	}

	// Test with incorrect password
	wrongPassword := "wrong_password"
	match, err = CheckPasswordHash(wrongPassword, hash)
	if err != nil {
		t.Fatalf("CheckPasswordHash() returned an error with wrong password: %v", err)
	}
	if match {
		t.Error("CheckPasswordHash() matched incorrect password")
	}
}

// TestHashPasswordUnique ensures that hashing the same password twice
// produces different hashes (due to unique salts).
func TestHashPasswordUnique(t *testing.T) {
	password := "my_password"
	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() first call failed: %v", err)
	}
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() second call failed: %v", err)
	}
	if hash1 == hash2 {
		t.Error("HashPassword() produced identical hashes for same password (salt not random)")
	}
}

// TestHashPasswordEmpty tests behavior with an empty password.
func TestHashPasswordEmpty(t *testing.T) {
	password := ""
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() failed with empty password: %v", err)
	}
	if hash == "" {
		t.Error("HashPassword() returned empty hash for empty password")
	}
}

// TestCheckPasswordHashInvalidHash tests behavior with an invalid hash format.
func TestCheckPasswordHashInvalidHash(t *testing.T) {
	password := "my_password"
	invalidHash := "not_a_valid_hash"

	match, err := CheckPasswordHash(password, invalidHash)
	if err == nil {
		t.Error("CheckPasswordHash() should return an error for invalid hash")
	}
	if match {
		t.Error("CheckPasswordHash() should not match with invalid hash")
	}
}
