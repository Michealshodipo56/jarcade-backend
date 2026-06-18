package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/Michealshodipo56/jarcade-backend/internal/auth"
	"github.com/Michealshodipo56/jarcade-backend/internal/middleware"
)

var emailPattern = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

type AuthHandler struct {
	auth *auth.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{auth: svc}
}

type signupRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

func (h *AuthHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	email := strings.TrimSpace(req.Email)
	password := req.Password
	if email == "" || password == "" {
		writeError(w, http.StatusBadRequest, "Email and password are required.")
		return
	}
	if !emailPattern.MatchString(email) {
		writeError(w, http.StatusBadRequest, "Enter a valid email address.")
		return
	}
	if len(password) < 6 {
		writeError(w, http.StatusBadRequest, "Password must be at least 6 characters.")
		return
	}
	if req.ConfirmPassword != "" && req.ConfirmPassword != password {
		writeError(w, http.StatusBadRequest, "Passwords do not match.")
		return
	}

	user, token, err := h.auth.Signup(r.Context(), email, password)
	if errors.Is(err, auth.ErrEmailTaken) {
		writeError(w, http.StatusConflict, "Email is already registered.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not create account.")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		email = strings.TrimSpace(req.Login)
	}
	password := req.Password
	if email == "" || password == "" {
		writeError(w, http.StatusBadRequest, "Email and password are required.")
		return
	}

	user, token, err := h.auth.Login(r.Context(), email, password)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "Invalid email or password.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not sign in.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":  user,
		"token": token,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "Authorization required.")
		return
	}

	user, err := h.auth.Me(r.Context(), claims.UserID)
	if errors.Is(err, auth.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "User not found.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not load profile.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

func (h *AuthHandler) Logout(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func Health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"service": "jarcade-api",
	})
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(dst)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
