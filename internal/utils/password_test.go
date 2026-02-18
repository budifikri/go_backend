package utils

import "testing"

func TestHashAndCheckPassword(t *testing.T) {
	password := "s3cr3t-P@ssw0rd"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword error: %v", err)
	}
	if hash == "" {
		t.Fatalf("expected non-empty hash")
	}
	if !CheckPasswordHash(password, hash) {
		t.Fatalf("expected password to match hash")
	}
	if CheckPasswordHash("wrong", hash) {
		t.Fatalf("expected wrong password to not match hash")
	}
}
