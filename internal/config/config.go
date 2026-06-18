package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port                   string
	JWTSecret              string
	SupabaseURL            string
	SupabaseServiceRoleKey string
	DatabaseURL            string
	CORSOrigins            []string
}

func Load() (Config, error) {
	cfg := Config{
		Port:                   envOrDefault("PORT", "8080"),
		JWTSecret:              strings.TrimSpace(os.Getenv("JWT_SECRET")),
		SupabaseURL:            strings.TrimRight(strings.TrimSpace(os.Getenv("SUPABASE_URL")), "/"),
		SupabaseServiceRoleKey: strings.TrimSpace(os.Getenv("SUPABASE_SERVICE_ROLE_KEY")),
		DatabaseURL:            strings.TrimSpace(os.Getenv("DATABASE_URL")),
	}

	if cfg.JWTSecret == "" {
		return cfg, fmt.Errorf("JWT_SECRET is required")
	}
	if len(cfg.JWTSecret) < 32 {
		return cfg, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}

	hasREST := cfg.SupabaseURL != "" && cfg.SupabaseServiceRoleKey != ""
	if cfg.DatabaseURL == "" && !hasREST {
		return cfg, fmt.Errorf("set DATABASE_URL or both SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY")
	}

	cors := envOrDefault("CORS_ORIGIN", "http://localhost:5500,http://127.0.0.1:5500")
	for _, origin := range strings.Split(cors, ",") {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			cfg.CORSOrigins = append(cfg.CORSOrigins, origin)
		}
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
