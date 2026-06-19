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
	auth            *auth.Service
	frontendURL     string
	exposeResetLink bool
	googleClientID  string
}

func NewAuthHandler(svc *auth.Service, frontendURL string, exposeResetLink bool, googleClientID string) *AuthHandler {
	return &AuthHandler{
		auth:            svc,
		frontendURL:     frontendURL,
		exposeResetLink: exposeResetLink,
		googleClientID:  googleClientID,
	}
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

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}
	email := strings.TrimSpace(req.Email)
	if email == "" || !emailPattern.MatchString(email) {
		writeError(w, http.StatusBadRequest, "Enter a valid email address.")
		return
	}

	token, err := h.auth.RequestPasswordReset(r.Context(), email)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not process reset request.")
		return
	}

	resp := map[string]string{
		"message": "If an account exists for that email, password reset instructions have been sent.",
	}
	if token != "" && h.exposeResetLink && h.frontendURL != "" {
		resp["resetUrl"] = h.frontendURL + "/reset-password.html?token=" + token
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token           string `json:"token"`
		Password        string `json:"password"`
		ConfirmPassword string `json:"confirmPassword"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}
	if req.Token == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "Token and new password are required.")
		return
	}
	if len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "Password must be at least 6 characters.")
		return
	}
	if req.ConfirmPassword != "" && req.ConfirmPassword != req.Password {
		writeError(w, http.StatusBadRequest, "Passwords do not match.")
		return
	}

	err := h.auth.ResetPassword(r.Context(), strings.TrimSpace(req.Token), req.Password)
	if errors.Is(err, auth.ErrInvalidResetToken) {
		writeError(w, http.StatusBadRequest, "This reset link is invalid or has expired.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not reset password.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Password updated successfully. You can sign in now.",
	})
}

func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	if h.googleClientID == "" {
		writeError(w, http.StatusServiceUnavailable, "Google sign-in is not configured on the server.")
		return
	}

	var req struct {
		IDToken string `json:"idToken"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body.")
		return
	}
	if strings.TrimSpace(req.IDToken) == "" {
		writeError(w, http.StatusBadRequest, "Google token is required.")
		return
	}

	user, token, err := h.auth.LoginWithGoogle(r.Context(), req.IDToken, h.googleClientID)
	if errors.Is(err, auth.ErrGoogleAuthDisabled) {
		writeError(w, http.StatusServiceUnavailable, "Google sign-in is not configured on the server.")
		return
	}
	if errors.Is(err, auth.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "Google sign-in failed. Try again.")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Could not sign in with Google.")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"user": user, "token": token})
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
