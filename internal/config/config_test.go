package config

import (
	"flag"
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	resetFlagSet(t)

	cfg := Load()
	if cfg.RunAddress != "localhost:8080" {
		t.Fatalf("RunAddress = %q", cfg.RunAddress)
	}
	if cfg.JWTSecret != "dev-secret" {
		t.Fatalf("JWTSecret = %q", cfg.JWTSecret)
	}
	if cfg.JWTTTL != 60 {
		t.Fatalf("JWTTTL = %d", cfg.JWTTTL)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("RUN_ADDRESS", "127.0.0.1:9000")
	t.Setenv("DATABASE_URI", "postgres://example")
	t.Setenv("ACCRUAL_SYSTEM_ADDRESS", "http://accrual.local")
	t.Setenv("JWT_SECRET", "secret")
	t.Setenv("JWT_TTL", "120")

	resetFlagSet(t)

	cfg := Load()
	if cfg.RunAddress != "127.0.0.1:9000" {
		t.Fatalf("RunAddress = %q", cfg.RunAddress)
	}
	if cfg.DatabaseURI != "postgres://example" {
		t.Fatalf("DatabaseURI = %q", cfg.DatabaseURI)
	}
	if cfg.AccrualSystemAddress != "http://accrual.local" {
		t.Fatalf("AccrualSystemAddress = %q", cfg.AccrualSystemAddress)
	}
	if cfg.JWTSecret != "secret" {
		t.Fatalf("JWTSecret = %q", cfg.JWTSecret)
	}
	if cfg.JWTTTL != 120 {
		t.Fatalf("JWTTTL = %d", cfg.JWTTTL)
	}
}

func TestLoadFallsBackOnInvalidTTL(t *testing.T) {
	t.Setenv("JWT_TTL", "invalid")
	resetFlagSet(t)

	cfg := Load()
	if cfg.JWTTTL != 60 {
		t.Fatalf("JWTTTL = %d", cfg.JWTTTL)
	}
}

func resetFlagSet(t *testing.T) {
	t.Helper()
	oldArgs := os.Args
	oldFlagSet := flag.CommandLine

	os.Args = []string{oldArgs[0]}
	flag.CommandLine = flag.NewFlagSet(oldArgs[0], flag.ContinueOnError)
	t.Cleanup(func() {
		os.Args = oldArgs
		flag.CommandLine = oldFlagSet
	})
}
