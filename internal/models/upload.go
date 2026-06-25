package models

import "time"

type UploadedGame struct {
	ID          string    `json:"id"`
	UserID      string    `json:"userId"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	FileName    string    `json:"fileName"`
	FileSize    int64     `json:"fileSize"`
	Thumbnail   string    `json:"thumbnail"`
	CreatedAt   time.Time `json:"createdAt"`
}
