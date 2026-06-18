package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/Michealshodipo56/jarcade-backend/internal/db"
	"github.com/Michealshodipo56/jarcade-backend/internal/models"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email is already registered")
)

type Service struct {
	store  db.Store
	tokens *TokenService
}

func NewService(store db.Store, tokens *TokenService) *Service {
	return &Service{store: store, tokens: tokens}
}

func (s *Service) Signup(ctx context.Context, email, password string) (models.PublicUser, string, error) {
	email = normalizeEmail(email)
	hash, err := HashPassword(password)
	if err != nil {
		return models.PublicUser{}, "", err
	}

	user, err := s.store.CreateUser(ctx, email, hash)
	if errors.Is(err, db.ErrAlreadyExists) {
		return models.PublicUser{}, "", ErrEmailTaken
	}
	if err != nil {
		return models.PublicUser{}, "", err
	}

	token, err := s.tokens.Sign(user.ID, user.Email)
	if err != nil {
		return models.PublicUser{}, "", err
	}
	return models.ToPublicUser(user), token, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (models.PublicUser, string, error) {
	email = normalizeEmail(email)
	user, err := s.store.GetUserByEmail(ctx, email)
	if errors.Is(err, db.ErrNotFound) {
		return models.PublicUser{}, "", ErrInvalidCredentials
	}
	if err != nil {
		return models.PublicUser{}, "", err
	}

	if !CheckPassword(user.PasswordHash, password) {
		return models.PublicUser{}, "", ErrInvalidCredentials
	}

	token, err := s.tokens.Sign(user.ID, user.Email)
	if err != nil {
		return models.PublicUser{}, "", err
	}
	return models.ToPublicUser(user), token, nil
}

func (s *Service) Me(ctx context.Context, userID string) (models.PublicUser, error) {
	user, err := s.store.GetUserByID(ctx, userID)
	if errors.Is(err, db.ErrNotFound) {
		return models.PublicUser{}, ErrInvalidCredentials
	}
	if err != nil {
		return models.PublicUser{}, err
	}
	return models.ToPublicUser(user), nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
