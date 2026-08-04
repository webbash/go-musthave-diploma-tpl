package auth_test

import (
	"testing"

	"go-musthave-diploma-tpl/internal/auth"
)

func TestHashPasswordAndCheckPassword(t *testing.T) {
	hash, err := auth.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == "secret" {
		t.Fatal("HashPassword() returned plain password")
	}

	if err := auth.CheckPassword(hash, "secret"); err != nil {
		t.Fatalf("CheckPassword() error = %v", err)
	}
}

func TestCheckPasswordRejectsWrongPassword(t *testing.T) {
	hash, err := auth.HashPassword("secret")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if err := auth.CheckPassword(hash, "other"); err == nil {
		t.Fatal("CheckPassword() expected error")
	}
}
