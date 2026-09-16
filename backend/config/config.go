package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv          string
	AppPort         string
	DatabaseURL     string
	JWTSecret       string
	JWTExpiration   time.Duration
	CORSOrigins     []string
	APIPrefix       string
}

func Load() (*Config, error) {
	cfg := &Config{
		AppEnv:        getEnv("APP_ENV", "development"),
		AppPort:       getEnv("APP_PORT", "8080"),
		DatabaseURL:   getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/service_app?sslmode=disable"),
		JWTSecret:     getEnv("JWT_SECRET", "dev-only-change-me"),
		JWTExpiration: mustDuration(getEnv("JWT_EXPIRATION", "24h")),
		CORSOrigins:   splitCSV(getEnv("CORS_ORIGINS", "http://localhost:5173")),
		APIPrefix:     "/api/v1",
	}

	if len(cfg.CORSOrigins) == 0 {
		cfg.CORSOrigins = []string{"http://localhost:5173"}
	}

	if cfg.AppEnv != "test" && cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET must be set")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func mustDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 24 * time.Hour
	}
	return d
}

func MustPort(p string) int {
	n, err := strconv.Atoi(p)
	if err != nil {
		return 8080
	}
	return n
}