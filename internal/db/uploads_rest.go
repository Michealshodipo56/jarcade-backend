package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/Michealshodipo56/jarcade-backend/internal/models"
)

type restUploadRow struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	Thumbnail   string    `json:"thumbnail"`
	CreatedAt   time.Time `json:"created_at"`
}

func restUploadFromRow(r restUploadRow) models.UploadedGame {
	return models.UploadedGame{
		ID:          r.ID,
		UserID:      r.UserID,
		Name:        r.Name,
		Category:    r.Category,
		Description: r.Description,
		FileName:    r.FileName,
		FileSize:    r.FileSize,
		Thumbnail:   r.Thumbnail,
		CreatedAt:   r.CreatedAt,
	}
}

func (s *restStore) CreateUploadedGame(ctx context.Context, game models.UploadedGame) (models.UploadedGame, error) {
	body, err := json.Marshal(map[string]any{
		"user_id":     game.UserID,
		"name":        game.Name,
		"category":    game.Category,
		"description": game.Description,
		"file_name":   game.FileName,
		"file_size":   game.FileSize,
		"thumbnail":   game.Thumbnail,
	})
	if err != nil {
		return models.UploadedGame{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/uploaded_games", bytes.NewReader(body))
	if err != nil {
		return models.UploadedGame{}, err
	}
	req.Header = s.headers()
	req.Header.Set("Prefer", "return=representation")

	res, err := s.client.Do(req)
	if err != nil {
		return models.UploadedGame{}, err
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return models.UploadedGame{}, fmt.Errorf("create upload: status %d: %s", res.StatusCode, string(raw))
	}

	var rows []restUploadRow
	if err := json.Unmarshal(raw, &rows); err != nil || len(rows) == 0 {
		return models.UploadedGame{}, fmt.Errorf("unexpected upload response")
	}
	return restUploadFromRow(rows[0]), nil
}

func (s *restStore) ListUploadedGames(ctx context.Context, limit int) ([]models.UploadedGame, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	u := fmt.Sprintf("%s/uploaded_games?select=*&order=created_at.desc&limit=%d", s.baseURL, limit)
	return s.fetchUploads(ctx, u)
}

func (s *restStore) ListUploadedGamesByUser(ctx context.Context, userID string) ([]models.UploadedGame, error) {
	u := fmt.Sprintf("%s/uploaded_games?select=*&user_id=eq.%s&order=created_at.desc",
		s.baseURL, url.QueryEscape(userID))
	return s.fetchUploads(ctx, u)
}

func (s *restStore) DeleteUploadedGame(ctx context.Context, id, userID string) error {
	u := fmt.Sprintf("%s/uploaded_games?id=eq.%s&user_id=eq.%s", s.baseURL, url.QueryEscape(id), url.QueryEscape(userID))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return err
	}
	req.Header = s.headers()

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusNoContent {
		return nil
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		raw, _ := io.ReadAll(res.Body)
		return fmt.Errorf("delete upload: status %d: %s", res.StatusCode, string(raw))
	}
	return nil
}

func (s *restStore) fetchUploads(ctx context.Context, url string) ([]models.UploadedGame, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header = s.headers()

	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("uploaded_games table not found — run supabase/migrations/003_uploaded_games.sql")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("list uploads: status %d: %s", res.StatusCode, string(raw))
	}

	var rows []restUploadRow
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, err
	}
	out := make([]models.UploadedGame, 0, len(rows))
	for _, row := range rows {
		out = append(out, restUploadFromRow(row))
	}
	return out, nil
}
