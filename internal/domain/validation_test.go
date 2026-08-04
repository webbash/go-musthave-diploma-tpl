package domain_test

import (
	"testing"

	"go-musthave-diploma-tpl/internal/domain"
)

func TestIsValidLuhn(t *testing.T) {
	if !domain.IsValidLuhn("79927398713") {
		t.Fatal("IsValidLuhn() = false, want true")
	}

	if domain.IsValidLuhn("79927398710") {
		t.Fatal("IsValidLuhn() = true, want false")
	}

	if domain.IsValidLuhn("12ab") {
		t.Fatal("IsValidLuhn() = true, want false")
	}
}
