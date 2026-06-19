package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/Michealshodipo56/jarcade-backend/internal/db"
)

var (
	ErrInvalidResetToken = errors.New("invalid or expired reset token")
)

func (s *Service) RequestPasswordReset(ctx context.Context, email string) (string, error) {
	email = normalizeEmail(email)
	user, err := s.store.GetUserByEmail(ctx, email)
	if errors.Is(err, db.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	rawToken, err := generateResetToken()
	if err != nil {
		return "", err
	}

	hash := hashResetToken(rawToken)
	expires := time.Now().Add(1 * time.Hour)
	if err := s.store.CreatePasswordResetToken(ctx, user.ID, hash, expires); err != nil {
		return "", err
	}
	return rawToken, nil
}

func (s *Service) ResetPassword(ctx context.Context, rawToken, newPassword string) error {
	if len(newPassword) < 6 {
		return errors.New("password too short")
	}

	userID, err := s.store.ConsumePasswordResetToken(ctx, hashResetToken(rawToken))
	if errors.Is(err, db.ErrNotFound) {
		return ErrInvalidResetToken
	}
	if err != nil {
		return err
	}

	hash, err := HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.store.UpdateUserPassword(ctx, userID, hash)
}

func generateResetToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
