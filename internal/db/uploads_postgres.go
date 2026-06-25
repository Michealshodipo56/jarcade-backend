package db

import (
	"context"

	"github.com/Michealshodipo56/jarcade-backend/internal/models"
	"github.com/jackc/pgx/v5"
)

func (s *postgresStore) CreateUploadedGame(ctx context.Context, game models.UploadedGame) (models.UploadedGame, error) {
	const q = `
		INSERT INTO uploaded_games (user_id, name, category, description, file_name, file_size, thumbnail)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, user_id, name, category, description, file_name, file_size, thumbnail, created_at`

	var out models.UploadedGame
	err := s.pool.QueryRow(ctx, q,
		game.UserID, game.Name, game.Category, game.Description,
		game.FileName, game.FileSize, game.Thumbnail,
	).Scan(
		&out.ID, &out.UserID, &out.Name, &out.Category, &out.Description,
		&out.FileName, &out.FileSize, &out.Thumbnail, &out.CreatedAt,
	)
	return out, err
}

func (s *postgresStore) ListUploadedGames(ctx context.Context, limit int) ([]models.UploadedGame, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	const q = `
		SELECT id, user_id, name, category, description, file_name, file_size, thumbnail, created_at
		FROM uploaded_games
		ORDER BY created_at DESC
		LIMIT $1`

	rows, err := s.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUploadedGames(rows)
}

func (s *postgresStore) ListUploadedGamesByUser(ctx context.Context, userID string) ([]models.UploadedGame, error) {
	const q = `
		SELECT id, user_id, name, category, description, file_name, file_size, thumbnail, created_at
		FROM uploaded_games
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := s.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanUploadedGames(rows)
}

func (s *postgresStore) DeleteUploadedGame(ctx context.Context, id, userID string) error {
	const q = `DELETE FROM uploaded_games WHERE id = $1 AND user_id = $2`
	tag, err := s.pool.Exec(ctx, q, id, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func scanUploadedGames(rows pgx.Rows) ([]models.UploadedGame, error) {
	var list []models.UploadedGame
	for rows.Next() {
		var g models.UploadedGame
		if err := rows.Scan(
			&g.ID, &g.UserID, &g.Name, &g.Category, &g.Description,
			&g.FileName, &g.FileSize, &g.Thumbnail, &g.CreatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, rows.Err()
}
