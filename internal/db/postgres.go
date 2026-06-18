package db

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Michealshodipo56/jarcade-backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresStore struct {
	pool *pgxpool.Pool
}

func newPostgresStore(databaseURL string) (*postgresStore, error) {
	pool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, formatConnectError(err)
	}
	return &postgresStore{pool: pool}, nil
}

func formatConnectError(err error) error {
	msg := err.Error()
	if strings.Contains(msg, "password authentication failed") {
		return fmt.Errorf(`database password rejected — check DATABASE_URL on Render:
  1. Use your Supabase *database* password (Settings → Database), NOT the anon or service_role API keys
  2. If you forgot it, reset it under Settings → Database → Database password
  3. URL-encode special characters in the password (@ → %%40, # → %%23, ! → %%21, etc.)
  4. Re-copy the Transaction pooler URI (port 6543) after resetting
original error: %w`, err)
	}
	return err
}

func (s *postgresStore) Close() {
	s.pool.Close()
}

func (s *postgresStore) CreateUser(ctx context.Context, email, passwordHash string) (models.User, error) {
	const q = `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at`

	var user models.User
	err := s.pool.QueryRow(ctx, q, email, passwordHash).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return models.User{}, ErrAlreadyExists
		}
		return models.User{}, err
	}
	return user, nil
}

func (s *postgresStore) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	const q = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE lower(email) = lower($1)
		LIMIT 1`

	var user models.User
	err := s.pool.QueryRow(ctx, q, strings.TrimSpace(email)).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (s *postgresStore) GetUserByID(ctx context.Context, id string) (models.User, error) {
	const q = `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE id = $1
		LIMIT 1`

	var user models.User
	err := s.pool.QueryRow(ctx, q, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrNotFound
	}
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}
