package db

import (
	"context"
	"time"
)

func (s *postgresStore) UpdateUserPassword(ctx context.Context, userID, passwordHash string) error {
	const q = `UPDATE users SET password_hash = $1 WHERE id = $2`
	_, err := s.pool.Exec(ctx, q, passwordHash, userID)
	return err
}

func (s *postgresStore) CreatePasswordResetToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	const q = `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`
	_, err := s.pool.Exec(ctx, q, userID, tokenHash, expiresAt)
	return err
}

func (s *postgresStore) ConsumePasswordResetToken(ctx context.Context, tokenHash string) (string, error) {
	const q = `
		UPDATE password_reset_tokens
		SET used_at = now()
		WHERE token_hash = $1
		  AND used_at IS NULL
		  AND expires_at > now()
		RETURNING user_id`

	var userID string
	err := s.pool.QueryRow(ctx, q, tokenHash).Scan(&userID)
	if err != nil {
		return "", ErrNotFound
	}
	return userID, nil
}
