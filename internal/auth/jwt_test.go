package auth_test

import (
	"testing"
	"time"

	"go-musthave-diploma-tpl/internal/auth"
)

func TestGenerateAndParseJWT(t *testing.T) {
	token, err := auth.GenerateJWT("42", "secret", time.Minute)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	subject, err := auth.ParseJWT(token, "secret")
	if err != nil {
		t.Fatalf("ParseJWT() error = %v", err)
	}

	if subject != "42" {
		t.Fatalf("ParseJWT() subject = %q, want %q", subject, "42")
	}

	userID, err := auth.SubjectToUserID(subject)
	if err != nil {
		t.Fatalf("SubjectToUserID() error = %v", err)
	}

	if userID != 42 {
		t.Fatalf("SubjectToUserID() = %d, want %d", userID, 42)
	}
}

func TestParseJWTRejectsInvalidToken(t *testing.T) {
	if _, err := auth.ParseJWT("bad-token", "secret"); err == nil {
		t.Fatal("ParseJWT() expected error")
	}
}
