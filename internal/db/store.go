package db

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Michealshodipo56/jarcade-backend/internal/config"
	"github.com/Michealshodipo56/jarcade-backend/internal/models"
)

var (
	ErrNotFound      = errors.New("user not found")
	ErrAlreadyExists = errors.New("user already exists")
)

type Store interface {
	CreateUser(ctx context.Context, email, passwordHash string) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
	GetUserByID(ctx context.Context, id string) (models.User, error)
	Close()
}

func New(cfg config.Config) (Store, error) {
	hasREST := cfg.SupabaseURL != "" && cfg.SupabaseServiceRoleKey != ""

	if cfg.DatabaseURL != "" {
		store, err := newPostgresStore(cfg.DatabaseURL)
		if err == nil {
			log.Println("database: connected via Postgres (DATABASE_URL)")
			return store, nil
		}
		if hasREST && isRecoverableDBError(err) {
			log.Printf("database: DATABASE_URL failed (%v), falling back to Supabase REST API", err)
			return newRESTStore(cfg.SupabaseURL, cfg.SupabaseServiceRoleKey)
		}
		return nil, err
	}

	if hasREST {
		log.Println("database: using Supabase REST API")
		return newRESTStore(cfg.SupabaseURL, cfg.SupabaseServiceRoleKey)
	}

	return nil, fmt.Errorf("set DATABASE_URL or both SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY")
}

func isRecoverableDBError(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "password authentication failed") ||
		strings.Contains(msg, "database password rejected")
}
