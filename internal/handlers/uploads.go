package handlers

import (
	"net/http"
	"strings"

	"github.com/Michealshodipo56/jarcade-backend/internal/db"
	"github.com/Michealshodipo56/jarcade-backend/internal/middleware"
	"github.com/Michealshodipo56/jarcade-backend/internal/models"
	"github.com/go-chi/chi/v5"
)

var allowedUploadCategories = map[string]struct{}{
	"action": {}, "adventure": {}, "arcade": {}, "casual": {}, "horror": {},
	"puzzle": {}, "racing": {}, "rpg": {}, "shooting": {}, "simulation": {},
	"sports": {}, "strategy": {}, "general": {},
}

type UploadHandler struct {
	store db.Store
}

func NewUploadHandler(store db.Store) *UploadHandler {
	return &UploadHandler{store: store}
}

type createUploadRequest struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Description string `json:"description"`
	FileName    string `json:"fileName"`
	FileSize    int64  `json:"fileSize"`
	Thumbnail   string `json:"thumbnail"`
}

func (h *UploadHandler) List(w http.ResponseWriter, r *http.Request) {
	games, err := h.store.ListUploadedGames(r.Context(), 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not load uploads.")
		return
	}
	if games == nil {
		games = []models.UploadedGame{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"games": games})
}

func (h *UploadHandler) Mine(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authorization required.")
		return
	}
	games, err := h.store.ListUploadedGamesByUser(r.Context(), claims.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not load your uploads.")
		return
	}
	if games == nil {
		games = []models.UploadedGame{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"games": games})
}

func (h *UploadHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authorization required.")
		return
	}

	var req createUploadRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	name := strings.TrimSpace(req.Name)
	category := strings.TrimSpace(req.Category)
	if name == "" {
		writeError(w, http.StatusBadRequest, "Game name is required.")
		return
	}
	if len(name) > 60 {
		writeError(w, http.StatusBadRequest, "Game name is too long.")
		return
	}
	if _, ok := allowedUploadCategories[category]; !ok {
		writeError(w, http.StatusBadRequest, "Invalid category.")
		return
	}
	if len(req.Description) > 300 {
		writeError(w, http.StatusBadRequest, "Description is too long.")
		return
	}
	if len(req.Thumbnail) > 600_000 {
		writeError(w, http.StatusBadRequest, "Thumbnail is too large.")
		return
	}
	if req.FileSize < 0 || req.FileSize > 25*1024*1024 {
		writeError(w, http.StatusBadRequest, "Invalid file size.")
		return
	}

	game, err := h.store.CreateUploadedGame(r.Context(), models.UploadedGame{
		UserID:      claims.UserID,
		Name:        name,
		Category:    category,
		Description: strings.TrimSpace(req.Description),
		FileName:    strings.TrimSpace(req.FileName),
		FileSize:    req.FileSize,
		Thumbnail:   req.Thumbnail,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not save upload.")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"game": game})
}

func (h *UploadHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authorization required.")
		return
	}

	id := strings.TrimSpace(chi.URLParam(r, "id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "Upload id is required.")
		return
	}

	if err := h.store.DeleteUploadedGame(r.Context(), id, claims.UserID); err != nil {
		if err == db.ErrNotFound {
			writeError(w, http.StatusNotFound, "Upload not found.")
			return
		}
		writeError(w, http.StatusInternalServerError, "Could not delete upload.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
