package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"

	"github.com/Michealshodipo56/jarcade-backend/internal/db"
	"github.com/Michealshodipo56/jarcade-backend/internal/models"
	"google.golang.org/api/idtoken"
)

var ErrGoogleAuthDisabled = errors.New("google sign-in is not configured")

func (s *Service) LoginWithGoogle(ctx context.Context, idToken, audience string) (models.PublicUser, string, error) {
	if audience == "" {
		return models.PublicUser{}, "", ErrGoogleAuthDisabled
	}

	payload, err := idtoken.Validate(ctx, idToken, audience)
	if err != nil {
		return models.PublicUser{}, "", ErrInvalidCredentials
	}

	email, _ := payload.Claims["email"].(string)
	if email == "" {
		return models.PublicUser{}, "", ErrInvalidCredentials
	}
	email = normalizeEmail(email)

	user, err := s.store.GetUserByEmail(ctx, email)
	if errors.Is(err, db.ErrNotFound) {
		randomPass, genErr := randomPassword()
		if genErr != nil {
			return models.PublicUser{}, "", genErr
		}
		hash, hashErr := HashPassword(randomPass)
		if hashErr != nil {
			return models.PublicUser{}, "", hashErr
		}
		user, err = s.store.CreateUser(ctx, email, hash)
		if err != nil {
			return models.PublicUser{}, "", err
		}
	} else if err != nil {
		return models.PublicUser{}, "", err
	}

	token, err := s.tokens.Sign(user.ID, user.Email)
	if err != nil {
		return models.PublicUser{}, "", err
	}
	return models.ToPublicUser(user), token, nil
}

func randomPassword() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
