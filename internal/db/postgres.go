package db

import (
	"context"
	"errors"
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
		return nil, err
	}
	return &postgresStore{pool: pool}, nil
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
