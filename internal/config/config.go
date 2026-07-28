package config

import (
	"flag"
	"os"
	"strconv"
)

type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	JWTSecret            string
	JWTTTL               int
}

func Load() Config {
	var cfg Config

	flag.StringVar(&cfg.RunAddress, "a", envOrDefaultString("RUN_ADDRESS", "localhost:8080"), "run address")
	flag.StringVar(&cfg.DatabaseURI, "d", envOrDefaultString("DATABASE_URI", ""), "database uri")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", envOrDefaultString("ACCRUAL_SYSTEM_ADDRESS", ""), "accrual system address")
	flag.StringVar(&cfg.JWTSecret, "j", envOrDefaultString("JWT_SECRET", "dev-secret"), "jwt secret")
	flag.IntVar(&cfg.JWTTTL, "t", envOrDefaultInt("JWT_TTL", 60), "jwt ttl")
	flag.Parse()

	return cfg
}

func envOrDefaultString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}
