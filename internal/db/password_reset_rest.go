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
)

func (s *restStore) UpdateUserPassword(ctx context.Context, userID, passwordHash string) error {
	body, _ := json.Marshal(map[string]string{"password_hash": passwordHash})
	q := url.Values{}
	q.Set("id", "eq."+userID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, s.baseURL+"/users?"+q.Encode(), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header = s.headers()

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		raw, _ := io.ReadAll(res.Body)
		return fmt.Errorf("update password failed (%d): %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	return nil
}

func (s *restStore) CreatePasswordResetToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	body, _ := json.Marshal(map[string]any{
		"user_id":     userID,
		"token_hash":  tokenHash,
		"expires_at":  expiresAt.UTC().Format(time.RFC3339),
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/password_reset_tokens", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header = s.headers()

	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		raw, _ := io.ReadAll(res.Body)
		return fmt.Errorf("create reset token failed (%d): %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	return nil
}

func (s *restStore) ConsumePasswordResetToken(ctx context.Context, tokenHash string) (string, error) {
	q := url.Values{}
	q.Set("select", "id,user_id")
	q.Set("token_hash", "eq."+tokenHash)
	q.Set("used_at", "is.null")
	q.Set("expires_at", "gt."+time.Now().UTC().Format(time.RFC3339))
	q.Set("limit", "1")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.baseURL+"/password_reset_tokens?"+q.Encode(), nil)
	if err != nil {
		return "", err
	}
	req.Header = s.headers()

	res, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return "", fmt.Errorf("lookup reset token failed (%d)", res.StatusCode)
	}

	var rows []struct {
		ID     string `json:"id"`
		UserID string `json:"user_id"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil || len(rows) == 0 {
		return "", ErrNotFound
	}

	patchBody, _ := json.Marshal(map[string]string{
		"used_at": time.Now().UTC().Format(time.RFC3339),
	})
	patchQ := url.Values{}
	patchQ.Set("id", "eq."+rows[0].ID)

	patchReq, err := http.NewRequestWithContext(ctx, http.MethodPatch, s.baseURL+"/password_reset_tokens?"+patchQ.Encode(), bytes.NewReader(patchBody))
	if err != nil {
		return "", err
	}
	patchReq.Header = s.headers()

	patchRes, err := s.client.Do(patchReq)
	if err != nil {
		return "", err
	}
	defer patchRes.Body.Close()
	if patchRes.StatusCode < 200 || patchRes.StatusCode >= 300 {
		return "", fmt.Errorf("consume reset token failed (%d)", patchRes.StatusCode)
	}

	return rows[0].UserID, nil
}
