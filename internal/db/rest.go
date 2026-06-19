package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Michealshodipo56/jarcade-backend/internal/models"
)

type restStore struct {
	baseURL string
	key     string
	client  *http.Client
}

type restUserRow struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

func newRESTStore(supabaseURL, serviceRoleKey string) (*restStore, error) {
	if supabaseURL == "" || serviceRoleKey == "" {
		return nil, fmt.Errorf("SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY are required for REST mode")
	}
	s := &restStore{
		baseURL: strings.TrimRight(supabaseURL, "/") + "/rest/v1",
		key:     serviceRoleKey,
		client:  &http.Client{Timeout: 15 * time.Second},
	}
	if err := s.ping(); err != nil {
		return nil, fmt.Errorf("supabase REST API unreachable: %w", err)
	}
	return s, nil
}

func (s *restStore) ping() error {
	req, err := http.NewRequest(http.MethodGet, s.baseURL+"/users?select=id&limit=0", nil)
	if err != nil {
		return err
	}
	req.Header = s.headers()

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound {
		return fmt.Errorf("users table not found — run supabase/migrations/001_users.sql in the SQL editor")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func (s *restStore) Close() {}

func (s *restStore) headers() http.Header {
	h := make(http.Header)
	h.Set("apikey", s.key)
	h.Set("Authorization", "Bearer "+s.key)
	h.Set("Content-Type", "application/json")
	return h
}

func (s *restStore) CreateUser(ctx context.Context, email, passwordHash string) (models.User, error) {
	body, err := json.Marshal(map[string]string{
		"email":         email,
		"password_hash": passwordHash,
	})
	if err != nil {
		return models.User{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/users", bytes.NewReader(body))
	if err != nil {
		return models.User{}, err
	}
	req.Header = s.headers()
	req.Header.Set("Prefer", "return=representation")

	res, err := s.client.Do(req)
	if err != nil {
		return models.User{}, err
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode == http.StatusConflict {
		return models.User{}, ErrAlreadyExists
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return models.User{}, fmt.Errorf("supabase insert failed (%d): %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}

	var rows []restUserRow
	if err := json.Unmarshal(raw, &rows); err != nil || len(rows) == 0 {
		return models.User{}, fmt.Errorf("unexpected supabase response")
	}
	return rowToUser(rows[0]), nil
}

func (s *restStore) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	q := url.Values{}
	q.Set("select", "id,email,password_hash,created_at")
	q.Set("email", "eq."+strings.TrimSpace(email))
	q.Set("limit", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/users?"+q.Encode(), nil)
	if err != nil {
		return models.User{}, err
	}
	req.Header = s.headers()

	res, err := s.client.Do(req)
	if err != nil {
		return models.User{}, err
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return models.User{}, fmt.Errorf("supabase query failed (%d)", res.StatusCode)
	}

	var rows []restUserRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return models.User{}, err
	}
	if len(rows) == 0 {
		return models.User{}, ErrNotFound
	}
	return rowToUser(rows[0]), nil
}

func (s *restStore) GetUserByID(ctx context.Context, id string) (models.User, error) {
	q := url.Values{}
	q.Set("select", "id,email,password_hash,created_at")
	q.Set("id", "eq."+id)
	q.Set("limit", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/users?"+q.Encode(), nil)
	if err != nil {
		return models.User{}, err
	}
	req.Header = s.headers()

	res, err := s.client.Do(req)
	if err != nil {
		return models.User{}, err
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return models.User{}, fmt.Errorf("supabase query failed (%d)", res.StatusCode)
	}

	var rows []restUserRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return models.User{}, err
	}
	if len(rows) == 0 {
		return models.User{}, ErrNotFound
	}
	return rowToUser(rows[0]), nil
}

func rowToUser(row restUserRow) models.User {
	return models.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		CreatedAt:    row.CreatedAt,
	}
}
