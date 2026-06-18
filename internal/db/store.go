package db

import (
	"context"
	"errors"

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
	if cfg.DatabaseURL != "" {
		return newPostgresStore(cfg.DatabaseURL)
	}
	return newRESTStore(cfg.SupabaseURL, cfg.SupabaseServiceRoleKey)
}
