package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/floriangaechter/rss/internal/store"
	"github.com/floriangaechter/rss/internal/utils"
)

type createUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *UserHandler) validateLoginRequest(req *loginRequest) error {
	if strings.TrimSpace(req.Username) == "" {
		return errors.New("username is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return errors.New("password is required")
	}
	return nil
}

type UserHandler struct {
	userStore    store.UserStore
	sessionStore store.SessionStore
	logger       *log.Logger
}

func NewUserHandler(userStore store.UserStore, sessionStore store.SessionStore, logger *log.Logger) *UserHandler {
	return &UserHandler{
		userStore:    userStore,
		sessionStore: sessionStore,
		logger:       logger,
	}
}

func (h *UserHandler) validateCreateRequest(req *createUserRequest) error {
	if strings.TrimSpace(req.Username) == "" {
		return errors.New("username is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		return errors.New("password is required")
	}

	return nil
}

func (h *UserHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req createUserRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		h.logger.Printf("ERROR: decoding HandleCreateUser %v", err)
		_ = utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request"})
		return
	}

	err = h.validateCreateRequest(&req)
	if err != nil {
		_ = utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}

	user := &store.User{
		Username: req.Username,
	}

	err = user.Password.Set(req.Password)
	if err != nil {
		h.logger.Printf("ERROR: hashing password %v", err)
		_ = utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	err = h.userStore.CreateUser(user)
	if err != nil {
		h.logger.Printf("ERROR: creating user %v", err)
		_ = utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	_ = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"user": user})
}

func (h *UserHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest

	// Check content type to handle both JSON and form-encoded data
	contentType := r.Header.Get("Content-Type")
	if contentType == "application/json" {
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			h.logger.Printf("ERROR: decoding HandleLogin %v", err)
			_ = utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request"})
			return
		}
	} else {
		// Handle form-encoded data
		err := r.ParseForm()
		if err != nil {
			h.logger.Printf("ERROR: parsing form %v", err)
			_ = utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": "invalid request"})
			return
		}
		req.Username = r.FormValue("username")
		req.Password = r.FormValue("password")
	}

	err := h.validateLoginRequest(&req)
	if err != nil {
		// For form submissions, redirect back with error
		if contentType != "application/json" {
			http.Redirect(w, r, "/?error="+err.Error(), http.StatusSeeOther)
			return
		}
		_ = utils.WriteJSON(w, http.StatusBadRequest, utils.Envelope{"error": err.Error()})
		return
	}

	user, err := h.userStore.GetUserByUsername(req.Username)
	if err != nil {
		h.logger.Printf("ERROR: getting user %v", err)
		if contentType != "application/json" {
			http.Redirect(w, r, "/?error=internal server error", http.StatusSeeOther)
			return
		}
		_ = utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}
	if user == nil {
		if contentType != "application/json" {
			http.Redirect(w, r, "/?error=invalid credentials", http.StatusSeeOther)
			return
		}
		_ = utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "invalid credentials"})
		return
	}

	matches, err := user.Password.Matches(req.Password)
	if err != nil {
		h.logger.Printf("ERROR: checking password %v", err)
		if contentType != "application/json" {
			http.Redirect(w, r, "/?error=internal server error", http.StatusSeeOther)
			return
		}
		_ = utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}
	if !matches {
		if contentType != "application/json" {
			http.Redirect(w, r, "/?error=invalid credentials", http.StatusSeeOther)
			return
		}
		_ = utils.WriteJSON(w, http.StatusUnauthorized, utils.Envelope{"error": "invalid credentials"})
		return
	}

	// Create session (24 hour expiration)
	session, err := h.sessionStore.CreateSession(user.ID, 24*time.Hour)
	if err != nil {
		h.logger.Printf("ERROR: creating session %v", err)
		if contentType != "application/json" {
			http.Redirect(w, r, "/?error=internal server error", http.StatusSeeOther)
			return
		}
		_ = utils.WriteJSON(w, http.StatusInternalServerError, utils.Envelope{"error": "internal server error"})
		return
	}

	// Set HTTP-only cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    session.Token,
		Path:     "/",
		MaxAge:   86400, // 24 hours in seconds
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	// Check if HTMX request
	isHTMX := r.Header.Get("HX-Request") == "true"
	if isHTMX {
		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusOK)
		return
	}

	// For form submissions, redirect to dashboard
	if contentType != "application/json" {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	// For JSON API requests, return JSON
	_ = utils.WriteJSON(w, http.StatusOK, utils.Envelope{
		"user":    user,
		"message": "login successful",
	})
}

func (h *UserHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	// Get session token from cookie
	cookie, err := r.Cookie("session_token")
	if err == nil && cookie.Value != "" {
		// Delete session from database
		err = h.sessionStore.DeleteSession(cookie.Value)
		if err != nil {
			h.logger.Printf("ERROR: deleting session %v", err)
			// Continue anyway - we'll still clear the cookie
		}
	}

	// Clear the session cookie by setting it to expire immediately
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Immediately expire
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	})

	// Check if HTMX request
	isHTMX := r.Header.Get("HX-Request") == "true"
	if isHTMX {
		// For HTMX, redirect to home page
		w.Header().Set("HX-Redirect", "/")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Regular HTTP request - redirect to login page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
