package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	JWTSecret            string
}

func Load() Config {
	var cfg Config

	flag.StringVar(&cfg.RunAddress, "a", envOrDefault("RUN_ADDRESS", "localhost:8080"), "run address")
	flag.StringVar(&cfg.DatabaseURI, "d", envOrDefault("DATABASE_URI", ""), "database uri")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", envOrDefault("ACCRUAL_SYSTEM_ADDRESS", ""), "accrual system address")
	flag.StringVar(&cfg.JWTSecret, "j", envOrDefault("JWT_SECRET", "dev-secret"), "jwt secret")
	flag.Parse()

	return cfg
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
