package utils

import (
	"testing"
	"time"
)

func TestJWTGenerateAndVerify(t *testing.T) {
	util := NewJWTUtil("test-secret", 10*time.Minute)
	payload := JWTPayload{
		UserID:    "user-1",
		Username:  "alice",
		Email:     "alice@example.com",
		CompanyID: "company-1",
		Role:      "admin",
	}

	token, err := util.GenerateToken(payload)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	if token == "" {
		t.Fatalf("expected token")
	}

	out, err := util.VerifyToken(token)
	if err != nil {
		t.Fatalf("VerifyToken error: %v", err)
	}
	if out.UserID != payload.UserID {
		t.Fatalf("userId mismatch: got %q want %q", out.UserID, payload.UserID)
	}
	if out.CompanyID != payload.CompanyID {
		t.Fatalf("companyId mismatch: got %q want %q", out.CompanyID, payload.CompanyID)
	}
	if out.Role != payload.Role {
		t.Fatalf("role mismatch: got %q want %q", out.Role, payload.Role)
	}
}

func TestJWTInvalidSignature(t *testing.T) {
	util1 := NewJWTUtil("secret-1", 10*time.Minute)
	util2 := NewJWTUtil("secret-2", 10*time.Minute)

	token, err := util1.GenerateToken(JWTPayload{UserID: "u"})
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	_, err = util2.VerifyToken(token)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestJWTExpired(t *testing.T) {
	util := NewJWTUtil("test-secret", -1*time.Second)
	token, err := util.GenerateToken(JWTPayload{UserID: "u"})
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	_, err = util.VerifyToken(token)
	if err == nil {
		t.Fatalf("expected expired token error")
	}
}
